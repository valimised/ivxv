# IVXV Internet voting framework
"""CLI utilities for service management."""

import datetime
import os
import subprocess
import sys

from . import init_cli_util, log
from .. import (COLLECTOR_STATE_CONFIGURED, COLLECTOR_STATE_FAILURE,
                COLLECTOR_STATE_INSTALLED, COLLECTOR_STATE_NOT_INSTALLED,
                COLLECTOR_STATE_PARTIAL_FAILURE, SERVICE_STATE_CONFIGURED,
                SERVICE_STATE_FAILURE, SERVICE_STATE_INSTALLED,
                SERVICE_STATE_REMOVED, SERVICE_TYPE_PARAMS, __version__)
from ..agent_daemon import get_collector_data, ping_service
from ..db import IVXVManagerDb
from ..lib import IvxvError, get_services
from ..service.service import Service, exec_remote_cmd

#: JSON formatting options
JSON_DUMP_ARGS = dict(indent=2, sort_keys=True)


def consolidate_votes_util():
    """Consolidate and export collected votes."""
    args = init_cli_util("""
    Export collected votes.

    This utility copies current ballot box from voting service to backup
    service and consolidates collected votes from all ballot box backup files.

    Usage: ivxv-export-votes [--consolidate] <output-file>

    Options:
        --consolidate   Consolidate all ballot box backup files
    """)

    try:
        consolidate_votes(args['--consolidate'], args['<output-file>'])
    except IvxvError as err:
        log.error(err)
        return 1

    return 0


def consolidate_votes(must_consolidate, output_filename):
    """Consolidate votes."""
    # fail if output file already exist
    if os.path.exists(output_filename):
        raise IvxvError(f'Output file {output_filename} already exist')

    # create backup of current ballot box
    log.info('Creating backup copy from current ballot box')
    try:
        subprocess.run(['ivxv-backup', 'ballot-box'], check=True)
    except OSError as err:
        raise IvxvError(
            'Creating ballot box backup failed: {}'.format(err.strerror))
    except subprocess.CalledProcessError as err:
        raise IvxvError('Creating ballot box backup failed: {}'.format(err))

    # create handler for backup service
    services = get_services(include_types=['backup'])
    if not services:
        raise IvxvError('Backup service is not defined')
    assert len(services) == 1

    backup_service = Service(*list(services.items())[0])
    log.debug('Backup service: %s', backup_service.service_id)
    ballot_box_filepath = datetime.datetime.now().strftime(
        '/var/lib/ivxv/ballot-box-consolidated-%Y%m%d_%H%M.zip')

    # run consolidation in backup service
    if must_consolidate:
        proc = backup_service.ssh([
            'ivxv-voteunion', ballot_box_filepath,
            '/var/backups/ivxv/ballot-box/ballot-box-????????_????.zip'
        ])
        if proc.returncode:
            raise IvxvError('Consolidation command failed in backup service')
    else:
        # export without consolidation - detect last backup filename
        cmd = [
            'ls', '/var/backups/ivxv/ballot-box/ballot-box-????????_????.zip'
        ]
        proc = backup_service.ssh(cmd, stdout=subprocess.PIPE, check=True)
        ballot_box_filepath = proc.stdout.decode('UTF-8').strip().split(
            '\n')[-1]

    # copy consolidated ballot box from backup
    log.info('Copying ballot box to management service')
    if not backup_service.scp(
            output_filename,
            ballot_box_filepath,
            'consolidated ballot box',
            to_remote=False):
        raise IvxvError('Failed to copy ballot box to management service')
    if must_consolidate:
        log.info('Removing consolidated ballot box from backup service')
        backup_service.ssh(['rm', '-v', ballot_box_filepath])

    log.info('Collected votes archive is written to %s', output_filename)


def copy_logs_to_logmon_util():
    """Initialize Log Monitor."""
    # validate CLI arguments
    args = init_cli_util("""
    Copy IVXV log files from service hosts to Log Monitor.

    This utility transports collected IVXV log files from IVXV services
    (including Log Collector Service) to Log Monitor.

    Usage: ivxv-copy-log-to-logmon [<hostname> ...]

    Options:
        <hostname>  Service host name.
    """)

    # check collector state
    with IVXVManagerDb() as db:
        collector_state = db.get_value('collector/state')
        logmonitor_address = db.get_value('logmonitor/address')
        services = db.get_all_values('service')
    if collector_state == COLLECTOR_STATE_NOT_INSTALLED:
        # emit warning only if attached to console
        # (suppress if executed from crontab)
        if sys.stdout.isatty():
            log.warning('Collector is not installed')
            return 1
        return 0

    # create service host list
    hostnames = args['<hostname>']
    if not hostnames:
        hostnames = []
        for service in services.values():
            is_main_service = (
                SERVICE_TYPE_PARAMS[service['service-type']]['main_service'])
            if (is_main_service
                    or service['service-type'] == 'log'
                    and service['state'] != SERVICE_STATE_REMOVED):
                hostnames.append(service['ip-address'].split(':')[0])
        hostnames = list(set(hostnames))

    # checking access to Log Monitor account
    if not logmonitor_address:
        log.error('Log monitor is not defined')
        return 1
    log.info('Using address "%s" for log monitor', logmonitor_address)
    logmon_account = f'logmon@{logmonitor_address}'
    log.info('Checking SSH access to Log Monitor account %s', logmon_account)
    proc = exec_remote_cmd(['ssh', logmon_account, 'true'])
    if proc.returncode:
        log.error('Cannot access to Log Monitor (%s)', logmon_account)
        return 1

    # copy log file from service hosts to Log Monitor
    for hostname in hostnames:
        remote_account = f'ivxv-admin@{hostname}'

        # check if service host have Log Monitor host key
        log.info('Checking if "%s" have Log Monitor host key', remote_account)
        proc = exec_remote_cmd(
            ['ssh', hostname, 'ssh-keygen', '-F', logmonitor_address],
            stdout=subprocess.PIPE)
        if proc.returncode:
            log.error(
                'Failed to check Log Monitor host key for %s', remote_account)
            log.info('Installing Log Monitor host key for %s', remote_account)
            proc = subprocess.run(
                ['sh', '-c',
                 f'ssh-keygen -F {logmonitor_address} | '
                 'grep -v ^# | '
                 f'ssh {hostname} tee --append .ssh/known_hosts'])
            if proc.returncode:
                log.error('Failed to install Log Monitor host key for %s',
                          remote_account)
                continue

        # execute rsync command in host to copy log file to Log Monitor
        log.info(
            'Copying IVXV service log files from host "%s" to Log Monitor',
            hostname)
        transfer_cmd = [
            'ivxv-admin-helper', 'copy-logs-to-logmon', hostname,
            logmon_account
        ]
        cmd = ['ssh-agent', 'ssh', '-A', hostname] + transfer_cmd
        proc = subprocess.run(cmd)
        if proc.returncode:
            log.error('Failed to copy log file from host "%s" to Log Monitor',
                      hostname)

    return 0


def update_software_pkg_util():
    """Update service packages in service hosts."""
    # validate CLI arguments
    args = init_cli_util("""
    Update service packages in IVXV service hosts.

    This utility checks versions of software packages in service hosts
    and installs new versions if required.

    Usage: ivxv-update-packages [--force]

    Options:
        --force     Update even package version does not require update
    """)

    # generate list of voting services that are in required state
    services = get_services(
        require_collector_state=[
            COLLECTOR_STATE_INSTALLED,
            COLLECTOR_STATE_CONFIGURED,
            COLLECTOR_STATE_FAILURE,
            COLLECTOR_STATE_PARTIAL_FAILURE,
        ],
        service_state=[
            SERVICE_STATE_INSTALLED,
            SERVICE_STATE_CONFIGURED,
            SERVICE_STATE_FAILURE,
        ])
    if not services:
        return 1

    host_versions = dict(
        [services[service_id]['ip-address'].split(':')[0], '']
        for service_id in sorted(services))
    update_res = dict(failure=[], install=[], skip=[])

    def get_installed_pkg_ver():
        """Get version string if installed package in service host."""
        service.log.info('Detect %s package version', pkg_name)
        proc = service.ssh(
            f'dpkg --status {pkg_name} | grep ^Version: | cut -d: -f2',
            stdout=subprocess.PIPE,
            account='ivxv-admin')
        return proc.stdout.decode('UTF-8').strip()

    for service_id, service_data in sorted(services.items()):
        service = Service(service_id, service_data)

        # check ivxv-common version
        host_version = host_versions.get(service.hostname)
        pkg_name = 'ivxv-common'
        install_pkg = bool(args['--force'])
        if not install_pkg and host_version != __version__:
            host_versions[service.hostname] = get_installed_pkg_ver()
            install_pkg = host_versions[service.hostname] != __version__

        # install ivxv-common if required
        if install_pkg:
            if not service.update_ivxv_common_pkg():
                update_res['failure'].append([service_id, pkg_name])
                continue
            update_res['install'].append([service_id, pkg_name])
            host_versions[service.hostname] = __version__
        else:
            update_res['skip'].append([service_id, pkg_name])

        # check service package version and upgrade if required
        pkg_name = service.deb_pkg_name
        install_pkg = args['--force'] or get_installed_pkg_ver() != __version__

        # install package if required
        if install_pkg:
            if not service.install_service_pkg(is_update=True):
                update_res['failure'].append([service_id, pkg_name])
                continue
            update_res['install'].append([service_id, pkg_name])
        else:
            update_res['skip'].append([service_id, pkg_name])

    # output result
    for service_id, pkg_name in update_res['install']:
        log.info('Successfully installed service %s package %s',
                 service_id, pkg_name)
    for service_id, pkg_name in update_res['failure']:
        log.error('Failed to install service %s package %s',
                  service_id, pkg_name)
    log.info('Service update stats: %d packages installed, '
             '%d package installations failed, %s packages skipped',
             len(update_res['install']),
             len(update_res['failure']),
             len(update_res['skip']))

    return 1 if update_res['failure'] else None


def manage_service():
    """Manage IVXV services."""
    # validate CLI arguments
    args = init_cli_util("""
    Manage IVXV services.

    Usage: ivxv-service <action> <service-id> ...

    Options:
        <action>    Management action: start, stop, restart, ping
    """)
    action = args['<action>']
    if action not in ['start', 'stop', 'restart', 'ping']:
        log.error('Invalid action: %s', action)
        return 1

    services = get_collector_data()
    if services is None:
        log.error('Election data is not loaded')
        return 1

    exit_code = 0
    for service_id in args['<service-id>']:
        if service_id not in services:
            log.error('Unknown service %s', service_id)
            exit_code = 1
            continue
        if action == 'ping':
            log.info('Pinging service %s', service_id)
            if ping_service(service_id, services[service_id]):
                log.info('Service %s is alive', service_id)
            else:
                log.error('Failed to query service %s status', service_id)
                exit_code = 1
            continue

        service = Service(service_id, services[service_id])

        if action == 'stop':
            log.info('Stopping service %s', service_id)
            if service.stop_service():
                log.info('Service %s stopped', service_id)
            else:
                log.error('Failed to stop service %s', service_id)
                exit_code = 1
            continue

        # start, restart
        log.info('%sing service %s', action.capitalize(), service_id)
        service = Service(service_id, services[service_id])
        if service.restart_service():
            log.info('Service %s %sed', service_id, action)
        else:
            log.error('Failed to %s service %s', action, service_id)
            exit_code = 1

    return exit_code
