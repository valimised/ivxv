# IVXV Internet voting framework
"""
Service management module for collector management service.
"""

import logging
import os
import subprocess
import tempfile

from debian import debfile
from jinja2 import Environment, PackageLoader

from . import lib
from . import SERVICE_SECRET_TYPES
from . import SERVICE_STATE_CONFIGURED, SERVICE_STATE_INSTALLED
from .config import CONFIG
from .db import IVXVManagerDb
from .ivxv_pkg import COLLECTOR_PACKAGE_FILENAMES


IVXV_ADMIN_SSH_PUBKEY_FILE = os.path.expanduser('~') + '/.ssh/id_rsa.pub'
RSYSLOG_CONFIG_FILENAME = '/etc/rsyslog.d/ivxv-logging.conf'
SERVICE_DIRECTORY = os.path.join(CONFIG['ivxv_admin_data_path'], 'service')

# create logger
log = logging.getLogger(__name__)


class ServiceLogger():
    """
    Logger wrapper for service object to prepend service ID for every logged
    message. Message level is also prepended for levels other than INFO.
    """
    log_prefix = None  #: Prefix for log messages (str)

    def __init__(self, service_id):
        """Constructor."""
        self.log_prefix = 'SERVICE %s: ' % service_id

    def __getattr__(self, name):
        """Get wrapper for logger method."""
        log_method = getattr(log, name)

        def log_method_wrapper(*args):
            """Wrapper for logger method to prepend log message prefix."""
            args = list(args)
            if name != 'info':
                args[0] = '{}: {}'.format(name.upper(), args[0])
            args[0] = self.log_prefix + args[0]
            return log_method(*tuple(args))

        return log_method_wrapper


class Service():
    """Service management for collector services."""
    service_id = None  #: Service ID (str)
    data = None  #: Service data (dict)
    log = None  #: Logger for service

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

    @property
    def hostname(self):
        """Hostname."""
        return self.data['ip-address'].split(':')[0]

    @property
    def service_type(self):
        """Service type."""
        return self.data['service-type']

    @property
    def service_account_name(self):
        """Service account name."""
        return ('haproxy' if self.service_type == 'proxy'
                else 'ivxv-' + self.service_type)

    @property
    def service_systemctl_id(self):
        """Service ID to use with systemctl."""
        return 'ivxv-{}@{}'.format(self.service_type, self.service_id)

    @property
    def service_data_dir(self):
        """Service data directory."""
        return os.path.join(SERVICE_DIRECTORY, self.service_id)

    def get_db_key(self, name):
        """Get database key for this service."""
        return 'service/{}/{}'.format(self.service_id, name)

    def remove_root_access(self):
        """Remove management service access to service host root account."""
        # disable ivxv-admin key in authorized_keys file
        self.log.info('Removing management service access '
                      'to root account in service host "%s"', self.hostname)
        proc = self.ssh('sudo ivxv-admin-helper remove-admin-root-access',
                        account='root')

        # report result
        if proc.returncode:
            self.log.info('Failed to remove management service access '
                          'to service host root account')
            return False

        self.log.info('Management service access to service host '
                      'root account removed successfully')
        return True

    def install_service(self):
        """Install service in service host."""
        self.log.info('Installing service to host "%s"', self.hostname)

        # install service package
        error = not self.install_service_pkg()

        # create SSH access to service account
        if not error:
            self.log.info('Creating access '
                          'to the service account "%s" in service host',
                          self.service_account_name)
            proc = self.ssh('sudo ivxv-admin-helper create-ssh-access {}'
                            .format(self.service_account_name),
                            account='ivxv-admin')
            error = bool(proc.returncode)

            if error:
                self.log.error('Failed to create access '
                               'to the service account in service host')

        # report result
        if error:
            self.log.error('Failed to install service')
        else:
            self.log.info('Service installed successfully')
        return not error

    def install_service_pkg(self):
        """Install service package to service host."""
        pkg_name = 'ivxv-' + self.service_type
        pkg_path = os.path.join(CONFIG.get('deb_pkg_path'),
                                COLLECTOR_PACKAGE_FILENAMES[pkg_name])

        # check package status in service host
        self.log.info('Querying state of the service software package "%s"',
                      pkg_name)
        proc = self.ssh('dpkg --status {}'.format(pkg_name),
                        account='ivxv-admin', stdout=subprocess.PIPE)
        if proc.returncode == 0:
            return True

        # copy packages to service host
        if not self.scp(local_path=pkg_path,
                        remote_path=CONFIG.get('deb_pkg_path'),
                        description='software package file',
                        account='ivxv-admin'):
            return False

        # initialize service
        if not self.init_service(pkg_name):
            return False

        # configure APT sources for ivxv-storage
        if self.service_type == 'storage':
            self.log.info('Configuring APT sources for package "etcd"')
            proc = self.ssh('sudo ivxv-admin-helper configure-etcd-apt-source',
                            account='ivxv-admin')
            if proc.returncode:
                self.log.error('Failed to configure APT sources '
                               'for package "etcd"')
                return False

        # install service package
        self.log.info('Installing package "%s"', pkg_name)
        proc = self.ssh(
            'sudo ivxv-admin-helper install-pkg {}'
            .format(COLLECTOR_PACKAGE_FILENAMES[pkg_name]),
            account='ivxv-admin')
        if proc.returncode:
            self.log.error('Failed to install package "%s"', pkg_name)
            return False
        self.log.info('Package "%s" is installed successfully', pkg_name)

        return True

    def init_service(self, pkg_name):
        """Initialize service."""
        # remove service package with all dependent packages
        if not self.remove_service_pkg(pkg_name):
            return False

        # initialize service data
        self.log.info('Initializing service data directory in service host')

        proc = self.ssh(
            'sudo ivxv-admin-helper init-service {}'.format(self.service_id),
            account='ivxv-admin')

        if proc.returncode:
            self.log.error('Failed to initialize service data directory')
            return False

        self.log.info('Service data directory initialized successfully')
        return True

    def init_service_host(self):
        """Initialize service host."""
        self.log.info('Initializing service host')

        # install ivxv-common package in service host if required
        self.log.info('Detect ivxv-common package status')
        proc = self.ssh('dpkg -l ivxv-common', account='ivxv-admin')
        if proc.returncode:
            self.log.info('Failed to detect ivxv-common package status')
            # detect ivxv-common package dependencies
            pkg_path = os.path.join(
                CONFIG.get('deb_pkg_path'),
                COLLECTOR_PACKAGE_FILENAMES['ivxv-common'])
            deb = debfile.DebFile(pkg_path)
            deps = []
            for field in 'Depends', 'Recommends':
                deps_str = deb.control.debcontrol().get(field)
                if deps_str is not None:
                    deps += [dep.split(' ')[0] for dep in deps_str.split(', ')]

            # copy package file to service host
            remote_path = os.path.join(
                '/root', COLLECTOR_PACKAGE_FILENAMES['ivxv-common'])
            if not self.scp(local_path=pkg_path,
                            remote_path=remote_path,
                            description='ivxv-common package file',
                            account='root'):
                return False

            # install ivxv-common package dependecies
            self.log.info('Updating package list')
            cmd = ['env', 'TERM=dumb', 'DEBIAN_FRONTEND=noninteractive',
                   'apt-get', 'update']
            proc = self.ssh(cmd, account='root')
            if proc.returncode:
                self.log.info('Failed to update package list')
                return False

            # install ivxv-common package dependecies
            self.log.info('Installing ivxv-common package dependencies')
            cmd = ['env', 'TERM=dumb', 'DEBIAN_FRONTEND=noninteractive',
                   'apt-get', '--yes', 'install'] + deps
            proc = self.ssh(cmd, account='root')
            if proc.returncode:
                self.log.info('Failed to install ivxv-common package '
                              'dependencies')
                return False

            # install ivxv-common package
            self.log.info('Installing ivxv-common package')
            proc = self.ssh('dpkg -i {}'.format(remote_path), account='root')
            if proc.returncode:
                self.log.info('Failed to install ivxv-common package')
                return False

            # create management service access to
            # service host ivxv-admin account
            self.log.info('Creating access to the management account '
                          '"ivxv-admin" in service host')
            cmd = 'tee --append /home/ivxv-admin/.ssh/authorized_keys'
            with open(IVXV_ADMIN_SSH_PUBKEY_FILE, 'rb') as key_fp:
                proc = self.ssh(cmd, account='root', stdin=key_fp)
            if proc.returncode:
                self.log.error('Failed to remove create SSH access '
                               'to management account in service host')
                return False

            # remove management service access to service host root account
            if not self.remove_root_access():
                return False

        # initialize service host
        proc = self.ssh('sudo ivxv-admin-helper init-host',
                        account='ivxv-admin')
        if proc.returncode:
            self.log.error('Failed to initialize service host')
            return False

        self.log.info('Service host initialized successfully')
        return True

    def remove_service_pkg(self, pkg_name):
        """
        Remove service package with all dependent packages in service host.
        """
        # check package state
        self.log.info('Checking package "%s" state in service host', pkg_name)
        cmd = 'dpkg --list {}'.format(pkg_name)
        proc = self.ssh(cmd, account='ivxv-admin', stdout=subprocess.PIPE)
        if proc.returncode:
            self.log.info('Package "%s" is not installed in service host',
                          pkg_name)
            return True

        # detect packages to remove
        pkg_to_remove = []
        for line in proc.stdout.decode().split('\n'):
            if line[:2] != 'un':
                for key in COLLECTOR_PACKAGE_FILENAMES:
                    if ' {} '.format(key) in line:
                        assert key != 'ivxv-common'
                        pkg_to_remove.append(key)

        # remove packages
        self.log.info('Removing package "%s" in service host', pkg_name)
        cmd = 'apt-get purge --yes {}'.format(' '.join(pkg_to_remove))
        proc = self.ssh(cmd, account='root', stdout=subprocess.PIPE)
        if proc.returncode:
            self.log.info('Failed to remove package "%s" in service host',
                          pkg_name)
            if proc.stdout:
                for line in proc.stdout.decode().split('\n'):
                    self.log.info(line)
            if proc.stderr:
                for line in proc.stderr.decode().split('\n'):
                    self.log.error(line)
            return False

        self.log.info('Package "%s" in service host removed successfully',
                      pkg_name)
        return True

    def restart_service(self):
        """Restart service."""
        # detect service state
        db = IVXVManagerDb()
        status = db.get_all_values('service')
        db.close()
        lib.generate_service_hints(status)

        # skip restarting of unconfigured service
        if 'hint' in status[self.service_id]:
            self.log.info(
                'Skipping restart of unconfigured service: %s',
                status[self.service_id]['hint'])
            return True

        # restart service
        self.log.info('Restarting service')

        cmd = 'systemctl restart --user {}'.format(self.service_systemctl_id)
        proc = self.ssh(cmd)
        error = bool(proc.returncode)

        if error:
            self.log.error('Failed to restart service')
        else:
            self.log.info('Service restarted successfully')

        return not error

    def configure_logging(self, tech_config):
        """Configure rsyslog logging in service host."""
        self.log.info('Configuring syslog logging in service host "%s"',
                      self.hostname)

        # prepare log collector services data
        log_collectors = []
        for network in tech_config['network']:
            for service_type, services in network['services'].items():
                if service_type == 'log' and services is not None:
                    service = services[0]
                    log_collectors.append({
                        'id': service['id'],
                        'address': service['address'].split(':')[0],
                        'port': service['address'].split(':')[1],
                    })

        # prepare log monitor data
        log_monitors = tech_config.get('logging', [])
        assert len(log_monitors) < 2

        # load rsyslog config template
        tmpl_env = Environment(loader=PackageLoader('ivxv_admin', 'templates'))
        template = tmpl_env.get_template('ivxv_service_rsyslog_conf.jinja')

        # create rsyslog config content
        rsyslog_conf = template.render(
            config_filepath=RSYSLOG_CONFIG_FILENAME,
            log_collectors=log_collectors,
            log_monitor=log_monitors[0])

        # read existing config file
        cmd = 'test -f {filename} && cat {filename}'.format(
            filename=RSYSLOG_CONFIG_FILENAME)
        proc = self.ssh(cmd, account='ivxv-admin', stdout=subprocess.PIPE)
        existing_rsyslog_conf = proc.stdout.decode()

        # don't replace file if not required
        if rsyslog_conf == existing_rsyslog_conf:
            self.log.info(
                'Rsyslog config file %s already exist with required content',
                RSYSLOG_CONFIG_FILENAME)
            return True

        # write config file
        temp_file = tempfile.NamedTemporaryFile(delete=False)
        temp_file.write(bytes(rsyslog_conf, 'UTF-8'))
        temp_file.close()
        if not self.scp(local_path=[temp_file.name],
                        remote_path='/etc/ivxv/ivxv-logging.conf',
                        description='logging configuration file',
                        account='ivxv-admin'):
            return False
        os.unlink(temp_file.name)

        # restart rsyslog service
        proc = self.ssh('sudo ivxv-admin-helper rsyslog-config-apply',
                        account='ivxv-admin')

        if proc.returncode:
            self.log.info('Failed to restart rsyslog server in service host')
            return False

        self.log.info('Logging configuration for service created successfully')
        return True

    def apply_tech_conf(self, conf_ver, tech_config):
        """Apply collector technical config to service."""
        # install service package to service host
        if not self.install_service():
            return False

        # configure service logging
        if self.service_type != 'log':
            if not self.configure_logging(tech_config):
                return False

        # mark log collector service as configured after install
        if self.service_type == 'log':
            self.register_config_version(
                'technical', conf_ver, SERVICE_STATE_CONFIGURED)
            return True

        # create data directory for service
        proc = self.ssh('mkdir --verbose --parents {}'
                        .format(self.service_data_dir))
        if proc.returncode:
            self.log.error('Failed to generate service data directory')
            return False

        # apply trust root config to service
        self.log.info('Applying trust root config to service')
        if not self.copy_config_to_service('trust'):
            self.log.error('Failed to apply trust root config to service')
            return False
        self.log.info('Trust root config successfully applied to service')

        # apply technical config to service
        self.log.info('Applying technical config to service')
        if not self.copy_config_to_service('technical'):
            self.log.error('Failed to apply technical config to service')
            return False
        self.log.info('Technical config successfully applied to service')

        # register config version and service state in management database
        self.register_config_version(
            'technical', conf_ver, SERVICE_STATE_INSTALLED)

        return True

    def apply_election_conf(self, conf_ver):
        """Apply elections config to service."""
        self.log.info('Applying election config to service')
        if not self.copy_config_to_service('election'):
            return False

        # enable service systemd unit if first config is loaded
        self.log.info('Enabling systemd unit')
        cmd = 'systemctl enable --user {}'.format(self.service_systemctl_id)
        proc = self.ssh(cmd)
        if proc.returncode:
            self.log.error('Failed to enable systemd unit')
            return False

        # register config version and service state in management database
        self.register_config_version(
            'election', conf_ver, SERVICE_STATE_CONFIGURED)

        # notify service about config change
        if not self.restart_service():
            return False

        return True

    def apply_list(self, list_type, list_no=None):
        """
        Apply choices list to service.

        :param list_type: List type (choices, voters)
        :type list_type: str
        :param list_no: List number (only for voters list)
        :type list_no: int
        :returns: True on success
        """
        # apply list
        self.log.info('Applying %s list to service', list_type)
        if not self.copy_config_to_service(list_type, list_no):
            self.log.error('Failed to apply %s list to service', list_type)
            return False

        # register list version
        db = IVXVManagerDb(for_update=True)
        if list_type == 'choices':
            conf_ver = db.get_value('list/choices')
            self.log.info('Registering applied choices list version "%s" '
                          'in management database', conf_ver)
            db.set_value('list/choices-loaded', conf_ver)
        else:
            conf_ver = db.get_value('list/voters{:02}'.format(list_no))
            self.log.info('Registering applied voters list #%d version "%s" '
                          'in management database', list_no, conf_ver)
            db.set_value('list/voters{:02}-loaded'.format(list_no), conf_ver)
        db.close()

        self.log.info('%s list applied successfully to service',
                      list_type.capitalize())
        return True

    def load_secret_file(self, secret_type, filepath, checksum):
        """
        Load secret key file to service and register
        file checksum in management service database.

        :param secret_type: Secret type
        :type secret_type: str
        :param filepath: Source file path
        :type filepath: str
        :param checksum: File checksum
        :type checksum: str
        :returns: True on success
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
        if not self.scp(local_path=[filepath],
                        remote_path=target_path,
                        account=remote_account,
                        description=secret_descr):
            return False

        # set file permissions
        proc = self.ssh('chmod --changes {} {}'.format(file_mode, target_path),
                        account=remote_account)
        if proc.returncode:
            self.log.error('Failed to set %s file permissions '
                           'in service host', secret_descr)
            return False

        # register key checksum
        db_key = SERVICE_SECRET_TYPES[secret_type]['db-key']
        db = IVXVManagerDb(for_update=True)
        self.log.info('Registering loaded %s checksum in management database',
                      secret_descr)
        db.set_value('service/{}/{}'.format(self.service_id, db_key), checksum)
        db.close()

        self.log.info('%s loaded successfully to service', secret_descr)

        # restart service
        return self.restart_service()

    def copy_config_to_service(self, conf_type, list_no=None):
        """
        Copy config to service host and reload service if needed.

        :param conf_type: Config type
                          (trust, technical, elections, choices, voters)
        :type conf_type: str
        :param list_no: List number (only for voters list)
        :type list_no: int
        :returns: True on success
        """
        # copy config to host
        conf_filename = '{}{}.bdoc'.format(
            conf_type,
            '-{:02}'.format(list_no) if conf_type == 'voters' else '')
        conf_path = os.path.join(CONFIG.get('active_config_files_path'),
                                 conf_filename)
        target_path = '/etc/ivxv/{}'.format(conf_filename)
        if not self.scp(local_path=[conf_path],
                        remote_path=target_path,
                        account='ivxv-admin',
                        description='{} config'.format(conf_type)):
            return False

        # set config file permissions
        self.log.info('Set %s config file permissions in service host',
                      conf_type)
        cmd = 'chmod --changes 0640 {}'.format(target_path)
        proc = self.ssh(cmd, account='ivxv-admin')
        if proc.returncode:
            self.log.error('Failed to set %s config file permissions '
                           'in service host', conf_type)
            return False

        # notify choices service about list change
        list_change_util = {'choices': 'ivxv-choiceimp',
                            'voters': 'ivxv-voterimp'}.get(conf_type)
        if list_change_util:
            self.log.info('Notify service about %s list change', conf_type)
            cmd = [list_change_util, '-instance', self.service_id, target_path]
            proc = self.ssh(cmd)
            if proc.returncode:
                self.log.info('Failed to notify service about %s list change',
                              conf_type)
                return False

        return True

    def register_config_version(self, conf_type, conf_ver, service_state):
        """Register config version and service state in database."""
        db = IVXVManagerDb(for_update=True)
        self.log.info('Registering %s config version "%s" '
                      'in management database', conf_type, conf_ver)
        db.set_value(self.get_db_key('%s-conf-version' % conf_type), conf_ver)
        self.log.info('Registering service state as "%s" '
                      'in management database', service_state)
        db.set_value(self.get_db_key('state'), service_state)
        db.close()

    def get_service_certificate(self):
        """Get service certificate from service host."""
        self.log.info('Importing service certificate')

        cmd = 'cat {}'.format(os.path.join(self.service_data_dir, 'tls.pem'))
        proc = self.ssh(cmd, stdout=subprocess.PIPE, universal_newlines=True)
        if proc.returncode:
            self.log.info('Failed to import service certificate')
            return False

        return proc.stdout

    def get_collected_votes(self, output_archive_name):
        """Get collected votes from voting service host."""
        assert self.service_type == 'voting'

        # generate collected votes archive in voting service host
        self.log.info(
            'Generating collected votes archive in voting service host')

        votes_filename = '/tmp/exported-votes-%d' % os.getpid()
        cmd = 'ivxv-voteexp -instance {} {}'.format(self.service_id,
                                                    votes_filename)
        proc = self.ssh(cmd)
        if proc.returncode == 2:
            self.log.warning(
                'The voteexp application reports non-fatal errors')
        elif proc.returncode:
            self.log.error('Failed to generate collected votes archive '
                           'in voting service host')
            return False

        # copy collected votes archive file from voting service host to
        # management service host
        if not self.scp(local_path=output_archive_name,
                        remote_path=votes_filename,
                        description='collected votes archive file',
                        to_remote=False):
            return False

        # remove exported voters archive in voting service host
        # don't care about errors
        self.ssh('rm --force {}'.format(votes_filename))

        return True

    def ping(self):
        """Ping service."""
        self.log.debug('Pinging service')

        # query service status
        cmd = 'systemctl status {}'.format(
            'rsyslog' if self.service_type == 'log'
            else '--user {}'.format(self.service_systemctl_id))
        proc = self.ssh(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        if proc.returncode:
            log.error('Pinging service failed')
            return False

        log.debug('Service is alive')
        return True

    def scp(self, local_path, remote_path, description, account=None,
            to_remote=True):
        """
        Copy file using SCP protocol.

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

    def ssh(self, cmd, account=None, **kw):
        """Execute command in remote host."""
        if not isinstance(cmd, list):
            cmd = [cmd]
        ssh_cmd = ['ssh', '{}@{}'.format(account or self.service_account_name,
                                         self.hostname)]

        proc = exec_remote_cmd(ssh_cmd + cmd, **kw)

        return proc


def exec_remote_cmd(cmd, **kw):
    """
    Execute service management command in service host (over OpenSSH SSH
    client) or copy file to service host file system (over secure copy).

    :param cmd: Command for subprocess.run()
    :type cmd: list
    :param kw: Parameters for subprocess.run()
    :type kw: dict
    :returns: subprocess.CompletedProcess
    """
    assert cmd[0] in ('ssh', 'scp')

    # patch command
    if cmd[0] == 'ssh':
        cmd.insert(1, '-T')  # disable pseudo-terminal allocation
    cmd.insert(1, '-o')  # set preferred authentication method
    cmd.insert(2, 'PreferredAuthentications=publickey')

    # execute command
    log.debug('Executing command: %s', ' '.join(cmd))
    proc = subprocess.run(cmd, **kw)
    if proc.returncode:
        log.debug('Command finished with error code %d', proc.returncode)
    else:
        log.debug('Command successfully executed')

    return proc
