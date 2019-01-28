# IVXV Internet voting framework
"""Collector Management Service."""

__version__ = '1.4.1'
DEB_PKG_VERSION = '1.4.1'

#: Management daemon data
MANAGEMENT_DAEMON_PORT = 8080
MANAGEMENT_DAEMON_URL = 'http://localhost:%s/' % MANAGEMENT_DAEMON_PORT

#: RFC3339 date format
RFC3339_DATE_FORMAT = '%Y-%m-%dT%H:%M:%S.%f'
#: RFC3339 date format (without second fractions)
RFC3339_DATE_FORMAT_WO_FRACT = '%Y-%m-%dT%H:%M:%SZ'

#: User permissions
PERMISSION_BALLOT_BOX_DOWNLOAD = 'download-ballot-box'
PERMISSION_ELECTION_CONF = 'election-conf-admin'
PERMISSION_LOG_VIEW = 'log-view'
PERMISSION_STATS_VIEW = 'stats-view'
PERMISSION_TECH_CONF = 'tech-conf-admin'
PERMISSION_USERS_ADMIN = 'user-admin'
#: User roles
USER_ROLES = {
    'admin': {
        'description': 'Administrator',
        'permissions': (
            PERMISSION_BALLOT_BOX_DOWNLOAD,
            PERMISSION_ELECTION_CONF,
            PERMISSION_LOG_VIEW,
            PERMISSION_STATS_VIEW,
            PERMISSION_TECH_CONF,
            PERMISSION_USERS_ADMIN,
        ),
    },
    'election-conf-manager': {
        'description': 'Election config manager',
        'permissions': (
            PERMISSION_BALLOT_BOX_DOWNLOAD,
            PERMISSION_ELECTION_CONF,
            PERMISSION_STATS_VIEW
        ),
    },
    'viewer': {
        'description': 'Viewer',
        'permissions': (
            PERMISSION_STATS_VIEW,
        ),
    },
    'none': {
        'description': 'No permissions',
        'permissions': tuple(),
    },
}

#: Config types
CFG_TYPES = {
    'trust': 'trust root configuration',
    'election': 'elections configuration',
    'technical': 'collectors technical configuration',
}
#: Voting list types
VOTING_LIST_TYPES = {
    'choices': 'choices list',
    'districts': 'districts list',
    'voters': 'voters list',
}
#: Command types
CMD_TYPES = list(CFG_TYPES) + list(VOTING_LIST_TYPES) + ['user']
#: Command descriptions
CMD_DESCR = {'user': 'user permissions configuration'}
CMD_DESCR.update(CFG_TYPES)
CMD_DESCR.update(VOTING_LIST_TYPES)


#: Collector states
COLLECTOR_STATE_NOT_INSTALLED = 'NOT INSTALLED'
COLLECTOR_STATE_INSTALLED = 'INSTALLED'
COLLECTOR_STATE_CONFIGURED = 'CONFIGURED'
COLLECTOR_STATE_FAILURE = 'FAILURE'
COLLECTOR_STATE_PARTIAL_FAILURE = 'PARTIAL FAILURE'
COLLECTOR_STATES = [
    COLLECTOR_STATE_NOT_INSTALLED,
    COLLECTOR_STATE_INSTALLED,
    COLLECTOR_STATE_CONFIGURED,
    COLLECTOR_STATE_FAILURE,
    COLLECTOR_STATE_PARTIAL_FAILURE,
]

#: Service states
SERVICE_STATE_NOT_INSTALLED = 'NOT INSTALLED'
SERVICE_STATE_INSTALLED = 'INSTALLED'
SERVICE_STATE_CONFIGURED = 'CONFIGURED'
SERVICE_STATE_FAILURE = 'FAILURE'
SERVICE_STATE_REMOVED = 'REMOVED'
SERVICE_STATES = [
    SERVICE_STATE_NOT_INSTALLED,
    SERVICE_STATE_INSTALLED,
    SERVICE_STATE_CONFIGURED,
    SERVICE_STATE_FAILURE,
    SERVICE_STATE_REMOVED,
]

#: Service states included to status monitoring
SERVICE_MONITORING_STATES = [SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE]

#: Service type parameters.
#: ``main_service`` - is service required to collect votes
#: (*False* = support service);
#: **require_config** - does service require election config package for
#: operation;
#: ``require_tls`` - does service require TLS certificate and key to
#: ``tspreg`` - can communicate with TSP registration service
#: ``dds`` - can communicate with Mobile ID service
#: communicate with other services;
SERVICE_TYPE_PARAMS = {
    'backup': {
        'main_service': False,
        'require_config': False,
        'require_tls': False,
        'tspreg': False,
        'dds': False,
    },
    'choices': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'dds': True,
    },
    'dds': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'dds': True,
    },
    'log': {
        'main_service': False,
        'require_config': False,
        'require_tls': False,
        'tspreg': False,
        'dds': False,
    },
    'proxy': {
        'main_service': True,
        'require_config': True,
        'require_tls': False,
        'tspreg': False,
        'dds': False,
    },
    'storage': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'dds': False,
    },
    'verification': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'dds': False,
    },
    'voting': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': True,
        'dds': True,
    },
}

#: Service secret types
SERVICE_SECRET_TYPES = {
    'tls-cert': {
        'description': 'Service TLS certificate',
        'db-key': 'tls-cert',
        'target-path': '/var/lib/ivxv/service/{service_id}/tls.pem',
        'shared': False,
    },
    'tls-key': {
        'description': 'Service TLS key',
        'db-key': 'tls-key',
        'target-path': '/var/lib/ivxv/service/{service_id}/tls.key',
        'shared': False,
    },
    'dds-token-key': {
        'description': 'Mobile ID identity token',
        'db-key': 'dds-token-key',
        'target-path': '/var/lib/ivxv/service/ticket.key',
        'shared': True,
    },
    'tsp-regkey': {
        'description': 'PKIX TSP registration key',
        'db-key': 'tspreg-key',
        'target-path': '/var/lib/ivxv/service/{service_id}/tspreg.key',
        'shared': False,
    },
}

#: Filenames of collector deb packages
COLLECTOR_PKG_FILENAMES = {
    'ivxv-admin': f'ivxv-admin_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-backup': f'ivxv-backup_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-choices': f'ivxv-choices_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-common': f'ivxv-common_{DEB_PKG_VERSION}_all.deb',
    'ivxv-dds': f'ivxv-dds_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-log': f'ivxv-log_{DEB_PKG_VERSION}_all.deb',
    'ivxv-proxy': f'ivxv-proxy_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-storage': f'ivxv-storage_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-verification': f'ivxv-verification_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-voting': f'ivxv-voting_{DEB_PKG_VERSION}_amd64.deb',
}

#: Event log filename
EVENT_LOG_FILENAME = 'ivxv-management-events.log'

#: Event descriptions
EVENTS = {
    # collector state events
    'COLLECTOR_INIT': 'Initialize Collector',
    'COLLECTOR_RESET':
    f'Reset Collector (state: "{COLLECTOR_STATE_NOT_INSTALLED}")',
    'COLLECTOR_STATE_CHANGE':
    'Collector state changed from "{last_state}" to "{state}"',
    # command loading events
    'CMD_LOAD': 'Load command "{cmd_type}" version "{version}"',
    'CMD_LOADED': 'Command "{cmd_type}" is loaded, version "{version}"',
    'CMD_REMOVED': 'Command "{cmd_type}" is removed, version "{version}"',
    # user permission management events
    'PERMISSION_SET': 'Add permission "{permission}" to user "{user_cn}"',
    'PERMISSION_RESET': 'Reset user "{user_cn}" permissions',
    # election start/stop times registering
    'SET_ELECTION_TIME': 'Election "{period}" timestamp set to {timestamp}',
    # service management events
    'SERVICE_REGISTER':
    'Add {service_type} service (state: "%s")' % SERVICE_STATE_NOT_INSTALLED,
    'SERVICE_CONFIG_APPLY':
    'Applied {cfg_descr} version "{cfg_version}"',
    'SERVICE_STATE_CHANGE':
    'Service state changed from "{last_state}" to "{state}"',
    'SECRET_INSTALL': '{secret_descr} loaded to service',
}
