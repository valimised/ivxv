# IVXV Internet voting framework
"""CLI utilities for command removing."""

import os.path
import re

from .. import init_cli_util, log
from ...config import cfg_path
from ...db import IVXVManagerDb
from ...event_log import register_service_event


def main():
    """Remove command from management service."""
    args = init_cli_util("""
    Remove command from IVXV Collector Management Service.

    Usage: ivxv-cmd-remove <type>

    Options:

        <type>  Config type. Currently only voters list
                (value "voters") is supported.
    """)

    if args['<type>'] != 'voters':
        log.error('Invalid config type specified: %s', args['<type>'])
        return 1

    with IVXVManagerDb(for_update=True) as db:
        # pending lists are always latest ones
        for key in sorted(db.get_all_values(section='list'), reverse=True):
            key_name = 'list/' + key
            if (re.match(r'voters[0-9]{2}-loaded', key)
                    and not db.get_value(key_name)):
                filepath = cfg_path(
                    'active_config_files_path',
                    re.sub(r'-loaded', '.bdoc', key))
                cfg_version = db.get_value(key_name.replace('-loaded', ''))
                log.info('Removing voters list "%s"', cfg_version)
                db.rm_value(key_name.replace('-loaded', ''))
                db.rm_value(key_name)
                try:
                    os.unlink(filepath)
                except FileNotFoundError:
                    log.warning('File %s is already removed', filepath)
                register_service_event(
                    'CMD_REMOVED',
                    params={'cmd_type': 'voters list', 'version': cfg_version})

    return 0
