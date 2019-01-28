# IVXV Internet voting framework
"""
Library for collector management service.
"""

import atexit
import datetime
import fcntl
import json
import logging
import os
import shutil

from . import (CFG_TYPES, RFC3339_DATE_FORMAT_WO_FRACT, SERVICE_STATE_REMOVED,
               SERVICE_TYPE_PARAMS, VOTING_LIST_TYPES)
from .collector_state import generate_collector_state
from .config import CONFIG, cfg_path
from .db import DB_SERVICE_SUBKEYS, IVXVManagerDb
from .event_log import register_service_event
from .service import generate_service_list
from .service.service import Service

# create logger
log = logging.getLogger(__name__)


class IvxvError(Exception):
    """IVXV exception class."""
    pass


class PidLocker:
    """Pidfile creator and locker.

    Create pidfile and keep it locked until program exit.
    Raise IOError if

    >>> try:
    ...     PidLocker('/path/to/pidfile')
    ... except IOError:
    ...     print('Locking failed')
    ...     exit(1)
    """
    filepath = None  #: Pidfile name
    fp = None  #: Pidfile descriptor

    def __init__(self, pidfile_name):
        """Create and lock pidfile.

        :raises IOError: On locking failure
        """
        self.filepath = pidfile_name
        self.fp = open(pidfile_name, 'ab')
        fcntl.flock(self.fp, fcntl.LOCK_EX | fcntl.LOCK_NB)
        atexit.register(self.rm_pid)
        self.fp.write(bytes('{}\n'.format(os.getpid()), 'ASCII'))
        self.fp.flush()

    def rm_pid(self):
        """Remove pidfile."""
        try:
            os.remove(self.filepath)
        except FileNotFoundError:
            log.warning('PID file %s already removed', self.filepath)

    @classmethod
    def get_pidfile_path(cls, pidfile_name):
        """Get pidfile path."""
        pidfile_path = cfg_path('ivxv_admin_data_path', pidfile_name)
        return pidfile_path

    @classmethod
    def pidfile_exists(cls, pidfile_name):
        """Check if pidfile exist."""
        pidfile_path = cls.get_pidfile_path(pidfile_name)
        return os.path.exists(pidfile_path)

    @classmethod
    def rm_stale_pidfile(cls, pidfile_name):
        """Remove stale or invalid pidfile."""
        if not cls.pidfile_exists(pidfile_name):
            return

        pidfile_path = cls.get_pidfile_path(pidfile_name)
        with open(pidfile_path) as fp:
            pidline = fp.readline()

        reason = '' if pidline.strip() else 'empty'
        if not reason:
            try:
                pid = int(pidline)
            except ValueError:
                reason = 'invalid'
        if not reason and not os.path.exists(f'/proc/{pid}'):
            reason = 'stale'
        if reason:
            log.error('Removing %s pidfile %s', reason, pidfile_path)
            os.remove(pidfile_path)


def register_tech_cfg_items(cfg, cfg_version):
    """Register technical config items in management database.

    Create database records for service hosts and services if not created and
    register removed services.

    :param cfg: Technical config
    :param cfg_version: Config version
    """
    with IVXVManagerDb(for_update=True) as db:
        register_removed_services(db, cfg, cfg_version)
        gen_service_record_defaults(db, cfg)
        gen_host_record_defaults(db, cfg)
        gen_logmon_data(db, cfg.get('logging'))


def register_removed_services(db, cfg, cfg_version):
    """Mark removed services in database with REMOVED state.

    Change state of removed services to REMOVED.

    :param db: Database handler
    :param cfg: Technical config
    """
    existing_services = get_services(db=db)
    services_from_cfg = generate_service_list(cfg['network'])

    # find removed services
    existing_service_ids = set(existing_services.keys())
    new_service_ids = set(
        new_service_cfg[1]['id']
        for new_service_cfg in services_from_cfg)
    removed_service_ids = existing_service_ids.difference(new_service_ids)
    if not removed_service_ids:
        return

    # change loading states of choices and voters lists
    # if all existing storage services are removed
    storages_removed = set(
        service_id
        for service_id, service_data in existing_services.items()
        if service_data['service-type'] == 'storage' and
        service_data['state'] != SERVICE_STATE_REMOVED)
    storages_added = set(
        new_service_cfg['id']
        for service_type, new_service_cfg in services_from_cfg
        if service_type == 'storage')
    if not storages_removed.intersection(storages_added):
        # all storages removed
        log_msg = (
            f'All existing storage service instances are removed with '
            f'technical config {cfg_version}')
        log.warning(log_msg)
        _reset_list_loading_state(db, 'choices', 'choices list', log_msg)
        for list_no in range(detect_voters_list_order_no(db), 0, -1):
            _reset_list_loading_state(
                db, f'voters{list_no:02d}', f'voterst list #{list_no:02d}',
                log_msg)

    # change removed service states
    db_values = db.get_all_values()['service']
    for service_id in sorted(removed_service_ids):
        log.info('Changing service %s state to REMOVED', service_id)
        service = Service(service_id, db_values[service_id])
        service.register_state(
            db,
            SERVICE_STATE_REMOVED,
            bg_info=f'Service removed with technical config: {cfg_version}',
        )


def _reset_list_loading_state(db, list_type, list_descr, log_msg):
    """Reset choices and voters list loading state."""
    # check if list is loaded
    list_version = db.get_value(f'list/{list_type}')
    if not list_version:
        return

    log_msgs = [log_msg]
    timestamp = datetime.datetime.now().strftime(RFC3339_DATE_FORMAT_WO_FRACT)

    # load list metadata file
    filepath = get_loaded_cfg_file_path(list_type)
    assert filepath, (
        f'Config file for {list_descr} "{list_version}" does not exist')
    filepath = filepath.replace('.bdoc', '.json')
    with open(filepath) as fp:
        choices_cfg_state = json.load(fp)

    # set list state to PENDING in management database
    db_key = f'list/{list_type}-loaded'
    if db.get_value(db_key):
        log_msg = f'Setting {list_descr} state to PENDING'
        log.info(log_msg)
        log_msgs.append(log_msg)
        db.set_value(db_key, '')

    # write config metadata file
    log_msgs = [f'{timestamp} {log_msg}' for log_msg in log_msgs]
    choices_cfg_state['log'] += [log_msgs]
    choices_cfg_state['attempts'] = 0
    choices_cfg_state['completed'] = False
    tmp_filepath = f'{filepath}.tmp'
    with open(tmp_filepath, 'w') as fp:
        json.dump(choices_cfg_state, fp)
    shutil.move(tmp_filepath, filepath)


def gen_service_record_defaults(db, cfg):
    """Add database records for services with default values.

    Parse the list of services from technical config and create
    missing database records for every defined service.

    :param db: Database handler
    :param cfg: Technical config
    """
    assert 'network' in cfg

    # generate list of default values
    service_values = {}
    for network in cfg['network']:
        for service_type, services in sorted(network['services'].items()):
            for service in services or []:
                service_id = service['id']
                service_values[service_id] = DB_SERVICE_SUBKEYS.copy()
                service_values[service_id].update({
                    'service-type': service_type,
                    'network': network['id'],
                    'ip-address': service['address'],
                })

    # set service default values
    for service_id, service_defaults in sorted(service_values.items()):
        db_key_prefix = f'service/{service_id}'
        try:
            db.get_value(db_key_prefix + '/service-type')
            continue
        except KeyError:
            log.info('Registering new service %s in management service',
                     service_id)
            register_service_event(
                'SERVICE_REGISTER',
                service=service_id,
                params={'service_type': service_defaults['service-type']})

        for key, val in service_defaults.items():
            db.set_value(db_key_prefix + '/' + key, val)

    _set_tech_cfg_service_cond_values(db, cfg)


def _set_tech_cfg_service_cond_values(db, cfg):
    """Set/remove service conditional values related to technical config."""
    for network in cfg['network']:
        for service_type, services in sorted(network['services'].items()):
            for service in services or []:
                # create 'tls-key' and 'tls-cert' for
                # choices/dds/storage/verification/voting services
                if SERVICE_TYPE_PARAMS[service_type]['require_tls']:
                    _manage_db_cond_value(db, service['id'], 'tls-key', True)
                    _manage_db_cond_value(db, service['id'], 'tls-cert', True)
                # create 'backup-times' for backup service
                elif service_type == 'backup':
                    _manage_db_cond_value(
                        db, service['id'], 'backup-times',
                        True, ' '.join(cfg.get('backup') or []))


def _manage_db_cond_value(db, service_id, key, set_value, value=None):
    """Manage (create/remove) service conditional database value."""
    key = f'service/{service_id}/{key}'
    if set_value:  # create key
        try:
            db.get_value(key)
        except KeyError:
            db.set_value(key, value or '')
    else:  # remove key
        try:
            db.rm_value(key)
        except KeyError:
            pass


def manage_db_dds_fields(db):
    """Create/remove service DDS token keys in database.

    Check 'ticket' authentication method in election config and manage
    "dds-token-key" keys in management database for dds, choices and voting
    services. Keys will created or removed as 'ticket' authentication method is
    used or not.
    """
    ticket_auth = 'ticket' in db.get_all_values('election').get('auth', {})
    for service_id, service_data in db.get_all_values('service').items():
        if service_data['service-type'] not in ['dds', 'choices', 'voting']:
            continue
        key = f'service/{service_id}/dds-token-key'
        if ticket_auth:
            if 'dds-token-key' not in service_data:
                db.set_value(key, '')
        elif 'dds-token-key' in service_data:
            db.rm_value(key)


def manage_db_tsp_fields(db):
    """Create/remove service TSP registration keys in database.

    Check 'qualification' protocol in election config and manage "tspreg-key"
    keys in management database for voting services. Keys will created or
    removed as 'tspreg' protocol is used or not.
    """
    tspreg = db.get_all_values('election').get('tspreg', '')
    for service_id, service_data in db.get_all_values('service').items():
        if service_data['service-type'] not in ['voting']:
            continue
        key = f'service/{service_id}/tspreg-key'
        if tspreg:
            if 'tspreg-key' not in service_data:
                db.set_value(key, '')
        elif 'tspreg-key' in service_data:
            db.rm_value(key)


def gen_host_record_defaults(db, cfg):
    """Add database records for service hosts with default values.

    Parse the list of services from technical config and create
    missing database records for service hosts.

    :param db: Database handler
    :param cfg: Technical config
    """
    # generate list of default values
    hostnames = set()
    for network in cfg['network']:
        for services in network['services'].values():
            for service in services or []:
                hostnames |= {service['address'].split(':')[0]}

    # set host default values
    for hostname in sorted(hostnames):
        db_key_prefix = f'host/{hostname}'
        try:
            db.get_value(db_key_prefix + '/state')
            continue
        except KeyError:
            log.info(
                'Registering new service host %s in management service',
                hostname)
        db.set_value(db_key_prefix + '/state', '')


def gen_logmon_data(db, logging_params):
    """Add database record for Log Monitor service.

    :param logging_params: Logging params from technical config.
    :type logging_params: list of dicts
    """
    if logging_params:
        db.set_value('logmonitor/address', logging_params[0]['address'])


def cfg_type_verbose(cfg_type):
    """Get config type as human readable string."""
    try:
        return CFG_TYPES[cfg_type]
    except KeyError:
        return VOTING_LIST_TYPES[cfg_type]


def get_services(db=None,
                 require_collector_state=None,
                 service_state=None,
                 include_types=None,
                 exclude_types=None):
    """Generate filtered list of services.

    :param require_collector_state: Require specified state for collector
    :type require_collector_state: list
    :param include_types: Include specified service types
    :type include_types: list
    :param exclude_types: Exclude specified service types
    :type exclude_types: list
    :return: dict of services (key is service ID and value
             is dict with service database values) or
             None if collector is not in expected state
    """
    require_collector_state = require_collector_state or []
    assert isinstance(require_collector_state, list)
    service_state = service_state or []
    assert isinstance(service_state, list)
    include_types = include_types or []
    exclude_types = exclude_types or []
    assert isinstance(include_types, list)
    assert isinstance(exclude_types, list)
    assert not include_types or not exclude_types

    # get collector state
    if db:
        collector_state = generate_collector_state(db)
    else:
        with IVXVManagerDb() as db:
            collector_state = generate_collector_state(db)

    # check collector state
    if (require_collector_state and
            collector_state['collector_state'] not in require_collector_state):
        log.info('Collector service state is %s',
                 collector_state['collector_state'])
        if len(require_collector_state) == 1:
            log.error('Collector state must be %s for this operation',
                      require_collector_state[0])
        else:
            log.error('Collector state must be %s or %s for this operation',
                      ', '.join(require_collector_state[:-1]),
                      require_collector_state[-1])
        return None

    # create list of services
    services = {}
    for service_id, service_data in collector_state.get('service', {}).items():
        if service_state and service_data.get('state') not in service_state:
            continue
        if include_types:
            if service_data['service-type'] in include_types:
                services[service_id] = service_data
            continue
        if exclude_types:
            if service_data['service-type'] not in exclude_types:
                services[service_id] = service_data
            continue
        services[service_id] = service_data

    return services


def clean_dir(path):
    """Remove all files from directory."""
    log.debug('Cleaning directory %s', path)
    for filename in os.listdir(path):
        role_filename = os.path.join(path, filename)
        os.unlink(role_filename)


def populate_user_permissions(db):
    """Populate user permissions for Apache web server."""
    permissions_path = CONFIG['permissions_path']
    permissions_to_create = []

    for user_cn, permissions in db.get_all_values('user').items():
        for permission_name in permissions.split(','):
            permissions_to_create.append(user_cn + '-' + permission_name)

    # removing permission files
    for permission_name in os.listdir(permissions_path):
        if permission_name not in permissions_to_create:
            filepath = os.path.join(permissions_path, permission_name)
            log.info('Removing Apache Web Server user permission file %s',
                     filepath)
            os.remove(filepath)

    # creating permission files
    for permission_name in permissions_to_create:
        filepath = os.path.join(permissions_path, permission_name)
        if not os.path.exists(filepath):
            log.info('Creating Apache Web Server user permission file %s',
                     filepath)
            with open(filepath, 'x') as fp:
                fp.write('Created %s' % datetime.datetime.now().strftime('%c'))


def detect_voters_list_order_no(db):
    """Detect next voters list order number from database.

    :rtype: int
    """
    for voter_list_no in range(1, 100):
        try:
            db.get_value(f'list/voters{voter_list_no:02d}')
        except KeyError:
            return voter_list_no - 1
    return 0


def get_loaded_cfg_file_path(cfg_type):
    """Get real file path for loaded config file.

    :return: File path or None if file does not exist
    :rtype: str
    """
    filepath = cfg_path('active_config_files_path', f'{cfg_type}.bdoc')
    if not os.path.exists(filepath):
        return None

    assert os.path.islink(filepath), (
        f'Config file {filepath} is not a symbolic link')

    return os.path.realpath(filepath)
