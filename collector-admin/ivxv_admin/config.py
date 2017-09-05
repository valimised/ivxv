# IVXV Internet voting framework
"""
Config for collectors management service.
"""

import configparser
import logging
import logging.config
import os


# load config from file
CONFIG_FILE_NAME = 'ivxv-collector-admin.conf'
CONFIG_DEFAULTS = {
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
CONFIG_PARSER = configparser.ConfigParser(defaults=CONFIG_DEFAULTS)
CONFIG_FILES_USED = []
CONFIG_PATHS = [os.path.join(os.curdir, CONFIG_FILE_NAME),
                os.path.join('/etc/ivxv', CONFIG_FILE_NAME)]
if os.environ.get('IVXV_ADMIN_CONF'):
    CONFIG_PATHS += [os.environ.get('IVXV_ADMIN_CONF'),
                     os.path.join(os.environ.get('IVXV_ADMIN_CONF'),
                                  CONFIG_FILE_NAME)]

for FILE_PATH in CONFIG_PATHS:
    if os.path.isfile(FILE_PATH):
        CONFIG_FILES_USED.append(FILE_PATH)
        logging.config.fileConfig(FILE_PATH)
        CONFIG_PARSER.read(FILE_PATH)

# check config files read
if not CONFIG_FILES_USED:
    log = logging.getLogger(__name__)
    log.error('IVXV collector admin utils config file %s not found '
              'in the search paths %s',
              CONFIG_FILE_NAME, CONFIG_PATHS)

CONFIG = CONFIG_PARSER['DEFAULT']

if __name__ == '__main__':
    log = logging.getLogger(__name__)
    log.info('Loading config file(s) %s succeeded',
             ', '.join((CONFIG_FILES_USED)))
    exit()
