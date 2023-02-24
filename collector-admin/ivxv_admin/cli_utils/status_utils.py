# IVXV Internet voting framework
"""CLI utilities for management service state."""

import json
import re

from jinja2 import Environment, PackageLoader

from .. import lib
from ..db import IVXVManagerDb
from . import init_cli_util, log


def users_list_util():
    """User listing utility for collector administrator interface."""
    init_cli_util("""
    List IVXV Collector Management Service registered users.

    Usage: ivxv-users-list
    """)

    # read users from database
    with IVXVManagerDb() as db:
        users = db.get_all_values('user')

    # output users
    user_no = 0
    for user_cn, permissions in sorted(users.items()):
        user_no += 1
        print(f"{user_no}. {user_cn}   Permissions: {permissions}")

    if not user_no:
        print('No users defined')


def status_util():
    """Output collector state."""
    args = init_cli_util("""
    Output IVXV Collector state.

    Usage: ivxv-status [--json] [--service=<service-id> ...] [<filter> ...]

    Options:
        --json                  Output full data in JSON format.
                                Note: filters have no effect in JSON output.
        --service=<service-id>  Filter output by service ID.
                                Note: This filter conflicts other section
                                filters than "smart" or "service".
        <filter>                Filter output by section. Possible values are:
                                  * collector - collector state;
                                  * election - election data;
                                  * config - versions of loaded config;
                                  * list - versions of loades lists;
                                  * service - service information;
                                  * ext - external service information;
                                  * storage - storage information;
                                  * smart - output only relevant data;
                                  * all - output all data;
                                [Default: smart].
    """)
    filters = args['<filter>'] or ['smart']

    # validate CLI args
    data_sections = [
        'collector', 'election', 'config', 'list', 'service', 'ext', 'storage']
    for arg in filters:
        if arg not in data_sections + ['all', 'smart']:
            log.error('Invalid section: %s', arg)
            return 1
    if 'all' in filters:
        filters = data_sections
    if args['--service']:
        if filters not in (['smart'], ['service']):
            log.error('Service ID can only used with '
                      '"smart" or "service" section filters')
            return 1
        filters = ['service']
    if 'smart' in filters:
        filters = ['smart', 'collector', 'storage']

    # generate state data
    with IVXVManagerDb() as db:
        state = lib.generate_collector_state(db)

    # output JSON
    if args['--json']:
        print(json.dumps(state, indent=4, sort_keys=True))
        return 0

    # apply smart filtering
    if 'smart' in filters:
        cfg_state_filters(filters, state)

    # apply service ID filters
    if args['--service']:
        for service_id in args['--service']:
            if service_id not in state['service'].keys():
                log.error("Unknown service ID %r", service_id)
                return 1
        for network_id in list(state['network']):
            for service_id in list(state['network'][network_id]):
                if service_id not in args['--service']:
                    del state['network'][network_id][service_id]
            if not state['network'][network_id]:
                del state['network'][network_id]

    # output plain text
    tmpl_env = Environment(loader=PackageLoader('ivxv_admin', 'templates'))
    template = tmpl_env.get_template('ivxv_status.jinja')
    output = template.render(state, sections=filters)
    output = re.sub(r'\n\n+', '\n\n', output)
    print(output.strip())

    return 0


def cfg_state_filters(filters, state):
    """Reconfigure list of state filters depending on actual state."""
    if (state['config']['trust'] or
            state['config']['technical'] or
            state['config']['election']):
        filters += ['config']
    if state['config']['election']:
        filters += ['election']
    if (state['list']['choices-loaded'] or
            state['list']['choices'] or
            state['list']['districts-loaded'] or
            state['list']['districts'] or
            state['list'].get('voters0000') or
            state['list'].get('voters0000-loaded')):
        filters += ['list']
    if state.get('service'):
        filters += ['service']
