# IVXV Internet voting framework
"""Event logging for Collector Management service."""

import datetime
import json
import os

from . import EVENT_LOG_FILENAME, EVENTS, RFC3339_DATE_FORMAT
from .cli_utils import init_cli_util
from .config import cfg_path

EVENT_LOG_FILEPATH = cfg_path('ivxv_admin_data_path', EVENT_LOG_FILENAME)


def init_event_log():
    """Initialize Management Service event log."""
    try:
        os.unlink(EVENT_LOG_FILEPATH)
    except FileNotFoundError:
        pass
    register_service_event('COLLECTOR_INIT')


def register_service_event(event, level='INFO', service=None, params=None):
    """Register Management Service event in event log file."""
    params = params or {}
    assert level in ['INFO', 'ERROR']
    log_event = {
        'event': event,
        'level': level,
        'message': EVENTS[event].format(**params),
        'service': service or 'management',
        'timestamp': datetime.datetime.now().strftime(RFC3339_DATE_FORMAT),
    }
    with open(EVENT_LOG_FILEPATH, 'a') as fp:
        json.dump(log_event, fp, sort_keys=True)
        fp.write('\n')


def event_log_dump_util():
    """Dump collector management event log."""
    init_cli_util(
        """
        Dump IVXV Collector Management event log in human readable format.

        Usage: ivxv-eventlog-dump
        """,
        allow_root=True)

    with open(EVENT_LOG_FILEPATH) as fp:
        while True:
            line = fp.readline()
            if not line:
                break
            event = json.loads(line)
            print(
                '{timestamp} {level} {service} {event} {message}'
                .format(**event))
