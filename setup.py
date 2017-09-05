# IVXV Internet voting framework

"""
Setup of Collector Management Service.
"""

from setuptools import setup

setup(
    name='IVXVCollectorAdminDaemon',
    version='0.9',
    description='IVXV Collector Management Service',
    author='IVXV Developer',
    author_email='info@ivotingcentre.ee',
    install_requires=['bottle', 'docopt', 'jinja2', 'pyopenssl',
                      'python-dateutil', 'python-debian', 'pyyaml'],
    packages=['ivxv_admin'],
    package_dir={'': 'collector-admin'},
    package_data={'ivxv_admin': ['templates/*.jinja', 'templates/*.json']},
    entry_points={
        'console_scripts': [
            # collector management
            'ivxv-collector-init'
            '=ivxv_admin.admin_util:ivxv_collector_init_util',
            'ivxv-create-data-dirs'
            '=ivxv_admin.admin_util:ivxv_create_data_dirs_util',

            # management service database
            'ivxv-db=ivxv_admin.admin_util:database_util',
            'ivxv-db-dump=ivxv_admin.admin_util:database_dump_util',
            'ivxv-db-reset=ivxv_admin.admin_util:database_reset_util',

            # user management
            'ivxv-users-list=ivxv_admin.admin_util:users_list_util',

            # config management
            'ivxv-config-apply=ivxv_admin.admin_util:config_apply_util',

            # service management
            'ivxv-logmonitor-copy-log'
            '=ivxv_admin.admin_util:initialize_logmonitor_util',
            'ivxv-secret-import'
            '=ivxv_admin.admin_util:import_secret_data_file_util',
            'ivxv-status=ivxv_admin.admin_util:status_util',
            'ivxv-votes-export=ivxv_admin.admin_util:export_votes_util',

            # command loading
            'ivxv-cmd-load=ivxv_admin.admin_util:command_load_util',

            # daemons
            'ivxv-admin-httpd=ivxv_admin.http_daemon:daemon',
            'ivxv-agent-daemon=ivxv_admin.agent_daemon:main_loop',
        ],
    },
)
