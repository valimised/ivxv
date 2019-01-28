# IVXV Internet voting framework
"""
Config file loader for collectors management service.
"""

import configparser
import logging
import logging.config
import os


#: Config file name.
CFG_FILE_NAME = 'ivxv-collector-admin.conf'
#: Default values for config.
CFG_DEFAULTS = {
    # base directory for management service data files
    'ivxv_admin_data_path': os.environ.get('IVXV_ADMIN_DATA_PATH',
                                           '/var/lib/ivxv'),
    # directory for admin UI static data files (used directly by web UI)
    'admin_ui_data_path': '%(ivxv_admin_data_path)s/admin-ui-data',
    # directory for admin UI permissions
    'permissions_path': '%(ivxv_admin_data_path)s/admin-ui-permissions',
    # directory for applied command files
    'command_files_path': '%(ivxv_admin_data_path)s/commands',
    # directory for collector config files
    'active_config_files_path': '/etc/ivxv',
    # directory for uploaded files
    'file_upload_path': '%(ivxv_admin_data_path)s/upload',
    # directory for exported votes
    'exported_votes_path': '%(ivxv_admin_data_path)s/ballot-box',
    # directory for ivxv debian packages
    'deb_pkg_path': '/etc/ivxv/debs',
    # management database directory
    'ivxv_db_path': '%(ivxv_admin_data_path)s/db',
    # management database file path
    'ivxv_db_file_path': '%(ivxv_db_path)s/ivxv-management.db',
}
CFG_PARSER = configparser.ConfigParser(defaults=CFG_DEFAULTS)
CFG_FILES_USED = []
#: Config file paths.
CFG_PATHS = [
    os.path.join(os.curdir, CFG_FILE_NAME),
    os.path.join('/etc/ivxv', CFG_FILE_NAME)
]
if os.environ.get('IVXV_ADMIN_CONF'):
    CFG_PATHS += [
        os.environ.get('IVXV_ADMIN_CONF'),
        os.path.join(os.environ.get('IVXV_ADMIN_CONF'), CFG_FILE_NAME)
    ]

for FILE_PATH in CFG_PATHS:
    if os.path.isfile(FILE_PATH):
        CFG_FILES_USED.append(FILE_PATH)
        logging.config.fileConfig(FILE_PATH)
        CFG_PARSER.read(FILE_PATH)

# check config files read
if not CFG_FILES_USED:
    log = logging.getLogger(__name__)
    log.error('IVXV collector admin utils config file %s not found '
              'in the search paths %s',
              CFG_FILE_NAME, CFG_PATHS)

CONFIG = CFG_PARSER['DEFAULT']


def cfg_path(cfg_path_name, filename):
    """Generate full path for specified file."""
    return os.path.join(CONFIG[cfg_path_name], filename)


if __name__ == '__main__':
    log = logging.getLogger(__name__)
    log.info('Loading config file(s) %s succeeded',
             ', '.join((CFG_FILES_USED)))
    exit()
