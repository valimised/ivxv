# IVXV Internet voting framework
"""Microservice management helper."""

import json
import os
import re
import shutil
import subprocess
import tempfile

from jinja2 import Environment, PackageLoader

from debian import debfile

from .. import (
    COLLECTOR_PKG_FILENAMES,
    DEB_PKG_VERSION,
    SERVICE_SECRET_TYPES,
    SERVICE_STATE_CONFIGURED,
    SERVICE_STATE_INSTALLED,
    SERVICE_STATE_NOT_INSTALLED,
    SERVICE_TYPE_PARAMS,
)
from ..config import CONFIG, cfg_path
from ..db import IVXVManagerDb
from ..event_log import register_service_event
from . import IVXV_ADMIN_SSH_PUBKEY_FILE, RSYSLOG_CFG_FILENAME, generate_service_hints
from .backup_service import install_backup_crontab
from .logging import ServiceLogger
from .remote_exec import exec_remote_cmd

#: Path to service directory.
SERVICE_DIR = cfg_path('ivxv_admin_data_path', 'service')


class Service:
    """Service management for collector services."""
    service_id = None  #: Service ID (str)
    data = None  #: Service data (dict)
    log = None  #: Logger for service
    #: State report file name for currently applied config (str)
    cfg_state_filepath = None
    cfg_state = None  #: State report for currently applied config (dict)

    def __init__(self, service_id, service_data):
        """Constructor."""
        self.service_id = service_id

        # Service data from config file and the same data from
        # management service database are slithgtly different.
        # Convert service data to database format.
        if 'service_type' in service_data:
            service_data = {
                'service-type': service_data.get('service_type'),
                'ip-address': service_data.get('address'),
            }
        self.data = service_data
        self.log = ServiceLogger(service_id)

    def __repr__(self):
        """Printable representation of an object."""
        return f'<Service id={self.service_id} type={self.service_type}>'

    @property
    def hostname(self):
        """Hostname (str)."""
        return self.data['ip-address'].split(':')[0]

    @property
    def service_type(self):
        """Service type (str)."""
        return self.data['service-type']

    @property
    def service_account_name(self):
        """Service account name (str)."""
        if self.service_type in ['backup', 'log']:
            return 'ivxv-admin'
        if self.service_type == 'proxy':
            return 'haproxy'
        return 'ivxv-' + self.service_type

    @property
    def service_systemctl_id(self):
        """Service ID to use with systemctl (str)."""
        return f'ivxv-{self.service_type}@{self.service_id}'

    @property
    def service_data_dir(self):
        """Path to service data directory (str)."""
        return os.path.join(SERVICE_DIR, self.service_id)

    @property
    def deb_pkg_name(self):
        """Debian package name (str)."""
        return f'ivxv-{self.service_type}'

    def get_db_key(self, name):
        """Get database key for this service."""
        return f'service/{self.service_id}/{name}'

    def register_event(self, event, params, level='INFO'):
        """Register service event."""
        register_service_event(
            event, level=level, service=self.service_id, params=params)

    def register_bg_info(self, bg_info):
        """Register service error message in database."""
        if bg_info:
            self.log.info('Registering background info: %s',
                          bg_info.split('\n')[0])
        with IVXVManagerDb(for_update=True) as db:
            db.set_value(self.get_db_key('bg_info'), bg_info, safe=True)

    def remove_root_access(self):
        """Remove management service access to service host root account.

        :return: True on success, False on error.
        :rtype: bool
        """
        # disable ivxv-admin key in authorized_keys file
        self.log.info(
            "Removing management service access to root account in service host %r",
            self.hostname,
        )
        proc = self.ssh(
            'sudo ivxv-admin-sudo remove-admin-root-access', account='root')

        # report result
        if proc.returncode:
            self.log.info('Failed to remove management service access '
                          'to service host root account')
            return False

        self.log.info('Management service access to service host '
                      'root account removed successfully')
        return True

    def install_service(self):
        """Install service in service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.log.info("Installing service to host %r", self.hostname)

        # initialize service on the first tech config
        with IVXVManagerDb() as db:
            current_cfg_ver = db.get_value(
                self.get_db_key('technical-conf-version'))
        if not current_cfg_ver:
            if not self.init_service():
                return False

        # install service package
        error = not self.install_service_pkg()

        # create SSH access to service account
        if not error and self.service_account_name != 'ivxv-admin':
            self.log.info(
                "Creating access to the service account %r in service host",
                self.service_account_name,
            )
            proc = self.ssh(
                'sudo ivxv-admin-sudo create-ssh-access {}'.format(
                    self.service_account_name),
                account='ivxv-admin')
            error = bool(proc.returncode)

            if error:
                self.log.error('Failed to create access '
                               'to the service account in service host')

        # report result
        if error:
            self.log.error('Failed to install service')
            return False

        self.log.info('Service installed successfully')
        return True

    def install_service_pkg(self, is_update=False):
        """Install service package to service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        pkg_path = cfg_path(
            'deb_pkg_path', COLLECTOR_PKG_FILENAMES[self.deb_pkg_name])

        # check package status in service host
        if not is_update:
            self.log.info(
                'Querying state of the service software package %r version %s',
                self.deb_pkg_name, DEB_PKG_VERSION)
            proc = self.ssh(
                f'dpkg --status {self.deb_pkg_name}',
                account='ivxv-admin',
                stdout=subprocess.PIPE)
            if proc.returncode == 0:
                for line in proc.stdout.decode('UTF-8').split('\n'):
                    if line.find('Version:') == 0:
                        pkg_version = line.split(':')[1].strip()
                        if DEB_PKG_VERSION == pkg_version:
                            self.log.info(
                                "Package %r is already installed",
                                self.deb_pkg_name,
                            )
                            return True
                        self.log.info(
                            "Package %r is installed with wrong version %s",
                            self.deb_pkg_name,
                            pkg_version,
                        )
                        break

        # copy packages to service host
        if not self.scp(
                local_path=pkg_path,
                remote_path=CONFIG['deb_pkg_path'],
                description='software package file',
                account='ivxv-admin'):
            return False

        # install service package
        self.log.info("Installing package %r", self.deb_pkg_name)
        proc = self.ssh(
            'sudo ivxv-admin-sudo install-pkg {}'.format(
                COLLECTOR_PKG_FILENAMES[self.deb_pkg_name]),
            account='ivxv-admin')
        if proc.returncode:
            self.log.error("Failed to install package %r", self.deb_pkg_name)
            return False
        self.log.info("Package %r is installed successfully", self.deb_pkg_name)

        return True

    def update_ivxv_common_pkg(self):
        """Update ivxv-common package in service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        # copy package file to service host
        pkg_name = 'ivxv-common'
        pkg_path = cfg_path('deb_pkg_path', COLLECTOR_PKG_FILENAMES[pkg_name])
        if not self.scp(
                local_path=pkg_path,
                remote_path=CONFIG['deb_pkg_path'],
                description='ivxv-common package file',
                account='ivxv-admin'):
            return False

        # install ivxv-common package
        self.log.info('Installing ivxv-common package')
        proc = self.ssh(
            'sudo ivxv-admin-sudo install-pkg {}'.format(
                COLLECTOR_PKG_FILENAMES[pkg_name]),
            account='ivxv-admin')
        if proc.returncode:
            self.log.info('Failed to install {} package', pkg_name)
            return False

        self.log.info("Package %r is installed successfully", pkg_name)
        return True

    def init_service(self):
        """Initialize service.

        :return: True on success, False on error.
        :rtype: bool
        """
        # remove service package with all dependent packages
        if not self.remove_service_pkg():
            return False

        # initialize service data
        self.log.info('Initializing service data directory in service host')

        service_id = (
            'backup' if self.service_type == 'backup' else self.service_id)
        proc = self.ssh(
            f'sudo ivxv-admin-sudo init-service {service_id}',
            account='ivxv-admin')

        if proc.returncode:
            self.log.error('Failed to initialize service data directory')
            return False

        self.log.info('Service data directory initialized successfully')
        return True

    def init_service_host(self):
        """Initialize service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.log.info('Initializing service host')

        # install ivxv-common package in service host if required
        self.log.info('Detect ivxv-common package status')
        proc = self.ssh('dpkg --list ivxv-common', account='ivxv-admin')
        if proc.returncode:
            self.log.info('Failed to detect ivxv-common package status')
            # detect ivxv-common package dependencies
            pkg_path = cfg_path(
                'deb_pkg_path', COLLECTOR_PKG_FILENAMES['ivxv-common'])

            # copy package file to service host
            remote_path = os.path.join(
                '/root', COLLECTOR_PKG_FILENAMES['ivxv-common'])
            if not self.scp(
                    local_path=pkg_path,
                    remote_path=remote_path,
                    description='ivxv-common package file',
                    account='root'):
                return False

            # update package list
            self.log.info('Updating package list')
            cmd = [
                'env', 'TERM=dumb', 'DEBIAN_FRONTEND=noninteractive',
                'apt-get', 'update'
            ]
            proc = self.ssh(cmd, account='root')
            if proc.returncode:
                self.log.info('Failed to update package list')
                return False

            # install ivxv-common package
            if not self.install_ivxv_common_pkg(pkg_path, remote_path):
                return False

            # create management service access to
            # service host ivxv-admin account
            self.log.info('Creating access to the management account '
                          '"ivxv-admin" in service host')
            cmd = [
                'tee', '--append',
                os.path.join(
                    os.path.expanduser('~ivxv-admin'), '.ssh',
                    'authorized_keys')
            ]
            with open(os.path.expanduser(IVXV_ADMIN_SSH_PUBKEY_FILE),
                      'rb') as key_fp:
                proc = self.ssh(cmd, account='root', stdin=key_fp)
            if proc.returncode:
                self.log.error('Failed to remove create SSH access '
                               'to management account in service host')
                return False

            # remove management service access to service host root account
            if not self.remove_root_access():
                return False

        # initialize service host
        proc = self.ssh('sudo ivxv-admin-sudo init-host', account='ivxv-admin')
        if proc.returncode:
            self.log.error('Failed to initialize service host')
            return False

        self.log.info('Service host initialized successfully')
        return True

    def install_ivxv_common_pkg(self, pkg_path, remote_path):
        """Install ivxv-common package with dependent packages."""
        # detect package dependencies
        deb = debfile.DebFile(pkg_path)
        deps = []
        for field in 'Depends', 'Recommends':
            deps_str = deb.control.debcontrol().get(field)
            if deps_str is not None:
                deps += [dep.split(' ')[0] for dep in deps_str.split(', ')]

        # install package dependecies
        self.log.info('Installing ivxv-common package dependencies')
        cmd = [
            'env', 'TERM=dumb', 'DEBIAN_FRONTEND=noninteractive',
            'apt-get', '--yes', 'install'
        ] + deps
        proc = self.ssh(cmd, account='root')
        if proc.returncode:
            self.log.info('Failed to install ivxv-common package dependencies')
            return False

        # install package
        self.log.info('Installing ivxv-common package')
        proc = self.ssh(f'dpkg --install {remote_path}', account='root')
        if proc.returncode:
            self.log.info('Failed to install ivxv-common package')
            return False

        return True

    def remove_service_pkg(self):
        """Remove service package with all dependent packages in service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        # check package state
        self.log.info("Checking package %r state in service host", self.deb_pkg_name)
        cmd = f'dpkg --list {self.deb_pkg_name}'
        proc = self.ssh(cmd, account='ivxv-admin', stdout=subprocess.PIPE)
        if proc.returncode:
            self.log.info(
                "Package %r is not installed in service host",
                self.deb_pkg_name,
            )
            return True

        # detect packages to remove
        pkg_to_remove = []
        for line in proc.stdout.decode().split('\n'):
            if line[:2] != 'un':
                for key in COLLECTOR_PKG_FILENAMES:
                    if f' {key} ' in line:
                        assert key != 'ivxv-common'
                        pkg_to_remove.append(key)

        # remove packages
        self.log.info("Removing package %r in service host", self.deb_pkg_name)
        cmd = 'apt-get purge --yes {}'.format(' '.join(pkg_to_remove))
        proc = self.ssh(cmd, account='root', stdout=subprocess.PIPE)
        if proc.returncode:
            self.log.info(
                "Failed to remove package %r in service host",
                self.deb_pkg_name,
            )
            if proc.stdout:
                for line in proc.stdout.decode().split('\n'):
                    self.log.info(line)
            if proc.stderr:
                for line in proc.stderr.decode().split('\n'):
                    self.log.error(line)
            return False

        self.log.info(
            "Package %r in service host removed successfully",
            self.deb_pkg_name,
        )
        return True

    def restart_service(self):
        """Restart service.

        :return: True on success, False on error.
        :rtype: bool
        """
        # detect service state
        with IVXVManagerDb() as db:
            state = db.get_all_values('service')
        state[self.service_id]['bg_info'] = None
        generate_service_hints(state)

        # skip restarting of unconfigured service
        if state[self.service_id]['bg_info']:
            self.log.info(
                'Skipping restart of unconfigured service: %s',
                state[self.service_id]['bg_info'])
            return True

        # restart service
        self.log.info('Restarting service')

        cmd = [
            'ivxv-admin-helper', 'restart-service', self.service_type,
            self.service_id, self.service_systemctl_id
        ]
        proc = self.ssh(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        error = bool(proc.returncode)

        if error:
            self.register_bg_info(
                'Service restart error: {}'
                .format(proc.stderr.decode('UTF-8').strip()))
            self.log.error('Failed to restart service')
            return False

        with IVXVManagerDb(for_update=True) as db:
            self.register_state(db, SERVICE_STATE_CONFIGURED)
        self.log.info('Service restarted successfully')
        return True

    def stop_service(self):
        """Stop service.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.log.info('Stopping service')

        cmd = ['systemctl', 'stop', '--user', self.service_systemctl_id]
        proc = self.ssh(cmd)
        error = bool(proc.returncode)

        if error:
            self.log.error('Failed to stop service')
            return False

        self.log.info('Service stopped successfully')
        return True

    def configure_logging(self, tech_cfg):
        """Configure rsyslog logging in service host.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.log.info("Configuring syslog logging in service host %r", self.hostname)

        # prepare log collector services data
        log_collectors = []
        for network in tech_cfg['network']:
            for service_type, services in network['services'].items():
                if service_type == 'log' and services:
                    for service in services:
                        log_collectors.append({
                            'id': service['id'],
                            'address': service['address'].split(':')[0],
                            'port': service['address'].split(':')[1],
                        })

        # prepare external log collectors data (first record is Log Monitor)
        ext_log_collectors = tech_cfg.get('logging', [])

        # load rsyslog config template
        tmpl_env = Environment(loader=PackageLoader('ivxv_admin', 'templates'))
        template = tmpl_env.get_template('ivxv_service_rsyslog_conf.jinja')

        # create rsyslog config content
        rsyslog_cfg = template.render(
            cfg_filepath=RSYSLOG_CFG_FILENAME,
            log_collectors=log_collectors,
            ext_log_collectors=ext_log_collectors)

        # read existing config file
        cmd = 'test -f {filename} && cat {filename}'.format(
            filename=RSYSLOG_CFG_FILENAME)
        proc = self.ssh(cmd, account='ivxv-admin', stdout=subprocess.PIPE)
        existing_rsyslog_cfg = proc.stdout.decode()

        # don't replace file if not required
        if rsyslog_cfg == existing_rsyslog_cfg:
            self.log.info(
                "Rsyslog config file %r already exist with required content",
                RSYSLOG_CFG_FILENAME,
            )
            return True

        # write config file
        temp_file = tempfile.NamedTemporaryFile(delete=False)
        temp_file.write(bytes(rsyslog_cfg, 'UTF-8'))
        temp_file.close()
        if not self.scp(
                local_path=[temp_file.name],
                remote_path='/etc/ivxv/ivxv-logging.conf',
                description='logging configuration file',
                account='ivxv-admin'):
            return False
        os.unlink(temp_file.name)

        # restart rsyslog service
        proc = self.ssh(
            'sudo ivxv-admin-sudo rsyslog-config-apply', account='ivxv-admin')

        if proc.returncode:
            self.log.info('Failed to restart rsyslog server in service host')
            return False

        self.log.info('Logging configuration for service created successfully')
        return True

    def load_apply_state(self, filepath, attempt_no=0):
        """Load config applying state from state file.

        :return: Attempt number
        :rtype: int
        """
        self.cfg_state_filepath = filepath
        with open(filepath) as fp:
            self.cfg_state = json.load(fp)

        if not attempt_no:
            self.cfg_state['attempts'] += 1
            self.cfg_state['log'].append([])
            attempt_no = self.cfg_state['attempts']
        assert attempt_no == self.cfg_state['attempts']
        assert attempt_no <= len(self.cfg_state['log'])
        self.update_apply_state()

        return attempt_no

    def update_apply_state(self, **kw):
        """Update config applying state file."""
        self.cfg_state.update(kw)
        self.cfg_state['log'][-1] += self.log.storage
        self.log.storage = []
        tmp_filepath = self.cfg_state_filepath + '.tmp'
        with open(tmp_filepath, 'x') as fp:
            json.dump(self.cfg_state, fp, indent=4, sort_keys=True)
        shutil.move(tmp_filepath, self.cfg_state_filepath)

    def apply_tech_cfg(self, cfg_ver, tech_cfg):
        """Apply collector technical config to service.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.update_apply_state()

        # install service package to service host
        if not self.install_service():
            return False
        self.update_apply_state()

        # configure service logging
        if self.service_type != 'log' and not self.configure_logging(tech_cfg):
            return False
        self.update_apply_state()

        # mark services that does not require config packages as CONFIGURED
        # after install
        if not SERVICE_TYPE_PARAMS[self.service_type]['require_config']:
            self.register_cfg_version(
                'technical', cfg_ver, SERVICE_STATE_CONFIGURED)
            install_backup_crontab()
            return True

        # create data directory for service
        proc = self.ssh(f'mkdir --verbose --parents {self.service_data_dir}')
        if proc.returncode:
            self.log.error('Failed to generate service data directory')
            return False
        self.update_apply_state()

        # apply trust root config to service
        self.log.info('Applying trust root config to service')
        if not self.copy_cfg_to_service('trust'):
            self.log.error('Failed to apply trust root config to service')
            return False
        self.log.info('Trust root config successfully applied to service')
        self.update_apply_state()

        # apply technical config to service
        self.log.info('Applying technical config to service')
        if not self.copy_cfg_to_service('technical'):
            self.log.error('Failed to apply technical config to service')
            return False
        self.update_apply_state()

        # notify service about config change
        with IVXVManagerDb() as db:
            election_cfg_ver = db.get_value(
                self.get_db_key('election-conf-version'))
        if election_cfg_ver and not self.restart_service():
            return False
        self.log.info('Technical config successfully applied to service')
        self.update_apply_state()

        # detect service current state
        with IVXVManagerDb() as db:
            service_old_state = db.get_value(self.get_db_key('state'))

        # register config version (and new service state for NOT INSTALLED
        # service) in management database
        self.register_cfg_version(
            'technical', cfg_ver,
            SERVICE_STATE_INSTALLED
            if service_old_state == SERVICE_STATE_NOT_INSTALLED
            else None
        )

        return True

    def apply_election_cfg(self, cfg_ver):
        """Apply elections config to service.

        :return: True on success, False on error.
        :rtype: bool
        """
        # detect service state
        with IVXVManagerDb() as db:
            state = db.get_all_values('service')

        # service issues must be fixed before installing first election config
        if not state[self.service_id]['election-conf-version']:
            state = {self.service_id: state[self.service_id]}
            generate_service_hints(state)
            if state[self.service_id].get('bg_info') not in [
                    None, 'Apply election config'
            ]:
                self.log.error('Cannot apply election config: %s',
                               state[self.service_id].get('bg_info'))
                return False
        del state

        self.log.info('Applying election config to service')
        if not self.copy_cfg_to_service('election'):
            return False
        self.update_apply_state()

        # enable service systemd unit if first config is loaded
        self.log.info('Enabling systemd unit')
        cmd = f'systemctl enable --user {self.service_systemctl_id}'
        proc = self.ssh(cmd)
        if proc.returncode:
            self.log.error('Failed to enable systemd unit')
            return False
        self.update_apply_state()

        # register config version and service state in management database
        self.register_cfg_version(
            'election', cfg_ver, SERVICE_STATE_CONFIGURED)

        # notify service about config change
        if not self.restart_service():
            return False
        self.update_apply_state()

        # install crontab for automatic backups
        if self.service_type == 'voting':
            install_backup_crontab()

        return True

    def apply_list(self, list_type, changeset_no=None):
        """Apply choices list to service.

        :param list_type: List type (choices, voters)
        :type list_type: str
        :param changeset_no: Changeset number for voter list
        :type changeset_no: int
        :return: True on success, False on error.
        :rtype: bool
        """
        # apply list
        self.log.info('Applying %s list to service', list_type)
        if not self.copy_cfg_to_service(list_type, changeset_no):
            self.log.error('Failed to apply %s list to service', list_type)
            return False
        self.update_apply_state()

        # register list version
        with IVXVManagerDb(for_update=True) as db:
            if list_type in ["choices", "districts"]:
                cfg_ver = db.get_value(f"list/{list_type}")
                self.log.info(
                    "Registering applied %s list version %r in management database",
                    list_type,
                    cfg_ver,
                )
                db.set_value(f"list/{list_type}-loaded", cfg_ver)
            else:
                assert list_type == "voters"
                cfg_ver = db.get_value(f"list/voters{changeset_no:04}")
                self.log.info(
                    "Registering applied voters list #%d "
                    "version %r in management database",
                    changeset_no,
                    cfg_ver,
                )
                db.set_value(f"list/voters{changeset_no:04}-state", "APPLIED")

        self.log.info('%s list applied successfully to service',
                      list_type.capitalize())
        return True

    def load_secret_file(self, secret_type, filepath, checksum):
        """Load secret key file to service and register file checksum.

        :param secret_type: Secret type
        :type secret_type: str
        :param filepath: Source file path
        :type filepath: str
        :param checksum: File checksum
        :type checksum: str
        :return: True on success, False on error.
        :rtype: bool
        """
        secret_descr = SERVICE_SECRET_TYPES[secret_type]['description']
        target_path = (SERVICE_SECRET_TYPES[secret_type]['target-path']
                       .format(service_id=self.service_id))
        shared_sercet = SERVICE_SECRET_TYPES[secret_type]['shared']
        file_mode = '0640' if shared_sercet else '0600'
        remote_account = ('ivxv-admin' if shared_sercet
                          else self.service_account_name)

        # copy file to service host
        self.log.info('Loading %s to service', secret_descr)
        if not self.scp(
                local_path=[filepath],
                remote_path=target_path,
                account=remote_account,
                description=secret_descr):
            return False

        # set file permissions
        proc = self.ssh(
            f'chmod --changes {file_mode} {target_path}',
            account=remote_account)
        if proc.returncode:
            self.log.error('Failed to set %s file permissions '
                           'in service host', secret_descr)
            return False

        # register key checksum
        with IVXVManagerDb(for_update=True) as db:
            self.log.info(
                'Registering loaded %s checksum in management database',
                secret_descr)
            db.set_value(
                self.get_db_key(SERVICE_SECRET_TYPES[secret_type]['db-key']),
                checksum)

        self.log.info('%s loaded successfully to service', secret_descr)

        # restart service
        return self.restart_service()

    def copy_cfg_to_service(self, cfg_type, changeset_no=None):
        """Copy config to service host and reload service if needed.

        :param cfg_type: Config type
                         (trust, technical, elections, choices, voters)
        :type cfg_type: str
        :param changeset_no: Changeset number for voter list
        :type changeset_no: int
        :return: True on success, False on error.
        :rtype: bool
        """
        # copy config to host
        cfg_filename_suffix = f"{changeset_no:04}" if cfg_type == "voters" else ""
        cfg_filename_ext = "zip" if changeset_no else "bdoc"
        cfg_filename = f"{cfg_type}{cfg_filename_suffix}.{cfg_filename_ext}"
        cfg_filepath = cfg_path('active_config_files_path', cfg_filename)
        target_path = f'/etc/ivxv/{cfg_filename}'
        if not self.scp(
                local_path=[cfg_filepath],
                remote_path=target_path,
                account='ivxv-admin',
                description=f'{cfg_type} config'):
            return False

        # set config file permissions
        self.log.info('Set %s config file permissions in service host',
                      cfg_type)
        cmd = ['chmod', '--changes', '0640', target_path]
        proc = self.ssh(cmd, account='ivxv-admin')
        if proc.returncode:
            self.log.error('Failed to set %s config file permissions '
                           'in service host', cfg_type)
            return False

        # notify choices service about list change
        list_change_util = {
            'choices': 'ivxv-choiceimp',
            'districts': 'ivxv-districtimp',
            'voters': 'ivxv-voterimp'
        }.get(cfg_type)
        if list_change_util:
            self.log.info('Notify service about %s list change', cfg_type)
            cmd = [list_change_util, '-instance', self.service_id, target_path]
            proc = self.ssh(cmd)
            if proc.returncode:
                self.log.info('Failed to notify service about %s list change',
                              cfg_type)
                return False

        return True

    def register_cfg_version(self, cfg_type, cfg_ver, service_state):
        """Register config version and service state in database."""
        with IVXVManagerDb(for_update=True) as db:
            self.log.info(
                "Registering %s config version %r in management database",
                cfg_type,
                cfg_ver,
            )
            db.set_value(
                self.get_db_key('%s-conf-version' % cfg_type), cfg_ver)

            if service_state is not None:
                self.register_state(db, service_state)

    def register_state(self, db, state, bg_info=None):
        """Register service state in database."""
        last_state = db.get_value(self.get_db_key('state'))
        if state == last_state:
            self.log.info("Service state %r not changed", state)
            return

        # set background info
        if state in SERVICE_STATE_CONFIGURED:
            bg_info = ''
        if bg_info is not None:
            db.set_value(self.get_db_key('bg_info'), bg_info, safe=True)

        # set service state
        self.log.info(
            "Registering service state as %r "
            "in management database (last state: %r)",
            state,
            last_state,
        )
        db.set_value(self.get_db_key('state'), state, safe=True)
        self.register_event(
            'SERVICE_STATE_CHANGE',
            params={
                'state': state,
                'last_state': last_state
            })

    def ping(self):
        """Ping service.

        :return: True on success, False on error.
        :rtype: bool
        """
        self.log.debug('Pinging service')

        # query service status
        cfg_check_enabled = False
        if self.service_type == 'backup':
            ping_cmd = 'true'
        elif self.service_type == 'log':
            ping_cmd = 'systemctl status rsyslog'
        else:
            ping_cmd = 'env LC_ALL=C systemctl status --user {}'.format(
                self.service_systemctl_id)
            cfg_check_enabled = True
        ping_proc = self.ssh(
            ping_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        if not ping_proc.returncode:
            if cfg_check_enabled:
                try:
                    self._verify_cfg_version()
                except LookupError as err:
                    self.register_bg_info(str(err))
                    self.log.error('Failed to detect service status')
                    return False
            self.register_bg_info('')
            self.log.debug('Service is alive')
            return True

        # return without config check
        cmd_err_output = ping_proc.stderr.decode('UTF-8').strip()
        if not cfg_check_enabled or cmd_err_output:
            self.register_bg_info(
                'Ping error: {}'.format(
                    cmd_err_output or
                    ping_proc.stdout.decode('UTF-8').strip() or
                    f"Command {' '.join(ping_cmd)!r} failed"))
            self.log.error('Pinging service failed')
            return False

        # check service config
        cfg_check_cmd = [
            'ivxv-admin-helper', 'check-service-config', self.service_type,
            self.service_id
        ]
        cfg_check_proc = self.ssh(
            cfg_check_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if cfg_check_proc.returncode == 0:  # valid config
            self.register_bg_info(
                'Ping error: {}'.format(
                    ping_proc.stdout.decode('UTF-8').strip() or
                    f"Command {' '.join(ping_cmd)!r} failed"))
            self.log.error('Pinging service failed')
        else:  # invalid config
            self.register_bg_info(
                'Ping error: Invalid config: {}'.format(
                    cfg_check_proc.stderr.decode('UTF-8').strip() or
                    cfg_check_proc.stdout.decode('UTF-8').strip()))
            self.log.error('Pinging service failed (invalid configuration)')

        return False

    def _get_cfg_versions_from_service(self):
        """Get config versions from service.

        Service version data is provided by commands:

        * ``systemctl show`` - trust, technical and election config
        * ``ivxv-choiceimp`` - choices list
        * ``ivxv-districtimp`` - districts list
        * ``ivxv-voterimp`` - voters lists

        :raises LookupError: if some internal check fails
        """
        cfg_versions = {
            'service_state': {},
            'choices_list_versions': [],
            'districts_list_versions': [],
            'voters_lists_versions': [],
        }

        # query status text for trust, technical and election config
        cmd = ('systemctl --user show --property StatusText '
               f'{self.service_systemctl_id}')
        proc = self.ssh(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc.returncode:
            raise LookupError(
                'Error while querying service status text: {}'.format(
                    proc.stderr.decode('UTF-8')))
        status_text = proc.stdout.decode('UTF-8')

        # parse status
        re_pattern = re.compile(r'StatusText=(.+)')
        if not re_pattern.match(status_text):
            raise LookupError(f'Invalid service state text: {status_text}')
        state_json = re_pattern.sub('\\1', status_text)
        try:
            cfg_versions['service_state'] = json.loads(state_json)
        except json.decoder.JSONDecodeError as err:
            raise LookupError(
                f'Error while decoding JSON data from service state: {err}. '
                f'JSON: {state_json}')
        self.log.debug('Service internal status: %s',
                       cfg_versions['service_state']['Status'])

        # query status text for choices and districts list
        if self.service_type == 'choices':
            for cfg_type in ["choices", "districts"]:
                utility = f"ivxv-{cfg_type[:-1]}imp"
                cmd = f"{utility} --check version --instance {self.service_id}"
                proc = self.ssh(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                if proc.returncode:
                    raise LookupError(
                        f"Error while querying {cfg_type} list version: "
                        f"{proc.stderr.decode('UTF-8')}"
                    )
                status_text = proc.stdout.decode("UTF-8")
                try:
                    list_versions = json.loads(status_text) if status_text else []
                    assert isinstance(list_versions, list)
                except json.decoder.JSONDecodeError as err:
                    raise LookupError(
                        "Error while decoding JSON data from service state: "
                        f"{err}. JSON: {status_text}"
                    )

        # query status text for voters lists
        elif self.service_type == 'voting':
            cmd = (
                f'ivxv-voterimp --check version --instance {self.service_id}')
            proc = self.ssh(
                cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            if proc.returncode:
                raise LookupError(
                    'Error while querying voters list versions: {}'.format(
                        proc.stderr.decode('UTF-8')))
            status_text = proc.stdout.decode('UTF-8')
            try:
                cfg_versions['voters_lists_versions'] = (
                    json.loads(status_text) if status_text else [])
                assert isinstance(cfg_versions['voters_lists_versions'], list)
            except json.decoder.JSONDecodeError as err:
                raise LookupError(
                    'Error while decoding JSON data from service state: '
                    f'{err}. JSON: {status_text}')

        return cfg_versions

    def _verify_cfg_version(self):
        """Verify config version.

        Compare config version data provided by service and version in
        management database.

        :return: Message to register in service background info
        :raises LookupError: if some internal check fails
        """
        cfg_versions = self._get_cfg_versions_from_service()
        version_in_db = None
        with IVXVManagerDb() as db:
            for cfg_type in 'trust', 'technical', 'election':
                db_key = (
                    'config/trust' if cfg_type == 'trust' else
                    self.get_db_key(f'{cfg_type}-conf-version'))
                version_in_db = db.get_value(db_key)
                versions_in_service = cfg_versions['service_state']['Version'][
                    cfg_type.capitalize()]
                if version_in_db not in versions_in_service:
                    raise LookupError(
                        f"{cfg_type.capitalize()} config version "
                        "in service status info "
                        f"does not contain version {version_in_db!r} "
                        "registered in management database. "
                        f"Service status info: {versions_in_service}"
                    )
            for list_type in ["choices", "districts"]:
                versions_in_service = cfg_versions[f"{list_type}_list_versions"]
                if versions_in_service:
                    version_in_db = db.get_value(f"list/{list_type}")
                    if version_in_db not in versions_in_service:
                        raise LookupError(
                            f"{list_type.capitalize()} list version "
                            "in service status info "
                            f"does not contain version {version_in_db!r} "
                            "as registered in management database. "
                            "Choices list versions in services: "
                            f"{versions_in_service}"
                        )

            version_lists_in_service = cfg_versions['voters_lists_versions']
            if version_lists_in_service:
                for changeset_no in range(10_000):
                    key = f"list/voters{changeset_no:04d}-loaded"
                    try:
                        version_in_db = db.get_value(key)
                    except KeyError:
                        break
                    try:
                        versions_in_service = version_lists_in_service[
                            changeset_no - 1]
                    except IndexError:
                        raise LookupError(
                            "Service status info does not contain "
                            "version info for voters list "
                            f"#{changeset_no:04d}. "
                            "Version for this list in management database "
                            f"is {version_in_db!r}"
                        )
                    if version_in_db not in versions_in_service:
                        raise LookupError(
                            f"Voters list #{changeset_no:04d} version "
                            "in service status info "
                            f"does not contain version {version_in_db!r} "
                            "as registered in management database. "
                            "Voters list versions in services: "
                            f"{versions_in_service}"
                        )

    def scp(self,
            local_path,
            remote_path,
            description,
            account=None,
            to_remote=True):
        """Copy file using SCP protocol.

        :param local_path: Local file path(s)
        :type local_path: str or list of strings
        :param remote_path: Remote file path
        :type remote_path: str
        :param description: Description of files to copy (for logging)
        :type description: str
        :param account: Remote account name (None = default service account)
        :type account: str
        :param to_remote: Copy to remote host (True) or from remote host
        :type to_remote: bool
        :return: True on success, False on error.
        :rtype: bool
        """
        if not isinstance(local_path, list):
            local_path = [local_path]
        assert to_remote or len(local_path) == 1
        remote_path = '{}@{}:{}'.format(account or self.service_account_name,
                                        self.hostname, remote_path)
        direction_str = 'to' if to_remote else 'from'

        # prepare command
        cmd = ['scp'] + (local_path + [remote_path]
                         if to_remote else [remote_path] + local_path)

        # execute command
        self.log.info('Copying %s %s service host', description, direction_str)
        proc = exec_remote_cmd(cmd)

        # return results
        if proc.returncode:
            self.log.error('Failed to copy %s %s service host',
                           description, direction_str)

        return proc.returncode == 0

    def ssh(self, cmd, account=None, fwd_auth_agent=False, **kw):
        """Execute command in remote host.

        :param cmd: Command to execute
        :type cmd: str or list
        :param account: Remote account name
        :type account: str
        :param fwd_auth_agent: Start ssh with authentication agent forwarding
        :type fwd_auth_agent: bool
        :param kw: Command arguments

        :return: subprocess.CompletedProcess
        """
        if not isinstance(cmd, list):
            cmd = [cmd]
        ssh_cmd = ['ssh', '-A'] if fwd_auth_agent else ['ssh']
        ssh_cmd.append(
            '{}@{}'.format(
                account or self.service_account_name, self.hostname))

        return exec_remote_cmd(ssh_cmd + cmd, **kw)
