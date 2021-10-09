# IVXV Internet voting framework
"""CLI utilities for management service data storage."""

import os
import re
import shutil
import sys

from ..config import CONFIG, cfg_path
from ..db import IVXVManagerDb, check_db_dir
from ..event_log import init_event_log
from ..lib import IvxvError, clean_dir
from ..service.backup_service import remove_backup_crontab
from . import ask_user_confirmation, init_cli_util, log

#: Config value names for Management Service data directories
MANAGEMENT_PATH_PARAM_NAMES = [
    'ivxv_admin_data_path',
    'admin_ui_data_path',
    'permissions_path',
    'command_files_path',
    'active_config_files_path',
    'file_upload_path',
    'exported_votes_path',
    'deb_pkg_path',
    'ivxv_db_path',
    'vis_path',
]


def ivxv_create_data_dirs_util():
    """Create management service data directories."""
    init_cli_util(
        """
        Create IVXV Collector Management Service data directories.

        NOTE: Directory owners and permissions are not set by this utility!

        Usage: ivxv-create-data-dirs
    """,
        allow_root=True)

    # config parameter names for directories
    for cfg_var in MANAGEMENT_PATH_PARAM_NAMES:
        dirname = CONFIG[cfg_var]
        if os.path.exists(dirname):
            log.info("Path %r already exist", dirname)
        else:
            log.info("Creating data directory %r", dirname)
            os.mkdir(dirname)
        if not os.path.isdir(dirname):
            log.error("Path %r is not a directory", dirname)
            return 1

    return 0


def ivxv_collector_init_util():
    """Initialize IVXV Collector."""
    args = init_cli_util("""
    Initialize IVXV Collector.

    Usage: ivxv-collector-init [--force]

    Options:
        --force     Don't ask user confirmation
    """)

    # ask confirmation
    if not args['--force']:
        if not ask_user_confirmation(
                'Do You want to initialize IVXV Collector (Y=yes) ?'):
            return 1

    # remove crontab if exist
    remove_backup_crontab()

    # initialize data directories
    try:
        init_data_directories()
    except IvxvError as err:
        log.error(err)
        return 1

    # initialize management database
    init_management_database()

    init_event_log()

    return 0


def database_util():
    """Management service database utility."""
    args = init_cli_util("""
    Add, remove or modify key/value pairs in
    IVXV Collector Management Service database.

    WARNING!
        Use this utility only for testing purposes!
        Never change database in production systems!

    Usage: ivxv-db <key> <value>
        ivxv-db --del <key>
    """)

    # check database directory
    if not check_db_dir():
        return 1

    # process record
    with IVXVManagerDb(for_update=True) as db:
        if args['--del']:
            db.rm_value(args['<key>'])
            log.info("Database value %r successfully removed", args["<key>"])
        else:
            db.set_value(args['<key>'], args['<value>'])
            log.info("Database value %r set to %r", args["<key>"], args["<value>"])

    return 0


def database_dump_util():
    """Dump management service database."""
    args = init_cli_util("""
        Dump IVXV Collector Management Service database.

        Usage: ivxv-db-dump [<key>] ...
    """)

    log.info('Dumping IVXV management database')

    # check database directory
    if not check_db_dir():
        return 1

    with IVXVManagerDb() as db:
        db.dump(args['<key>'])

    return 0


def database_reset_util():
    """Reset management service database."""
    args = init_cli_util("""
        Reset IVXV Collector Management Service database.

        Usage: ivxv-db-reset [--force]

        Options:
            --force     Don't ask user confirmation
    """)

    # check database directory
    db_file_path = check_db_dir()
    if not db_file_path:
        return 1

    # ask confirmation
    if not args['--force']:
        if not ask_user_confirmation(
                'Do You want to reset IVXV management database (Y=yes) ?'):
            return 1

    # initialize database
    init_management_database()

    # initialize data files
    init_management_datafiles()

    # initialize data directories
    clean_dir(CONFIG['file_upload_path'])

    return 0


def init_data_directories():
    """Initialize data directories."""
    for cfg_var in MANAGEMENT_PATH_PARAM_NAMES:
        dirpath = CONFIG[cfg_var]
        if not os.path.exists(dirpath):
            log.info("Creating directory %r", dirpath)
            os.mkdir(dirpath)
        elif not os.path.isdir(dirpath):
            raise IvxvError(f"Path {dirpath!r} is not a directory")
        if cfg_var not in [
            "ivxv_admin_data_path",
            "deb_pkg_path",
            "active_config_files_path",
        ]:
            clean_dir(dirpath)
        patterns = (
            r"(choices|districts|election|technical|trust|voters0000)\.bdoc$",
            r"voters[0-9]{2}\.zip$",
        )
        for filename in os.listdir(CONFIG["active_config_files_path"]):
            if any(re.match(pattern, filename) for pattern in patterns):
                os.unlink(cfg_path("active_config_files_path", filename))

    init_management_datafiles()


def init_management_database():
    """Initialize management database."""
    log.debug('Initializing IVXV management database')

    IVXVManagerDb.reset()

    log.info('New management database is created with default values')


def init_management_datafiles():
    """Initialize management data files."""
    # install empty stats.json to admin UI data path
    module_path = os.path.dirname(sys.modules['ivxv_admin'].__file__)
    file_path = os.path.join(module_path, 'templates/stats.json')
    shutil.copy(file_path, CONFIG['admin_ui_data_path'])
