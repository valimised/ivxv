# IVXV Internet voting framework
"""CLI utilities for command loading."""

import json
import os
import re
import shutil
import subprocess

from .. import init_cli_util, log
from ... import CFG_TYPES, CMD_DESCR, CMD_TYPES, VOTING_LIST_TYPES
from ...command_file import (check_cmd_signature, load_cfg_file_content,
                             load_collector_cmd_file)
from ...config import cfg_path
from ...db import IVXVManagerDb
from ...event_log import register_service_event
from ...lib import (IvxvError, detect_voters_list_order_no,
                    manage_db_dds_fields, manage_db_tsp_fields,
                    populate_user_permissions, register_tech_cfg_items)
from ...service.backup_service import remove_backup_crontab


def main():
    """Load command to IVXV Collector Management Service."""
    args = init_cli_util("""
    Load command to IVXV Collector Management Service.

    Usage: ivxv-cmd-load [--autoapply] [--show-version] <type> FILE

    Options:
        <type>              Command type. Possible values are:
                            - election: election config
                            - technical: collector technical config
                            - trust: trust root config
                            - choices: choices list
                            - districts: districts list
                            - voters: voters list
                            - user: user account and role(s)
        --autoapply         Apply command file automatically (by Agent Daemon).
        --show-version      Output config file version and exit.
    """)

    # validate CLI arguments
    cmd_type = args['<type>'].lower()
    if cmd_type not in CMD_TYPES:
        log.error('Invalid command type "%s". Possible values are: %s',
                  cmd_type, ', '.join(CMD_TYPES))
        return 1
    cfg_filename = args['FILE']

    # check command file signature and loading state
    try:
        check_cmd_loading_state(cmd_type)
        cfg_timestamp, cfg_version = check_signer_permissions(
            cmd_type, filename=args['FILE'])
    except IvxvError as err:
        log.error(str(err))
        return 1

    # output config version
    log.info('Config file version is "%s"', cfg_version)
    if args['--show-version']:
        return 0

    # raise error if reloading current version
    if cmd_type in CFG_TYPES:
        with IVXVManagerDb() as db:
            if db.get_value(f'config/{cmd_type}') == cfg_version:
                log.error('%s version "%s" is already loaded',
                          CFG_TYPES[cmd_type].capitalize(), cfg_version)
                return 1

    # load config (includes config validation)
    log.info('Loading command "%s" from file %s',
             CMD_DESCR[cmd_type], cfg_filename)
    cfg_data = load_collector_cmd_file(cmd_type, args['FILE'])
    if cfg_data is None:
        return 1

    # validate voting lists consistency
    if cmd_type in VOTING_LIST_TYPES:
        if not _validate_lists_consistency(cmd_type, args['FILE']):
            return 1

    register_service_event(
        'CMD_LOAD', params={'cmd_type': cmd_type, 'version': cfg_version})

    # reset database and remove crontab on trust root config loading
    if cmd_type == 'trust':
        log.info('Resetting collector management database')
        db = IVXVManagerDb()
        db.reset()

        remove_backup_crontab()
        register_service_event('COLLECTOR_RESET')

    # register new and removed services on technical config loading
    elif cmd_type == 'technical':
        register_tech_cfg_items(cfg_data, cfg_version)

    # write districts JSON to web server data directory
    elif cmd_type == 'districts':
        districts_filename = cfg_path('admin_ui_data_path', 'districts.json')
        districts_list = sorted(
            [[dist_id, f'{dist_id} - {val["name"]}']
             for dist_id, val in cfg_data['districts'].items()])
        log.info('Writing simplified district list to %s', districts_filename)
        with open(districts_filename, 'w') as fp:
            json.dump(districts_list, fp)

    # register loaded config
    register_cfg(
        cmd_type, cfg_data, cfg_filename, cfg_timestamp, cfg_version,
        args['--autoapply'])
    register_service_event(
        'CMD_LOADED', params={'cmd_type': cmd_type, 'version': cfg_version})

    return 0


def register_cfg(cmd_type,
                 cfg_data,
                 cfg_filename,
                 cfg_timestamp,
                 cfg_version,
                 autoapply):
    """Register config version in database and file system."""
    db_key = 'config' if cmd_type in CFG_TYPES else 'list'

    # connect to management service database
    with IVXVManagerDb(for_update=True) as db:
        # detect order number for voters list
        if cmd_type == 'voters':
            voter_list_no = detect_voters_list_order_no(db) + 1
            active_cfg_filename = f'{cmd_type}{voter_list_no:02}.bdoc'
            db_key += f'/{cmd_type}{voter_list_no:02}'
        else:
            active_cfg_filename = f'{cmd_type}.bdoc'
            db_key += '/' + cmd_type

        # write config file to admin ui data path
        cmd_filepath = cfg_path(
            'command_files_path', f'{cmd_type}-{cfg_timestamp}.bdoc')
        log.debug('Copying file %s to %s', cfg_filename, cmd_filepath)
        shutil.copy(cfg_filename, cmd_filepath)
        log.info('%s file loaded successfully',
                 CMD_DESCR[cmd_type].capitalize())

        # register config file version
        if cmd_type in CFG_TYPES or cmd_type in VOTING_LIST_TYPES:
            db.set_value(db_key, cfg_version)
            if cmd_type == 'voters':
                db.set_value(db_key + '-loaded', '')

        # register user permissions
        if cmd_type == 'trust':  # initial permissions
            log.info('Resetting user permissions')
            for user_cn in db.get_all_values('user'):
                db.rm_value('user/' + user_cn)
            for user_cn in cfg_data['authorizations']:
                _register_user_permissions(db, user_cn, ['admin'])

        elif cmd_type == 'user':  # permissions from user management command
            user_cn = cfg_data['cn']
            log.info('Resetting user "%s" permissions', cfg_data['cn'])
            register_service_event(
                'PERMISSION_RESET', params={'user_cn': user_cn})
            for existing_user_cn in db.get_all_values('user'):
                if existing_user_cn == user_cn:
                    db.rm_value('user/' + user_cn)
            _register_user_permissions(db, user_cn, cfg_data['roles'])

        elif cmd_type == 'election':  # register election params
            cfg_data = load_cfg_file_content(
                cmd_type,
                re.compile(r'(.+\.)?{}.yaml'.format(cmd_type)), cfg_filename)
            db.set_value('election/election-id', cfg_data['identifier'])

            # start/stop times
            for key in [
                    'servicestart', 'electionstart', 'electionstop',
                    'servicestop'
            ]:
                db.set_value('election/' + key, cfg_data['period'][key])
                register_service_event(
                    'SET_ELECTION_TIME',
                    params={
                        'period': key, 'timestamp': cfg_data['period'][key]})

            # authentication methods
            auth_in_db = set(db.get_all_values('election').get('auth', []))
            auth_in_cfg = set(cfg_data.get('auth', []).keys())
            for key in auth_in_db.difference(auth_in_cfg):
                db.rm_value(f'election/auth/{key}')
            for key in auth_in_cfg.difference(auth_in_db):
                db.set_value(f'election/auth/{key}', 'TRUE')

            # TSP qualification protocol
            for key in cfg_data.get('qualification', []):
                if key.get('protocol') == 'tspreg':
                    db.set_value('election/tsp-qualification', 'TRUE')
                    break
            else:
                try:
                    db.rm_value('election/tsp-qualification')
                except KeyError:
                    pass

        if cmd_type in ['technical', 'election']:
            manage_db_dds_fields(db)
            manage_db_tsp_fields(db)

    register_cfg_in_fs(
        cmd_type, cmd_filepath, active_cfg_filename, cfg_version, autoapply)


def register_cfg_in_fs(cmd_type,
                       src_path,
                       tgt_filename,
                       cfg_version,
                       autoapply):
    """Register config in file system.

    - Create symlink to active config directory
      (e.g. /var/lib/ivxv/commands/<cfg-version>.bdoc -> /etc/ivxv/cmd.bdoc)
    - Create state file for some command types
      (/var/lib/ivxv/commands/<cfg-version>.json)
    """
    if cmd_type not in CFG_TYPES and cmd_type not in VOTING_LIST_TYPES:
        return

    tgt_path = cfg_path('active_config_files_path', tgt_filename)
    state_path = os.path.splitext(src_path)[0] + '.json'

    # create symlink to active config directory
    try:
        os.remove(tgt_path)
    except FileNotFoundError:
        pass
    log.debug('Creating symlink for config file %s -> %s', src_path, tgt_path)
    os.symlink(src_path, tgt_path)

    # create state file for loaded command
    if cmd_type in ['technical', 'election', 'choices', 'voters']:
        try:
            os.remove(state_path)
        except FileNotFoundError:
            pass
        log.debug('Generating config state file %s', state_path)
        default_state_data = {
            # config data
            'config_type': CMD_DESCR[cmd_type],
            'config_file': os.path.basename(tgt_path),
            'config_version': cfg_version,
            # applying data
            'autoapply': autoapply,
            'completed': False,
            'attempts': 0,
            'log': [],
        }
        with open(state_path, 'w') as fp:
            json.dump(default_state_data, fp, indent=4, sort_keys=True)

    log.info('%s file is registered in management service',
             CMD_DESCR[cmd_type].capitalize())


def check_cmd_loading_state(cmd_type):
    """Check command loading state.

    :raises IvxvError: on any known error
    """
    # don't allow to load choices list more than once
    if cmd_type == 'choices':
        with IVXVManagerDb() as db:
            choices_list_version = db.get_value('list/choices')
        if choices_list_version:
            raise IvxvError(
                'Choices list is already loaded (version: {})'
                .format(choices_list_version))

    # don't allow to load technical config if trust root config is not loaded
    elif cmd_type == 'technical':
        with IVXVManagerDb() as db:
            trust_cfg = db.get_value('config/trust')
        if not trust_cfg:
            raise IvxvError(
                'Trust root must be loaded before technical configuration')


def check_signer_permissions(cmd_type, filename):
    """Check command file signer permissions.

    Detect config file timestamp and version based on signature(s).

    :return: config timestamp, config version
    :rtype: list

    :raises IvxvError: on any known error
    """
    # check config permissions
    try:
        cfg_signatures, all_signatures = check_cmd_signature(
            cmd_type, filename)
    except LookupError as err:
        raise IvxvError(f'Failed to verify config file signatures: {err}')
    except (OSError, subprocess.SubprocessError) as err:
        raise IvxvError(str(err))

    for signature in all_signatures:
        log.info('Config file is signed by: %s', signature[2])
    if not cfg_signatures:
        raise IvxvError('No signatures by authorized users')
    cfg_version, role = cfg_signatures[0]
    log.info('User %s with role "%s" is authorized to execute "%s" '
             'commands', cfg_version.split(' ')[0], role, cmd_type)
    log.info('Using signature "%s" as config file version', cfg_version)
    cfg_timestamp = cfg_version.split(' ')[1]

    return cfg_timestamp, cfg_version


def _register_user_permissions(db, user_cn, roles):
    """Register user permissions."""
    roles = sorted(roles)
    register_service_event(
        'PERMISSION_SET',
        params={'user_cn': user_cn, 'permission': ','.join(roles)})
    db.set_value('user/' + user_cn, ','.join(roles))
    populate_user_permissions(db)


def _validate_lists_consistency(cmd_type, cmd_filepath):
    """Validate voting lists consistency."""
    # create list of existing voting lists
    list_files = {}
    with IVXVManagerDb() as db:
        for db_key, db_val in db.get_all_values('list').items():
            if '-loaded' in db_key or not db_val:
                continue
            list_files[db_key] = cfg_path(
                'active_config_files_path', f'{db_key}.bdoc')

    # add current list
    list_files[cmd_type] = cmd_filepath

    # validate
    if len(list_files) > 1:
        cmd = ['ivxv-config-validate']
        for cfg_type, filepath in sorted(list_files.items()):
            cfg_type = re.sub(r'[0-9]+', '', cfg_type)
            cmd.append(f'--{cfg_type}={filepath}')
        log.info(
            'Some voting lists are already loaded, '
            'executing consistency checks: %s', ' '.join(cmd))
        proc = subprocess.run(cmd)
        if proc.returncode != 0:
            log.error('Config validation command raised exception')
            return False

    return True
