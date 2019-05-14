# IVXV Internet voting framework
"""CLI utilities for backup service."""

import datetime
import os
import random
import subprocess
import sys
import time

from jinja2 import Environment, PackageLoader

from . import init_cli_util, log
from .. import SERVICE_STATE_CONFIGURED
from ..config import CONFIG
from ..lib import get_services
from ..service.service import Service


def backup_crontab_generator_util():
    """Generate crontab for backup automation."""
    args = init_cli_util("""
    Generate crontab for IVXV backup automation.

    This utility must be called as editor by crontab utility:

        $ env VISUAL=ivxv-backup-crontab crontab -e

    Usage: ivxv-backup-crontab <filename>
    """)
    filepath = args['<filename>']

    # check input file
    try:
        with open(filepath) as fp:
            fp.read()
    except OSError as err:
        log.error('Can\'t read file "%s": %s', filepath, err.strerror)
        return 1

    # load crontab template
    tmpl_env = Environment(loader=PackageLoader('ivxv_admin', 'templates'))
    template = tmpl_env.get_template('ivxv_backup_crontab.jinja')

    # detect service states
    template_params = {'backup_times': []}
    backup_services = get_services(
        include_types=['backup'], service_state=[SERVICE_STATE_CONFIGURED])
    for service_val in backup_services.values():
        if service_val['backup-times']:
            template_params['backup_times'] = sorted(
                [[int(timeval.split(':')[0]), int(timeval.split(':')[1])]
                 for timeval in service_val['backup-times'].split(' ')])
    template_params['configured_backup_services'] = backup_services
    template_params['configured_voting_services'] = sorted(
        get_services(
            include_types=['voting'],
            service_state=[SERVICE_STATE_CONFIGURED]))
    template_params['configured_log_collectors'] = sorted(
        get_services(
            include_types=['log'],
            service_state=[SERVICE_STATE_CONFIGURED],
        ))

    # render crontab
    crontab = template.render(
        time_generated=datetime.datetime.now().strftime('%d.%M.%Y %H:%M:%S'),
        **template_params,
    )

    # Pause for 1 second. Crontab checks mtime to detect file modifications. It
    # seems that crontab can't detect mtime change if changes happens too
    # quickly (tested in Ubuntu Xenial).
    time.sleep(1)

    # write crontab
    try:
        with open(filepath, 'w') as fp:
            fp.write(crontab)
    except OSError as err:
        log.error('Can\'t write file "%s": %s', filepath, err.strerror)
        return 1

    return 0


def backup_util():
    """Backup collector data."""
    args = init_cli_util("""
    Backup IVXV collector data.

    Usage: ivxv-backup management-conf
           ivxv-backup ballot-box [<voting_service_id>]
           ivxv-backup log
    """)

    # execute ballot box and log backup command with ssh-agent wrapper
    if ((args['ballot-box'] or args['log'])
            and not os.environ.get('SSH_AUTH_SOCK')
            and not os.environ.get('SSH_AGENT_PID')):
        log.debug('Starting command with ssh-agent wrapper')
        os.execvp('ssh-agent', ['ssh-agent'] + sys.argv)

    services = get_services(include_types=['backup'])
    if not services:
        log.error('Backup service is not defined')
        return 1
    assert len(services) == 1

    backup_service = Service(*list(services.items())[0])
    log.debug('Backup service: %s', backup_service.service_id)
    if backup_service.data['state'] != SERVICE_STATE_CONFIGURED:
        log.error('Backup service state is "%s" (expected state is "%s")',
                  backup_service.data['state'], SERVICE_STATE_CONFIGURED)
        return 1

    if args['management-conf']:
        return _backup_management_cfg(backup_service)

    # copy list of known SSH hosts to backup server
    backup_service.scp(
        os.path.expanduser('~/.ssh/known_hosts'), '~/.ssh/',
        'list of known SSH hosts')
    backup_timestamp = datetime.datetime.now().strftime('%Y%m%d_%H%M')

    if args['ballot-box']:
        backup_target = datetime.datetime.now().strftime(
            f'ballot-box-{backup_timestamp}.zip')
        services = get_services(
            include_types=['voting'],
            service_state=[SERVICE_STATE_CONFIGURED],
        )
        voting_service_id = (
            args['<voting_service_id>'] or random.choice(list(services)))
        if args['<voting_service_id>'] and voting_service_id not in services:
            log.error('Unknown voting service ID: %s', voting_service_id)
            return 1

        service = Service(voting_service_id, services[voting_service_id])
        proc = backup_service.ssh(
            [
                'ivxv-admin-sudo',
                'backup-ballot-box',
                service.hostname,
                voting_service_id,
                backup_target,
            ],
            fwd_auth_agent=True,
        )

    else:
        assert args['log']
        services = get_services(
            include_types=['log'], service_state=[SERVICE_STATE_CONFIGURED])
        for log_collector_id in services:
            service = Service(log_collector_id, services[log_collector_id])
            proc = backup_service.ssh(
                [
                    'ivxv-admin-sudo',
                    'backup-log',
                    service.hostname,
                    backup_timestamp,
                ],
                fwd_auth_agent=True,
            )
            if proc.returncode:
                break

    if proc.returncode:
        log.error('Command execution failed with error code %d',
                  proc.returncode)
        return 1

    return 0


def _exec_backup_service_cmd(backup_service, *cmd):
    """Execute command in backup service.

    :raises OSError: if command fails
    """
    proc = backup_service.ssh(list(cmd))
    if proc.returncode:
        raise OSError('Command "{}" failed in backup service with exit code {}'
                      .format(' '.join(cmd), proc.returncode))


def _backup_management_cfg(backup_service):
    """Creating management config backup."""
    log.info('Creating management config backup')
    backup_target = datetime.datetime.now().strftime('%Y%m%d_%H%M')
    backup_basedir = '/var/backups/ivxv/management-conf'
    backup_tmpdir = os.path.join(backup_basedir, f'tmp-{backup_target}')
    backup_tgt_path = os.path.join(backup_basedir, backup_target)

    dirs_to_backup = [
        ['config', '/etc/ivxv', 'etc'],
        [
            'admin UI permissions', CONFIG['permissions_path'],
            'admin-ui-permissions'
        ],
        ['command history', CONFIG['command_files_path'], 'commands'],
    ]
    try:
        _exec_backup_service_cmd(backup_service, 'rm', '-rfv', backup_tmpdir)
        _exec_backup_service_cmd(backup_service, 'mkdir', '-v', backup_tmpdir)
        for description, src_dir, tgt_dir in dirs_to_backup:
            log.info('Backing up %s directory %s', description, src_dir)
            subprocess.run(
                [
                    'rsync',
                    '-av',
                    '--del',
                    src_dir + '/',
                    f'{backup_service.hostname}:{backup_tmpdir}/{tgt_dir}/',
                ],
                check=True,
            )
        _exec_backup_service_cmd(backup_service, 'rm', '-rfv', backup_tgt_path)
        _exec_backup_service_cmd(backup_service, 'mv', '-v', backup_tmpdir,
                                 backup_tgt_path)
    except (OSError, subprocess.CalledProcessError) as err:
        log.error(err)
        return 1

    return 0
