# IVXV Internet voting framework
"""
Collector state for collector management service.
"""

import datetime
import json
import os

import dateutil.parser

from . import (COLLECTOR_PKG_FILENAMES, COLLECTOR_STATE_CONFIGURED,
               COLLECTOR_STATE_FAILURE, COLLECTOR_STATE_INSTALLED,
               COLLECTOR_STATE_NOT_INSTALLED, COLLECTOR_STATE_PARTIAL_FAILURE,
               SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE,
               SERVICE_STATE_INSTALLED, SERVICE_STATE_NOT_INSTALLED,
               SERVICE_TYPE_PARAMS)
from .config import CONFIG, cfg_path
from .service import generate_service_hints


def generate_collector_state(db):
    """Generate collector state data."""
    state = db.get_all_values()

    # generate service data
    if state.get('service'):
        generate_service_hints(state['service'])

        # group services by network
        state['network'] = dict()
        for service_id, service_rec in state.get('service').items():
            nw_name = service_rec.pop('network')
            state['network'].setdefault(nw_name, dict())
            state['network'][nw_name][service_id] = service_rec

    # check package files
    assert 'storage' not in state
    state['storage'] = {
        'debs_exists': [],
        'debs_missing': [],
        'command_files': [],
        'command_files_applied': [],
        'command_files_pending': [],
    }
    for pkg_filename in COLLECTOR_PKG_FILENAMES.values():
        pkg_filepath = cfg_path('deb_pkg_path', pkg_filename)
        key = 'debs_exists' if os.path.exists(pkg_filepath) else 'debs_missing'
        state['storage'][key].append(pkg_filepath)

    # collect command files state
    for filename in os.listdir(CONFIG['command_files_path']):
        filepath = cfg_path('command_files_path', filename)
        if os.path.splitext(filename)[1] == '.json':
            continue
        try:
            with open('{}.json'.format(os.path.splitext(filepath)[0])) as fp:
                cmd_state = json.load(fp)
            state_key = (
                'command_files_applied' if cmd_state['completed']
                else 'command_files_pending')
        except FileNotFoundError:
            state_key = 'command_files_applied'
        state['storage']['command_files'].append(filepath)
        state['storage'][state_key].append(filepath)

    # detect collector state
    state['collector_state'] = detect_collector_state(state)

    generate_voters_list_state(state)
    generate_election_state(state)

    return state


def detect_collector_state(state):
    """Detect collector state.

    :param state: Collector state data. Probably values from
                  Collector Management Database.
    :type state: dict
    """
    service_states = [
        val['state'] for val in state.get('service', {}).values()
    ]

    # NOT INSTALLED
    # Stay in NOT INSTALLED state while:
    # - some service is not installed
    if (state['collector']['state'] == COLLECTOR_STATE_NOT_INSTALLED and
            (not state['config']['technical'] or
             not service_states or
             SERVICE_STATE_NOT_INSTALLED in service_states)):
        return COLLECTOR_STATE_NOT_INSTALLED

    # INSTALLED
    # Stay in INSTALLED state until technical config is applied to all services
    if (state['collector']['state'] == COLLECTOR_STATE_INSTALLED
            and SERVICE_STATE_INSTALLED in service_states):
        return COLLECTOR_STATE_INSTALLED

    # CONFIGURED
    if SERVICE_STATE_FAILURE not in service_states:
        return COLLECTOR_STATE_CONFIGURED

    # collect service states by type
    service_state_by_type = {}
    for _, service_data in sorted(state['service'].items()):
        service_state_by_type[service_data['service-type']] = (
            service_state_by_type.get(service_data['service-type'], []))
        service_state_by_type[service_data['service-type']].append(
            service_data['state'])

    # FAILURE
    for service_type, service_params in SERVICE_TYPE_PARAMS.items():
        if not service_params['main_service']:
            continue
        if (service_type in service_state_by_type
                and SERVICE_STATE_CONFIGURED not in
                service_state_by_type[service_type]):
            return COLLECTOR_STATE_FAILURE

    # PARTIAL FAILURE
    return COLLECTOR_STATE_PARTIAL_FAILURE


def generate_voters_list_state(state):
    """Generate voters list block for collector state data."""
    state['list'].update({'voters-list-loaded': 0, 'voters-list-pending': 0})
    for voter_list_no in range(1, 100):
        key = f'voters{voter_list_no:02d}'
        if key not in state['list']:
            break
        if state['list'][key] == state['list'][f'{key}-loaded']:
            state['list']['voters-list-loaded'] += 1
        else:
            state['list']['voters-list-pending'] += 1


def generate_election_state(state):
    """Generate election block for collector state data."""
    state['election']['phase'] = None
    state['election']['phase-start'] = None
    state['election']['phase-end'] = None

    if not state['election']:
        return

    def get_ts(name):
        """
        Get timestamp value as datetime.datetime object
        or None if value is not set.
        """
        value = state['election'][name]
        return dateutil.parser.parse(value) if value else None

    electionstart = get_ts('electionstart')
    electionstop = get_ts('electionstop')
    servicestart = get_ts('servicestart')
    servicestop = get_ts('servicestop')
    ts_format = '%Y-%m-%dT%H:%M %Z'

    phases = [
        [not electionstart, 'PREPARING', None, None],
        [
            servicestart
            and datetime.datetime.now(servicestart.tzinfo) < servicestart,
            'WAITING FOR SERVICE START', None, servicestart
        ],
        [
            electionstart
            and datetime.datetime.now(electionstart.tzinfo) < electionstart,
            'WAITING FOR ELECTION START', servicestart, electionstart
        ],
        [
            electionstop
            and datetime.datetime.now(electionstop.tzinfo) < electionstop,
            'ELECTION', electionstart, electionstop
        ],
        [
            servicestop
            and datetime.datetime.now(servicestop.tzinfo) < servicestop,
            'WAITING FOR SERVICE STOP', electionstop, servicestop
        ],
        [True, 'FINISHED', servicestop, None],
    ]
    for phase_active, name, start_time, stop_time in phases:
        if phase_active:
            state['election']['phase'] = name
            state['election']['phase-start'] = (
                start_time.strftime(ts_format) if start_time else '-')
            state['election']['phase-end'] = (
                stop_time.strftime(ts_format) if stop_time else '-')
            break
