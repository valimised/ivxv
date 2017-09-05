# IVXV Internet voting framework
"""
Database abstraction layer for collector management service.
"""

import datetime
import dbm.gnu
import logging
import re
import time

import dateutil.parser

from . import (COLLECTOR_STATES, COLLECTOR_STATE_NOT_INSTALLED,
               RFC3339_DATE_FORMAT, SERVICE_STATES, SERVICE_STATE_NOT_INSTALLED)
from .config import CONFIG

DB_FILE_PATH = CONFIG.get('ivxv_db_file_path')

# database keys with default values
DB_KEYS = {
    # collector state
    'collector/status': COLLECTOR_STATE_NOT_INSTALLED,
    # election config file version in management service
    'config/election': '',
    # technical config file version in management service
    'config/technical': '',
    # trust root config file version in management service
    'config/trust': '',
    # choices list file version in management service
    'list/choices': '',
    # choices list file version in choices service
    'list/choices-loaded': '',
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
# database keys for service hosts with default values
DB_HOST_SUBKEYS = {
    # host state
    'state': '',
}
# database keys for services with default values
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
    # last data by service PING
    'last-data': '',
    # count of ping errors
    'ping-errors': '0',
    # service IP address
    'ip-address': None,
}
# database keys for certain service types
DB_SERVICE_CONDITIONAL_SUBKEYS = {
    # mobile ID identity token key file checksum (sha256)
    'dds-token-key': '',
    # TLS certificate file checksum (sha256)
    'tls-cert': '',
    # TLS key file checksum (sha256)
    'tls-key': '',
    # TSP registration key file checksum (sha256)
    'tspreg-key': '',
}
# full list of allowed database keys
ALLOWED_SERVICE_KEYS = (
    list(DB_SERVICE_SUBKEYS.keys()) +
    list(DB_SERVICE_CONDITIONAL_SUBKEYS.keys()))

# create logger
log = logging.getLogger(__name__)


class IVXVManagerDb:
    """
    Management service database abstraction class.

    Based on dbm.gnu module.
    """
    db = None  #: database object
    read_only = None  #: database access mode

    def __init__(self, open_db=True, for_update=False):
        """
        Constructor.

        :param open_db: Open database
        :type open_db: bool
        :param for_update: Open database for update
        :type for_update: bool
        :param retries: Retries before giving up
        :type retries: int
        :param pause_between_retries: Pause between retries (in seconds)
        :type pause_between_retries: float
        """
        self.read_only = not for_update
        if open_db:
            self.open(mode='ws' if for_update else 'r')

    def close(self):
        """Close database."""
        log.debug('Closing management database')
        self.db.close()

    def get_value(self, key):
        """Get value from database."""
        return self.db[key].decode('UTF-8')

    def set_value(self, key, value):
        """Validate and set database value."""
        assert isinstance(key, str)

        # validate value
        if isinstance(value, datetime.datetime):
            value = value.strftime(RFC3339_DATE_FORMAT)
        assert isinstance(value, str), 'Invalid value type: %s' % type(value)

        # set value
        if key == 'collector/status':
            assert value in COLLECTOR_STATES, (
                'Invalid value for %s: %s' % (key, value))
        elif key == 'election/election-id':
            pass
        elif re.match('election/(election|service)(start|stop)$', key):
            if value:
                assert dateutil.parser.parse(value)
        elif key in DB_KEYS or re.match(r'list/voters[0-9]{2}(-loaded)?$', key):
            if value != '':
                assert len(value.split(' ')) == 2
                cn, timestamp = value.split(' ')
                assert re.match(r'^(\w+,){2}\d{11}$', cn)
                dateutil.parser.parse(timestamp)
        elif re.match(r'host/.+/.+$', key):
            key_type = key.split('/')[2]
            assert key_type in DB_HOST_SUBKEYS, (
                'Invalid host key type %s' % key_type)
        elif re.match(r'service/.+/.+$', key):
            key_type = key.split('/')[2]
            assert key_type in ALLOWED_SERVICE_KEYS, (
                'Invalid service key type %s' % key_type)
            if key_type == 'state':
                assert value in SERVICE_STATES, (
                    'Invalid value for %s: %s' % (key, value))
        elif re.match(r'user/.+$', key):
            assert re.match(r'user/.+,.+,[0-9]{11}$', key), (
                'Invalid user CN: %s' % key.split('/')[1])
        else:
            raise KeyError('Invalid database field name: ' + key)

        log.debug('Setting value %s = "%s"', key, value)
        self.db[key] = value

    def rm_value(self, key):
        """Remove database record."""
        assert isinstance(key, str)

        log.debug('Removing record %s', key)
        del self.db[key]

    def get_all_values(self, section=None):
        """
        Get all database values as a dictionary.

        :param section: Config section to get
        :type section: string

        :returns: dict
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

    def open(self, mode):
        """Open database."""
        log.debug('Opening management database %s (mode: %s)',
                  DB_FILE_PATH, mode)
        retries = 10
        pause_between_retries = 0.1
        db_error = Exception()
        while retries:
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
            time.sleep(pause_between_retries)
        else:
            log.error('Error while opening management database: %s',
                      db_error)
            raise db_error

    def reset(self):
        """Reset database."""
        assert self.db is None
        log.info('Initializing management database %s', DB_FILE_PATH)
        self.open(mode='n')

        # write default values
        for key, value in sorted(DB_KEYS.items()):
            self.set_value(key, value)

        self.close()

    def dump(self):
        """Dump database."""
        for key in sorted(self.keys()):
            print('{}: {}'.format(key, self.get_value(key)))

    def keys(self):
        """Return all database keys in sorted order."""
        return [key.decode('UTF-8') for key in sorted(self.db.keys())]
