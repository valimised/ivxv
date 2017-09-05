# IVXV Internet voting framework
"""
Agent daemon for collector management service.

This daemon is managed by systemd.

    Systemd service file:

        /lib/systemd/system/ivxv-admin.service

    Query daemon status:

        systemctl status ivxv-admin

    Start daemon:

        systemctl stop ivxv-admin

    Stop daemon:

        systemctl status ivxv-admin
"""

import datetime
import json
import logging
import os
import re
import time

import dateutil.parser

from . import (RFC3339_DATE_FORMAT, SERVICE_STATE_CONFIGURED,
               SERVICE_STATE_FAILURE)
from . import lib
from .cli import init_cli_util
from .config import CONFIG
from .db import DB_FILE_PATH, IVXVManagerDb
from .service_config import Service

# create logger
log = logging.getLogger(__name__)

# monitor services in the following states
MONITORED_STATES = [SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE]
# service ping interval in seconds
PING_INTERVAL = 60


def main_loop():
    """Agent daemon main loop."""
    init_cli_util("""
        IVXV Collector Management Service agent daemon.

        Usage: ivxv-agent-daemon
    """)

    log.info('Starting Collector Management Service agent daemon')
    check_management_db()

    while True:
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
                if service_data['state'] not in MONITORED_STATES:
                    # log.debug('Service %s state is %s, skipping check',
                    #           service_id, service_data['state'])
                    continue
                next_check_timestamp = (
                    dateutil.parser.parse(last_data_timestamp) +
                    datetime.timedelta(seconds=PING_INTERVAL))
                if next_check_timestamp > datetime.datetime.now():
                    log.debug('Service %s next check is in the future, '
                              'skipping check', service_id)
                    continue

                ping_service(service_id, service_data)
                time.sleep(1)

        # generate status data
        db = get_db_handler()
        status = register_collector_status(db)
        db.close()

        # write status data to file
        status_filename = os.path.join(CONFIG.get('admin_ui_data_path'),
                                       'status.json')
        status['meta'] = {
            'generator': 'IVXV Management Service Agent Daemon',
            'time_generated':
            datetime.datetime.now().strftime(RFC3339_DATE_FORMAT),
        }
        with open(status_filename, 'w') as fp:
            fp.write(json.dumps(status, indent=4, sort_keys=True))

        time.sleep(1)


def ping_service(service_id, service_data):
    """Ping service."""
    service = Service(service_id, service_data)
    service_ok = service.ping()

    # register result in database
    db = get_db_handler(for_update=True)
    # try to read service ping last data from database
    # before writing it. this is required to detect
    # database initialization during ping command
    try:
        db.get_value(service.get_db_key('last-data'))
    except KeyError:
        db.close()
        return

    db.set_value(service.get_db_key('last-data'),
                 datetime.datetime.now().strftime(RFC3339_DATE_FORMAT))

    ping_errors_key = service.get_db_key('ping-errors')
    ping_errors_old = db.get_value(ping_errors_key)

    service_state_key = service.get_db_key('state')
    service_state_old = db.get_value(service_state_key)

    service_state_new = service_state_old
    if service_ok:
        log.debug('Service %s is alive', service_id)
        ping_errors_new = '0'
        service_state_new = SERVICE_STATE_CONFIGURED
        if ping_errors_old != '0':
            db.set_value(ping_errors_key, '0')
    else:
        ping_errors_new = str(int(ping_errors_old) + 1)
        if (int(ping_errors_new) >= 3 and
                service_state_old != SERVICE_STATE_FAILURE):
            log.warning(
                'Status check failed three times, '
                'setting service state from %s to FAILURE', service_state_old)
            service_state_new = SERVICE_STATE_FAILURE
        else:
            log.warning('Status check for service %s failed (%s times). '
                        'Service state is %s',
                        service_id, ping_errors_new, service_state_old)

    if ping_errors_old != ping_errors_new:
        db.set_value(ping_errors_key, ping_errors_new)
    if service_state_old != service_state_new:
        db.set_value(service_state_key, service_state_new)
        register_collector_status(db)

    db.close()


def check_management_db():
    """
    Check management database file and wait if database file does not exist.
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
    """
    Read collector data from management database.

    Returns dict with service data or None if election config is not loaded.
    """
    db = get_db_handler()
    election_conf_ver = db.get_value('config/election')
    if not election_conf_ver:
        return

    services = {}
    for key in db.keys():
        if re.match(r'service/', key):
            service_id, field_name = key.split('/')[1:]
            services[service_id] = services.get(service_id, {})
            services[service_id][field_name] = db.get_value(key)
    db.close()

    return services


def register_collector_status(db):
    """Detect collector status and register new status if required."""
    status = lib.generate_collector_status(db)
    if status['collector_status'] != status['collector']['status']:
        log.info('Registering new state for collector: %s',
                 status['collector_status'])
        if db.read_only:
            db.close()
            db = get_db_handler(for_update=True)
        db.set_value('collector/status', status['collector_status'])

    return status


def get_db_handler(for_update=False):
    """Create handler for management database."""
    db = None
    while db is None:
        try:
            db = IVXVManagerDb(for_update=for_update)
        except OSError as err:
            log.error('Failed to open collector management database: %s. '
                      'Retrying after 3 seconds', err)
            time.sleep(3)
            continue
    return db
