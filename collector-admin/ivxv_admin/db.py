# IVXV Internet voting framework
"""
Database abstraction layer for collector management service.
"""

import datetime
import dbm.gnu
import logging
import os
import re
import time

import dateutil.parser

from . import (COLLECTOR_STATE_NOT_INSTALLED, COLLECTOR_STATES,
               RFC3339_DATE_FORMAT, SERVICE_STATE_NOT_INSTALLED,
               SERVICE_STATES)
from .config import CONFIG

#: Path to database file
DB_FILE_PATH = CONFIG['ivxv_db_file_path']

#: Database keys with default values
DB_KEYS = {
    # collector state
    'collector/state': COLLECTOR_STATE_NOT_INSTALLED,
    # election config file version in management service
    'config/election': '',
    # technical config file version in management service
    'config/technical': '',
    # trust root config file version in management service
    'config/trust': '',
    # logmonitor address
    'logmonitor/address': '',
    # logmonitor timestamp of last fetch
    'logmonitor/last-data': '',
    # choices list file version in management service
    'list/choices': '',
    # choices list file version in choices service
    'list/choices-loaded': '',
    # districts list file version in management service
    'list/districts': '',
    # districts list file version in choices service
    'list/districts-loaded': '',
    # election ID
    'election/election-id': '',
    # election start time
    'election/electionstart': '',
    # election stop time
    'election/electionstop': '',
    # collector service start time
    'election/servicestart': '',
    # collector service stop time
    'election/servicestop': '',
}
#: Database keys for service hosts with default values
DB_HOST_SUBKEYS = {
    # host state
    'state': '',
}
#: Database keys for services with default values
DB_SERVICE_SUBKEYS = {
    # service type
    'service-type': None,
    # technical config file version in service
    'technical-conf-version': '',
    # election config file version in service
    'election-conf-version': '',
    # service network ID
    'network': None,
    # service state
    'state': SERVICE_STATE_NOT_INSTALLED,
    # last data by service PING (timestamp)
    'last-data': '',
    # count of ping errors
    'ping-errors': '0',
    # service IP address
    'ip-address': None,
    # registered background information (usually error message)
    'bg_info': '',
}
#: Database keys for certain service types
DB_SERVICE_CONDITIONAL_SUBKEYS = {
    # mobile ID identity token key file checksum (sha256)
    'mid-token-key': '',
    # TLS certificate file checksum (sha256)
    'tls-cert': '',
    # TLS key file checksum (sha256)
    'tls-key': '',
    # TSP registration key file checksum (sha256)
    'tspreg-key': '',
    # automatic backup times for backup service
    'backup-times': '',
}
#: Full list of allowed database keys
ALLOWED_SERVICE_KEYS = (
    list(DB_SERVICE_SUBKEYS) + list(DB_SERVICE_CONDITIONAL_SUBKEYS))

# create logger
log = logging.getLogger(__name__)


class IVXVManagerDb:
    """
    Management service database abstraction class.

    Based on :py:mod:`dbm.gnu` module.
    """
    db = None  #: Database object.
    read_only = None  #: Database access mode.
    _retries = None  #: Retry count for failed database open.
    _retry_delay = None  #: Delay between retries.
    _db_mode = None  #: Mode string for :py:func:`dbm.open`.

    def __init__(self, for_update=False, retries=30, retry_delay=0.1):
        """
        Constructor.

        :param for_update: Open database for update
        :type for_update: bool
        :param retries: Retries before giving up
        :type retries: int
        :param retry_delay: Pause between retries (in seconds)
        :type retry_delay: float
        """
        self.read_only = not for_update
        self._db_mode = 'ws' if for_update else 'r'
        self._retries = retries
        self._retry_delay = retry_delay

    def __enter__(self, mode=None):
        """Enter the runtime context, open database."""
        mode = mode or self._db_mode
        log.debug("Opening management database %r (mode: %s)", DB_FILE_PATH, mode)
        retries = self._retries
        if mode != 'n' and not os.path.exists(DB_FILE_PATH):
            log.error("Database file %r not found", DB_FILE_PATH)
            raise FileNotFoundError()
        db_error = OSError()
        while retries > 0:
            try:
                self.db = dbm.gnu.open(DB_FILE_PATH, mode)
                break
            except OSError as err:
                if err.errno == 11:  # database is locked
                    log.debug('Database is locked, retrying')
                else:
                    log.debug("Can't open database, retrying (%s)", err)
                db_error = err
            retries -= 1
            time.sleep(self._retry_delay)
        else:
            log.error('Error while opening management database: %s', db_error)
            raise db_error

        return self

    def __exit__(self, *args):
        """Exit the runtime context, close database."""
        log.debug('Closing management database')
        self.db.close()

    def get_value(self, key):
        """Get value from database."""
        return self.db[key].decode('UTF-8')

    def set_value(self, key, value, safe=False):
        """Validate and set database value.

        :param key: Key name
        :type key: str
        :param value: Value to set
        :type value: str
        :param safe: Is operation safe or not. Safe operation will try to read
                     old value before writing new one.
        :type safe: bool
        """
        assert isinstance(key, str)

        # validate value
        if isinstance(value, datetime.datetime):
            value = value.strftime(RFC3339_DATE_FORMAT)
        assert isinstance(value, str), f"Invalid value type: {type(value)}"

        # set value
        if key in [
                'election/election-id', 'logmonitor/address',
                'logmonitor/last-data'
        ]:
            pass
        elif key == 'collector/state':
            assert value in COLLECTOR_STATES, f"Invalid value for {key!r}: {value!r}"
        elif re.match('election/(election|service)(start|stop)$', key):
            assert not value or dateutil.parser.parse(value)
        elif (re.match('election/auth/.+$', key)
              or key == 'election/tsp-qualification'):
            assert value == 'TRUE'
        elif key in DB_KEYS or key == "list/voters0000":
            if value != '':
                assert value.count(" ") == 1, f"Invalid value {value!r}"
                cn, timestamp = value.split(' ')
                assert re.match(r'^(.+,){2}\d{11}$', cn)
                dateutil.parser.parse(timestamp)
        elif re.match(r"list/voters[0-9]{4}$", key):
            assert value
            assert value.count(" ") == 1, f"Invalid value {value!r}"
            url_or_signature, timestamp = value.split(" ")
            dateutil.parser.parse(timestamp)
        elif re.match(r"list/voters[0-9]{4}-state$", key):
            assert value in ["PENDING", "APPLIED", "INVALID", "SKIPPED"]
        elif re.match(r'host/.+/.+$', key):
            key_type = key.split('/')[2]
            assert key_type in DB_HOST_SUBKEYS, f"Invalid host key type {key_type}"
        elif re.match(r'service/.+/.+$', key):
            key_type = key.split('/')[2]
            assert (
                key_type in ALLOWED_SERVICE_KEYS
            ), f"Invalid service key type {key_type}"
            if key_type == 'state':
                assert value in SERVICE_STATES, f"Invalid value for {key!r}: {value!r}"
            elif key_type == 'backup-times':
                assert value == "" or re.match(
                    r"[0-9]{2}:[0-9]{2}( [0-9]{2}:[0-9]{2})*$", value
                ), f"Invalid value for {key!r}: {value!r}"
        elif re.match(r'user/.+$', key):
            assert re.match(r'user/(.+,){2}\d{11}$', key), (
                f"Invalid user CN {key.split('/')[1]!r}")
        else:
            raise KeyError(f"Invalid database field name {key!r}")

        # Safe operation reads value before writing it.
        # This is to avoid unneeded values after database initialization (e.g.
        # agent daemon may try to update service data in background).
        if safe:
            self.db[key]

        log.debug("Setting value %r = %r", key, value)
        self.db[key] = value

    def rm_value(self, key):
        """Remove database record."""
        assert isinstance(key, str)

        log.debug("Removing record %r", key)
        del self.db[key]

    def get_all_values(self, section=None):
        """
        Get all database values as a dictionary.

        :param section: Config section to get
        :type section: string

        :return: dict
        """
        values = {}
        for key in sorted(self.keys()):
            path = key.split('/')
            assert len(path) in (2, 3)
            if path[0] not in values:
                values[path[0]] = {}
            if len(path) == 2:
                values[path[0]][path[1]] = self.get_value(key) or None
            else:
                if path[1] not in values[path[0]]:
                    values[path[0]][path[1]] = {}
                values[path[0]][path[1]][path[2]] = self.get_value(key) or None

        return values.get(section, {}) if section else values

    @classmethod
    def reset(cls):
        """Reset database."""
        log.info("Initializing management database %r", DB_FILE_PATH)
        db = cls()
        db.__enter__(mode='n')

        # write default values
        for key, value in sorted(DB_KEYS.items()):
            db.set_value(key, value)

        db.__exit__()

    def dump(self, filter_keys=None):
        """
        Dump database to stdout.

        :param filter_keys: Limit dump with specified values
        :type filter_keys: list
        """
        for key in self.keys():
            if not filter_keys or key in filter_keys:
                print(f'{key}: {self.get_value(key)}')

    def keys(self):
        """Return all database keys in sorted order."""
        return [key.decode('UTF-8') for key in sorted(self.db.keys())]


def check_db_dir():
    """
    Check database directory exist or not.

    :return: Database file directory path or None if path does not exist.
    :rtype: str
    """
    db_file_path = CONFIG['ivxv_db_file_path']
    db_path = os.path.dirname(db_file_path)
    if os.path.exists(db_path):
        return db_file_path

    log.error('Database directory %r does not exist', db_path)

    return None
