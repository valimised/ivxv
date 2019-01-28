# IVXV Internet voting framework
"""
Agent daemon for collector management service.

This daemon is managed by systemd.

* Systemd service file:
  :file:`/lib/systemd/system/ivxv-admin.service`

* Query daemon status:
  :command:`systemctl status ivxv-admin`

* Start daemon:
  :command:`systemctl stop ivxv-admin`

* Stop daemon:
  :command:`systemctl status ivxv-admin`
"""

import datetime
import json
import logging
import os
import re
import subprocess
import time

import dateutil.parser

from . import (CMD_DESCR, RFC3339_DATE_FORMAT, SERVICE_MONITORING_STATES,
               SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE, lib)
from .cli_utils import init_cli_util
from .config import cfg_path
from .db import DB_FILE_PATH, IVXVManagerDb
from .event_log import register_service_event
from .service.service import Service

# create logger
log = logging.getLogger(__name__)

#: Service ping interval in seconds.
PING_INTERVAL = 60
#: Stats file path.
STATS_FILEPATH = cfg_path('admin_ui_data_path', 'stats.json')
#: Maximum count of automatic attempts to apply config
MAX_AUTO_ATTEMPTS = 3
#: Unix epoch timestamp
UNIX_EPOCH_TIMESTAMP = '1970-01-01T00:00:00'


def main_loop():
    """Agent daemon main loop."""
    args = init_cli_util("""
        IVXV Collector Management Service agent daemon.

        Usage: ivxv-agent-daemon [--get-stats]

        Options:
            --get-stats     Copy statistics from Log Monitor to
                            Management Service without daemonizing.
    """)

    # copy stats and finish process
    if args['--get-stats']:
        return generate_stats_data(force=True)

    # daemon process
    log.info('Starting Collector Management Service agent daemon')
    check_management_db()

    while True:
        loop_start_time = datetime.datetime.now()

        # generate stats/state data
        state = None
        try:
            lib.PidLocker.rm_stale_pidfile('ivxv-config-apply.pid')
            state = generate_state_data()
            generate_stats_data()
        except OSError as err:
            log.error(err)
        except Exception as err:  # pylint: disable=broad-except
            log.error(
                'Unknown error while generating stats/state data: %s', err)

        # start config applying if required
        if state:
            apply_cfg(state)

        # pause after loop
        duration = datetime.datetime.now() - loop_start_time
        if duration.seconds < 5:
            log.debug('Sleeping for 5 seconds')
            time.sleep(5)
        else:
            time.sleep(1)


def generate_state_data():
    """
    Ping services, generate state data and write it to :file:`status.json`.
    """
    # fetch services data
    services = get_collector_data()
    if services:
        # sort services list in order of last data
        services_sorted = sorted([
            [service['last-data'], service_id]
            for service_id, service in services.items()])

        # check services
        for last_data_timestamp, service_id in services_sorted:
            service_data = services[service_id]
            if service_data['state'] not in SERVICE_MONITORING_STATES:
                # log.debug('Service %s state is %s, skipping check',
                #           service_id, service_data['state'])
                continue
            next_check_timestamp = (
                dateutil.parser.parse(last_data_timestamp or
                                      UNIX_EPOCH_TIMESTAMP) +
                datetime.timedelta(seconds=PING_INTERVAL))
            if next_check_timestamp > datetime.datetime.now():
                log.debug('Service %s next check is in the future, '
                          'skipping check', service_id)
                continue

            ping_service(service_id, service_data)
            time.sleep(1)

    # generate state data
    with IVXVManagerDb(for_update=True) as db:
        state = register_collector_state(db)

    # generate config applying state
    state['config-apply'] = {}
    for cfg_key in ['trust', 'technical', 'election', 'choices', 'districts']:
        _get_cfg_applying_state(cfg_key, state)
    for list_no in range(1, 100):
        if not _get_cfg_applying_state(f'voters{list_no:02}', state):
            break

    # write state data to file
    state_filename = cfg_path('admin_ui_data_path', 'status.json')
    state['meta'] = get_agent_metadata()
    with open(state_filename, 'w') as fp:
        fp.write(json.dumps(state, indent=4, sort_keys=True))

    return state


def _get_cfg_applying_state(cfg_key, state):
    """Get config applying state.

    :return: True if config file exist and state data is generated,
             False if config file does not exist.
    """
    cfg_filepath = lib.get_loaded_cfg_file_path(cfg_key)
    if cfg_filepath is None:
        return False
    state_file_filepath = cfg_filepath.replace('.bdoc', '.json')

    try:
        with open(state_file_filepath) as fp:
            apply_state = json.load(fp)
        state['config-apply'][cfg_key] = {
            'version': apply_state['config_version'],
            'attempts': apply_state['attempts'],
            'completed': apply_state['completed'],
            'state_file': os.path.basename(state_file_filepath),
        }
    except FileNotFoundError:
        state['config-apply'][cfg_key] = {}
    state['config-apply'][cfg_key]['cmd_file'] = os.path.basename(cfg_filepath)

    return True


def get_agent_metadata():
    """Generate agent metadata block."""
    timestamp = datetime.datetime.utcnow()
    return dict(generator='IVXV Management Service Agent Daemon',
                time_generated=timestamp.strftime('%Y-%m-%dT%H:%M:%SZ'))


def apply_cfg(state):
    """Apply config files."""
    _apply_cfg_for_services('technical', 'technical', state)
    if state['config']['technical']:
        for cfg_key in ['election', 'choices']:
            _apply_cfg_for_services(cfg_key, cfg_key, state)
    for list_no in range(1, 100):
        if _apply_cfg_for_services(
                'voters', f'voters{list_no:02}', state) is None:
            break


def _apply_cfg_for_services(cfg_type, cfg_key, state):
    """Apply config for services.

    :return: None if config file does not exist,
             True if config applying is started,
             False if config applying is not started.
    """
    cfg_filepath = lib.get_loaded_cfg_file_path(cfg_key)
    if cfg_filepath is None:
        return None

    state_file_filepath = cfg_filepath.replace('.bdoc', '.json')
    with open(state_file_filepath) as fp:
        state = json.load(fp)

    # check preconditions
    if (not state['autoapply'] or state['completed']
            or state['attempts'] >= MAX_AUTO_ATTEMPTS):
        return False

    if lib.PidLocker.pidfile_exists('ivxv-config-apply.pid'):
        log.info('Can\'t start automatic applying of %s, pidfile exists',
                 CMD_DESCR[cfg_type])
        return False

    # execute config applying command
    log.info('Automatically apply %s, attempt #%d',
             CMD_DESCR[cfg_type], state['attempts'] + 1)
    subprocess.Popen(['ivxv-config-apply', f'--type={cfg_type}'])

    return True


def ping_service(service_id, service_data):
    """Ping service."""
    service = Service(service_id, service_data)
    service_ok = service.ping()

    if not _is_db_accessible(service.get_db_key('last-data')):
        return False

    # register result in database
    with IVXVManagerDb(for_update=True) as db:
        db.set_value(
            service.get_db_key('last-data'),
            datetime.datetime.now().strftime(RFC3339_DATE_FORMAT))

        ping_errors_key = service.get_db_key('ping-errors')
        ping_errors_old = db.get_value(ping_errors_key)

        service_state_key = service.get_db_key('state')
        service_state_old = db.get_value(service_state_key)

        if service_state_old not in SERVICE_MONITORING_STATES:
            log.warning('Service has removed from monitoring')
            return False

        service_state_new = service_state_old
        if service_ok:
            log.debug('Service %s is alive', service_id)
            ping_errors_new = '0'
            service_state_new = SERVICE_STATE_CONFIGURED
            if ping_errors_old != '0':
                db.set_value(ping_errors_key, '0')
        else:
            ping_errors_new = str(int(ping_errors_old) + 1)
            if (int(ping_errors_new) >= 3
                    and service_state_old != SERVICE_STATE_FAILURE):
                log.warning('Status check failed three times, '
                            'setting service state from %s to FAILURE',
                            service_state_old)
                service_state_new = SERVICE_STATE_FAILURE
            else:
                log.warning('Status check for service %s failed (%s times). '
                            'Service state is %s',
                            service_id, ping_errors_new, service_state_old)

        if ping_errors_old != ping_errors_new:
            db.set_value(ping_errors_key, ping_errors_new)
        if service_state_old != service_state_new:
            db.set_value(service_state_key, service_state_new)
            register_collector_state(db)

    return service_ok


def generate_stats_data(force=False):
    """Generate stats data and write it to :file:`stats.json`.

    :param force: Force check even the next check timestamp is in the future.
    :type force: bool
    """
    # read existing stats file
    stats = {}
    try:
        with open(STATS_FILEPATH) as fp:
            stats = json.load(fp)
    except json.decoder.JSONDecodeError as err:
        log.error('Invalid JSON in existing stats file %s: %s',
                  STATS_FILEPATH, err)
    except OSError as err:
        log.error('Cannot load existing stats JSON file %s: %s',
                  STATS_FILEPATH, err)
    stats.setdefault('data', {})

    # import stats file from Log Monitor
    logmon_address, last_data_timestamp = get_logmon_data()
    if logmon_address:
        next_check_timestamp = (
            dateutil.parser.parse(last_data_timestamp) +
            datetime.timedelta(seconds=PING_INTERVAL))
        if not force and next_check_timestamp > datetime.datetime.now():
            log.debug('Log Monitor service next check is in the future, '
                      'skipping check')
        else:
            stats_data = query_logmon_stats(logmon_address)
            if 'error' in stats_data:
                stats['error'] = stats_data['error']
            else:
                _normalize_stats(stats_data)
                stats['data'] = stats_data
                try:
                    del stats['error']
                except KeyError:
                    pass
    else:
        stats = {'error': 'Log Monitor address is not defined'}
        log.error(stats['error'])

    # update metadata
    stats['meta'] = get_agent_metadata()

    # write stats data to file
    with open(STATS_FILEPATH, 'w') as fp:
        fp.write(json.dumps(stats, indent=4, sort_keys=True))


def _normalize_stats(stats_data):
    """Normalize Log Monitor stats.

    Convert dictionaries to sorted lists.
    """
    for district_id, district_stats in sorted(stats_data.items()):
        if district_id == 'time':
            continue
        for stats_key, stats_val in district_stats.items():
            if isinstance(stats_val, dict):
                stats_data[district_id][stats_key] = sorted(
                    [[item[0], item[1]] for item in stats_val.items()],
                    reverse=True
                )[:10 if stats_key == 'operating-systems' else 100]

    # generate empty values for missing districts.
    # Log Monitor does not have district list and cannot generate stats for
    # districts that have no voter data. Generate empty data for such
    # districts.
    districts = []
    districts_filename = cfg_path('admin_ui_data_path', 'districts.json')
    try:
        with open(districts_filename) as fp:
            districts = json.load(fp)
    except FileNotFoundError:
        log.debug('Missing districts list. '
                  'Will not generate empty blocks for missing districts')
    empty_stats = dict(
        [key, (list() if isinstance(val, list) else 0)]
        for key, val in stats_data['TOTAL'].items())
    for district_id, _ in districts:
        if district_id not in stats_data:
            stats_data[district_id] = empty_stats


def query_logmon_stats(address):
    """Query stats file from Log Monitor service.

    Query stats file from Log Monitor service and register last query timestamp
    in the management database.

    :return: Stats data. Existing stats data is returned with error message in
             ``error`` value if query fails.
    :rtype: dict
    """
    ssh_cmd = [
        'ssh', '-T', '-o', 'PreferredAuthentications=publickey',
        '{}@{}'.format('logmon', address), 'cat', '/var/lib/ivxv/stats.json'
    ]

    log.debug('Querying stats from Log Monitor')
    proc = subprocess.run(
        ssh_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    stats = {}
    if proc.returncode:
        log.error(
            'Querying stats from Log Monitor finished with error code %d',
            proc.returncode)
        stats['error'] = (
            'Error while transporting stats data from Log Monitor over SSH:'
            '\n {}'.format(proc.stderr.decode('utf-8')))
    else:
        try:
            stats = json.loads(proc.stdout.decode('utf-8'))
        except json.JSONDecodeError as err:
            errmsg = '%s: line %d column %d (char %d)' % (
                err.msg, err.lineno, err.colno, err.pos)
            stats['error'] = f'Invalid JSON data from Log Monitor: {errmsg}'
            log.error('Error while parsing JSON stats from Log Monitor: %s',
                      errmsg)

    if _is_db_accessible('logmonitor/last-data'):
        with IVXVManagerDb(for_update=True) as db:
            db.set_value('logmonitor/last-data',
                         datetime.datetime.now().strftime(RFC3339_DATE_FORMAT))

    return stats


def check_management_db():
    """Check management database file.

    Wait if database file does not exist.
    """
    # wait if database does not exist
    check_interval = 1
    check_notif_interval = 60
    check_counter = 0
    while not os.path.exists(DB_FILE_PATH):
        if check_counter % check_notif_interval == 0:
            log.warning('Collector management database does not exist, '
                        'waiting')
        check_counter += 1
        time.sleep(check_interval)

    log.info('Collector management database is available')


def get_collector_data():
    """Read collector data from management database.

    :return: service data or None if election config is not loaded.
    :rtype: dict
    """
    with IVXVManagerDb() as db:
        election_cfg_ver = db.get_value('config/election')
        if not election_cfg_ver:
            return None

        services = {}
        for key in db.keys():
            if re.match(r'service/', key):
                service_id, field_name = key.split('/')[1:]
                services[service_id] = services.get(service_id, {})
                services[service_id][field_name] = db.get_value(key)

    return services


def get_logmon_data():
    """Read logmonitor address from management database.

    :return: Log Monitor data: [address, last_data_timestamp].
    :rtype: list
    """
    with IVXVManagerDb() as db:
        address = db.get_value('logmonitor/address')
        last_data_timestamp = db.get_value('logmonitor/last-data')

    if not last_data_timestamp:
        last_data_timestamp = UNIX_EPOCH_TIMESTAMP

    return [address, last_data_timestamp]


def register_collector_state(db):
    """Detect collector state and register state change in database."""
    state = lib.generate_collector_state(db)
    last_state = state['collector']['state']
    if state['collector_state'] != last_state:
        log.info('Registering new state for collector: %s',
                 state['collector_state'])
        db.set_value('collector/state', state['collector_state'])
        register_service_event(
            'COLLECTOR_STATE_CHANGE',
            params={
                'state': state['collector_state'],
                'last_state': last_state
            })

    return state


def _is_db_accessible(check_value):
    """Try to read field value from database before writing it.

    If this fails, then something is happened with database (e.g. database is
    recreated during service reset) and writing to database must be cancelled.
    """
    with IVXVManagerDb() as db:
        try:
            db.get_value(check_value)
        except KeyError:
            log.warning('Cannot read value "%s" from database', check_value)
            return False

    return True
