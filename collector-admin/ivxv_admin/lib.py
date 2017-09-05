# IVXV Internet voting framework
"""
Library for collector management service.
"""

import datetime
import logging
import os

import dateutil.parser

from . import CONFIG_TYPES, VOTING_LIST_TYPES
from . import (COLLECTOR_STATE_CONFIGURED, COLLECTOR_STATE_FAILURE,
               COLLECTOR_STATE_INSTALLED, COLLECTOR_STATE_NOT_INSTALLED,
               COLLECTOR_STATE_PARTIAL_FAILURE)
from . import (SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE,
               SERVICE_STATE_INSTALLED, SERVICE_STATE_NOT_INSTALLED)
from .config import CONFIG
from .db import IVXVManagerDb, DB_SERVICE_SUBKEYS
from .ivxv_pkg import COLLECTOR_PACKAGE_FILENAMES

# create logger
log = logging.getLogger(__name__)


def gen_service_record_defaults(db, config):
    """
    Add database records for services with default values.

    Parse the list of services from technical config and create
    missing database records for every defined service.
    """
    assert 'network' in config

    # generate list of default values
    service_values = {}
    for network in config['network']:
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
    for service_id, service_defaults in service_values.items():
        db_key_prefix = 'service/{}'.format(service_id)
        try:
            db.get_value(db_key_prefix + '/service-type')
            continue
        except KeyError:
            log.info('Registering new service %s in management service',
                     service_id)

        for key, val in service_defaults.items():
            db.set_value(db_key_prefix + '/' + key, val)

    set_service_cond_values(db, config)


def set_service_cond_values(db, config):
    """Set/remove service conditional values."""

    for network in config['network']:
        for service_type, services in sorted(network['services'].items()):
            for service in services or []:
                db_key_prefix = 'service/{}'.format(service['id'])
                # create 'dds-token-key' for dds/choices/voting services
                # only if auth.ticket block exist in technical config
                if service_type in ['dds', 'choices', 'voting']:
                    manage_db_cond_value(db, db_key_prefix + '/dds-token-key',
                                         'ticket' in config.get('auth', {}))
                # create 'tls-key' and 'tls-cert' for
                # choices/dds/storage/verification/voting services
                if service_type in ['choices', 'dds', 'storage',
                                    'verification', 'voting']:
                    manage_db_cond_value(db, db_key_prefix + '/tls-key', True)
                    manage_db_cond_value(db, db_key_prefix + '/tls-cert', True)
                # create 'tspreg-key' for voting service
                # only if registration service is DigiDocService
                if service_type == 'voting':
                    manage_db_cond_value(
                        db, db_key_prefix + '/tspreg-key',
                        [True
                         for i in config.get('qualification', [])
                         if i.get('protocol') == 'tspreg'])


def manage_db_cond_value(db, key, set_value):
    """Manage (create/remove) database value."""
    if set_value:  # create key
        try:
            db.get_value(key)
        except KeyError:
            db.set_value(key, '')
    else:  # remove key
        try:
            db.rm_value(key)
        except KeyError:
            pass


def gen_host_record_defaults(db, config):
    """
    Add database records for service hosts with default values.

    Parse the list of services from technical config and create
    missing database records for service hosts.
    """
    # generate list of default values
    hostnames = set()
    for network in config['network']:
        for services in network['services'].values():
            for service in services or []:
                hostnames |= {service['address'].split(':')[0]}

    # set host default values
    for hostname in sorted(hostnames):
        db_key_prefix = 'host/{}'.format(hostname)
        try:
            db.get_value(db_key_prefix + '/state')
            continue
        except KeyError:
            log.info(
                'Registering new service host %s in management service',
                hostname)
        db.set_value(db_key_prefix + '/state', '')


def get_service_conf_status(db, config):
    """
    Get list of services that requires config update.

    Check technical config records against database and generate list of
    services that require config update.

    Returns dict:

    .. code-block:: text

        {
            'service-id': {
                'technical': boolean,
                'election': boolean,
                'choices': boolean,
                'voters': [list-number, ...],
            },
            ...
        }
    """
    assert 'network' in config

    # detect config versions
    tech_conf_ver = db.get_value('config/technical')
    election_conf_ver = db.get_value('config/election')

    # detect need for list updates
    update_choices_list = (db.get_value('list/choices') !=
                           db.get_value('list/choices-loaded'))
    update_voters_list = []
    voter_list_no = 1
    while True:
        try:
            voters_list_not_applied = (
                db.get_value('list/voters%02d' % voter_list_no) !=
                db.get_value('list/voters%02d-loaded' % voter_list_no))
        except KeyError:
            break
        if voters_list_not_applied:
            update_voters_list.append(voter_list_no)
        voter_list_no += 1

    service_list = {}

    # create list of services
    for network in config['network']:
        for service_type, services in sorted(network['services'].items()):
            for service in services or []:
                db_key_prefix = 'service/{}'.format(service['id'])

                service_tech_conf_ver = db.get_value(
                    db_key_prefix + '/technical-conf-version')
                service_election_conf_ver = (
                    election_conf_ver and
                    db.get_value(db_key_prefix + '/election-conf-version'))

                service_list[service['id']] = {
                    'technical': service_tech_conf_ver != tech_conf_ver,
                    'election': service_election_conf_ver != election_conf_ver,
                    'choices':
                    service_type == 'choices' and update_choices_list,
                    'voters': update_voters_list,
                }
                if service_list[service['id']]['technical']:
                    log.debug('Service %s have no latest version of '
                              'technical config', service['id'])

                if service_list[service['id']]['election']:
                    log.debug('Service %s have no latest version of '
                              'election config', service['id'])

    return service_list


def config_type_verbose(config_type):
    """Get config type as human readable string."""
    try:
        return CONFIG_TYPES[config_type]
    except KeyError:
        return VOTING_LIST_TYPES[config_type]


def generate_service_list(service_networks, service_id_filter):
    """
    Generate service list from collector technical config.

    .. code-block:: text

        [
            [<service_type>, {param: value, ...}],
            ...
        ]

    :param service_networks: Service networks from collector technical config
                             (section 'networks').
    :type service_networks: dict
    :param service_id_filter: Service IDs to include.
                              All services are included if empty.
    :type service_id_filter: list

    :returns: List of services or None if some of specified services not found.
    """
    service_list = []
    service_ids = service_id_filter[:]
    for network in service_networks:
        for service_type, services in sorted(network['services'].items()):
            for service_data in services or []:
                service_id = service_data['id']
                if service_id_filter and service_id not in service_ids:
                    log.debug('Skipping service %s', service_id)
                    continue
                if service_ids:
                    service_ids.remove(service_id)
                service_list.append([service_type, service_data])

    if service_ids:
        log.error('Invalid service ID: %s', ', '.join(service_ids))
        return

    return service_list


def get_services(require_collector_status=None, service_status=None,
                 include_types=None, exclude_types=None):
    """
    Generate filtered list of services.

    :param require_collector_status: Require specified status for collector
    :type require_collector_status: list
    :param include_types: Include specified service types
    :type include_types: list
    :param exclude_types: Exclude specified service types
    :type exclude_types: list
    :returns: dict of services (key is service ID and value
              is dict with service database values) or
              None if collector is not in expected state
    """
    require_collector_status = require_collector_status or []
    assert isinstance(require_collector_status, list)
    service_status = service_status or []
    assert isinstance(service_status, list)
    include_types = include_types or []
    exclude_types = exclude_types or []
    assert isinstance(include_types, list)
    assert isinstance(exclude_types, list)
    assert not include_types or not exclude_types

    # collect status data
    db = IVXVManagerDb()
    collector_status = generate_collector_status(db)
    db.close()

    # check collector status
    if (require_collector_status and
            collector_status['collector_status']
            not in require_collector_status):
        log.info('Collector service status is %s',
                 collector_status['collector_status'])
        if len(require_collector_status) == 1:
            log.error('Collector status must be %s for this operation',
                      require_collector_status[0])
        else:
            log.error('Collector status must be %s or %s for this operation',
                      ', '.join(require_collector_status[:-1]),
                      require_collector_status[-1])
        return

    # create list of services
    services = {}
    for service_id, service_data in collector_status['service'].items():
        if service_status and service_data.get('state') not in service_status:
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


def clean_directory(path):
    """Remove all files from directory."""
    log.debug('Cleaning directory %s', path)
    for filename in os.listdir(path):
        role_filename = os.path.join(path, filename)
        os.unlink(role_filename)


def generate_collector_status(db):
    """Generate collector status data."""
    status = db.get_all_values()
    service_states = [val['state']
                      for val in status.get('service', {}).values()]

    # generate service hints
    if status.get('service'):
        generate_service_hints(status['service'])

    # group services by network
    if status.get('service'):
        status['network'] = dict()
        for service_id, service_rec in status.get('service').items():
            nw_name = service_rec.pop('network')
            status['network'][nw_name] = status['network'].get(nw_name, dict())
            status['network'][nw_name][service_id] = service_rec

    # check package files
    assert 'storage' not in status
    status['storage'] = {'debs_exists': [], 'debs_missing': []}
    pkg_path = CONFIG.get('deb_pkg_path')
    for pkg_filename in COLLECTOR_PACKAGE_FILENAMES.values():
        pkg_filepath = os.path.join(pkg_path, pkg_filename)
        key = 'debs_exists' if os.path.exists(pkg_filepath) else 'debs_missing'
        status['storage'][key].append(pkg_filepath)

    # detect collector status
    status['collector_status'] = detect_collector_state(status, service_states)

    # detect counts of voter lists
    list_counters = {'voters-list-loaded': 0,
                     'voters-list-pending': 0}
    voter_list_no = 0
    while True:
        key = 'voters%02d' % (voter_list_no + 1)
        try:
            if not status['list'][key]:
                break
        except KeyError:
            break
        if status['list'][key] == status['list'][key + '-loaded']:
            list_counters['voters-list-loaded'] += 1
        else:
            list_counters['voters-list-pending'] += 1
        voter_list_no += 1
    status['list'].update(list_counters)

    # detect election phase
    status['election']['phase'] = None
    status['election']['phase-start'] = None
    status['election']['phase-end'] = None
    if status['election']:

        def get_ts(name):
            """
            Get timestamp value as datetime.datetime object
            or None if value is not set.
            """
            value = status['election'][name]
            return dateutil.parser.parse(value) if value else None

        electionstart = get_ts('electionstart')
        electionstop = get_ts('electionstop')
        servicestart = get_ts('servicestart')
        servicestop = get_ts('servicestop')
        ts_format = '%Y-%m-%dT%H:%M %Z'

        phases = [
            [not electionstart, 'PREPARING', None, None],
            [servicestart and
             datetime.datetime.now(servicestart.tzinfo) < servicestart,
             'WAITING FOR SERVICE START', None, servicestart],
            [electionstart and
             datetime.datetime.now(electionstart.tzinfo) < electionstart,
             'WAITING FOR ELECTION START', servicestart, electionstart],
            [electionstop and
             datetime.datetime.now(electionstop.tzinfo) < electionstop,
             'ELECTION', electionstart, electionstop],
            [servicestop and
             datetime.datetime.now(servicestop.tzinfo) < servicestop,
             'WAITING FOR SERVICE STOP', electionstop, servicestop],
            [True, 'FINISHED', servicestop, None],
        ]
        for phase_data in phases:
            if phase_data[0]:
                break
        status['election']['phase'] = phase_data[1]
        status['election']['phase-start'] = (
            phase_data[2].strftime(ts_format)
            if phase_data[2] else '-')
        status['election']['phase-end'] = (
            phase_data[3].strftime(ts_format)
            if phase_data[3] else '-')

    return status


def generate_service_hints(services):
    """Generate hints for services."""
    for service_id, params in services.items():
        # ordered list of hints
        hints = [
            ['Apply technical config', not params['technical-conf-version']],
        ]
        # add other hints if services is not a log collector
        if params['service-type'] != 'log':
            hints += [
                ['Install TLS key', not params.get('tls-key', True)],
                ['Install TLS certificate', not params.get('tls-cert', True)],
                ['Install mobile ID identity token key',
                 not params.get('dds-token-key', True)],
                ['Install TSP registration key',
                 not params.get('tspreg-key', True)],
                ['Apply election config', not params['election-conf-version']],
            ]

        for hint, is_relevant in hints:
            if is_relevant:
                services[service_id]['hint'] = hint
                break


def populate_user_permissions(db):
    """Populate user permissions for Apache web server."""
    permissions_path = CONFIG.get('permissions_path')
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


def detect_collector_state(status, service_states):
    """Detect collector state."""
    # NOT INSTALLED
    # Stay in NOT INSTALLED state while:
    # - some service is not installed
    if (status['collector']['status'] == COLLECTOR_STATE_NOT_INSTALLED and
            (not status['config']['technical'] or
             not service_states or
             SERVICE_STATE_NOT_INSTALLED in service_states)):
        return COLLECTOR_STATE_NOT_INSTALLED

    # INSTALLED
    # Stay in INSTALLED state until technical config is applied to all services
    if (status['collector']['status'] == COLLECTOR_STATE_INSTALLED and
            SERVICE_STATE_INSTALLED in service_states):
        return COLLECTOR_STATE_INSTALLED

    # CONFIGURED
    if SERVICE_STATE_FAILURE not in service_states:
        return COLLECTOR_STATE_CONFIGURED

    # collect service states by type
    service_state_by_type = {}
    for _, service_data in sorted(status['service'].items()):
        service_state_by_type[service_data['service-type']] = (
            service_state_by_type.get(service_data['service-type'], []))
        service_state_by_type[service_data['service-type']].append(
            service_data['state'])

    # FAILURE
    important_service_types = ['choices', 'dds', 'proxy', 'storage',
                               'verification', 'voting']
    for service_type in important_service_types:
        if (service_type in service_state_by_type and
                SERVICE_STATE_CONFIGURED not in
                service_state_by_type[service_type]):
            return COLLECTOR_STATE_FAILURE

    # PARTIAL FAILURE
    return COLLECTOR_STATE_PARTIAL_FAILURE


def detect_voters_list_order_no(db):
    """Detect next voters list order number from database."""
    voter_list_no = 0
    while True:
        try:
            db.get_value('list/voters%02d' % (voter_list_no + 1))
        except KeyError:
            break
        voter_list_no += 1
    return voter_list_no


def ask_user_confirmation(question):
    """Ask user confirmation."""
    while True:
        answer = input(question).upper()
        if answer in 'YN':
            return answer == 'Y'
