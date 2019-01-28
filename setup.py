# IVXV Internet voting framework
"""
Setup of Collector Management Service.
"""

from setuptools import setup

setup(
    name='IVXVCollectorAdminDaemon',
    version='1.4.1',
    description='IVXV Collector Management Service',
    author='IVXV Developer',
    author_email='info@ivotingcentre.ee',
    install_requires=[
        'bottle',
        'docopt',
        'jinja2',
        'pyopenssl',
        'python-dateutil',
        'python-debian',
        'pyyaml',
    ],
    packages=[
        'ivxv_admin',
        'ivxv_admin.cli_utils',
        'ivxv_admin.cli_utils.config_utils',
        'ivxv_admin.config_validator',
        'ivxv_admin.service',
    ],
    package_dir={'': 'collector-admin'},
    package_data={'ivxv_admin': ['templates/*.jinja', 'templates/*.json']},
    entry_points={
        'console_scripts': [
            # collector management (storage utilities)
            'ivxv-collector-init'
            '=ivxv_admin.cli_utils.admin_storage_utils:'
            'ivxv_collector_init_util',
            'ivxv-create-data-dirs'
            '=ivxv_admin.cli_utils.admin_storage_utils:'
            'ivxv_create_data_dirs_util',

            # management service database (storage utilities)
            'ivxv-db=ivxv_admin.cli_utils.admin_storage_utils:database_util',
            'ivxv-db-dump='
            'ivxv_admin.cli_utils.admin_storage_utils:database_dump_util',
            'ivxv-db-reset='
            'ivxv_admin.cli_utils.admin_storage_utils:database_reset_util',

            # user management
            'ivxv-users-list='
            'ivxv_admin.cli_utils.status_utils:users_list_util',

            # config management
            'ivxv-config-apply='
            'ivxv_admin.cli_utils.config_utils.config_apply:main',
            'ivxv-secret-load'
            '=ivxv_admin.cli_utils.config_utils.load_secret_data_file:main',
            'ivxv-config-validate='
            'ivxv_admin.cli_utils.config_utils.config_validate:main',

            # service management
            'ivxv-backup='
            'ivxv_admin.cli_utils.backup_utils:backup_util',
            'ivxv-backup-crontab='
            'ivxv_admin.cli_utils.backup_utils:backup_crontab_generator_util',
            'ivxv-logmonitor-copy-log'
            '=ivxv_admin.cli_utils.service_utils:copy_logs_to_logmon_util',
            'ivxv-status='
            'ivxv_admin.cli_utils.status_utils:status_util',
            'ivxv-service'
            '=ivxv_admin.cli_utils.service_utils:manage_service',
            'ivxv-consolidate-votes='
            'ivxv_admin.cli_utils.service_utils:consolidate_votes_util',
            'ivxv-update-packages'
            '=ivxv_admin.cli_utils.service_utils:update_software_pkg_util',

            # command loading
            'ivxv-cmd-load='
            'ivxv_admin.cli_utils.config_utils.command_load:main',
            'ivxv-cmd-remove='
            'ivxv_admin.cli_utils.config_utils.command_remove:main',

            # logging
            'ivxv-eventlog-dump=ivxv_admin.event_log:event_log_dump_util',

            # daemons
            'ivxv-admin-httpd=ivxv_admin.http_daemon:daemon',
            'ivxv-agent-daemon=ivxv_admin.agent_daemon:main_loop',
        ],
    },
)
