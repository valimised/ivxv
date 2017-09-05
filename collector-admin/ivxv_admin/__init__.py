# IVXV Internet voting framework
"""
Collector management service.
"""

# management daemon data
MANAGEMENT_DAEMON_PORT = 8080
MANAGEMENT_DAEMON_URL = 'http://localhost:%s/' % MANAGEMENT_DAEMON_PORT

# RFC3339 date format
RFC3339_DATE_FORMAT = '%Y-%m-%dT%H:%M:%S.%f'
# RFC3339 date format (without second fractions)
RFC3339_DATE_FORMAT_WO_FRACT = '%Y-%m-%dT%H:%M:%SZ'

# user permissions
PERMISSION_BALLOT_BOX_DOWNLOAD = 'download-ballot-box'
PERMISSION_ELECTION_CONF = 'election-conf-admin'
PERMISSION_LOG_VIEW = 'log-view'
PERMISSION_STATS_VIEW = 'stats-view'
PERMISSION_TECH_CONF = 'tech-conf-admin'
PERMISSION_USERS_ADMIN = 'user-admin'
# user roles
ROLES = {
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

# config types
CONFIG_TYPES = {
    'trust': 'trust root configuration',
    'election': 'elections configuration',
    'technical': 'collectors technical configuration',
}
# voting list types
VOTING_LIST_TYPES = {
    'choices': 'choices list',
    'voters': 'voters list',
}
# command types
COMMAND_TYPES = (list(CONFIG_TYPES.keys()) +
                 list(VOTING_LIST_TYPES.keys()) +
                 ['user'])

# collector states
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

# service states
SERVICE_STATE_NOT_INSTALLED = 'NOT INSTALLED'
SERVICE_STATE_INSTALLED = 'INSTALLED'
SERVICE_STATE_CONFIGURED = 'CONFIGURED'
SERVICE_STATE_FAILURE = 'FAILURE'
SERVICE_STATES = [
    SERVICE_STATE_NOT_INSTALLED,
    SERVICE_STATE_INSTALLED,
    SERVICE_STATE_CONFIGURED,
    SERVICE_STATE_FAILURE,
]

# service secret types
SERVICE_SECRET_TYPES = {
    'tls-cert': {
        'description': 'Service TLS certificate',
        'db-key': 'tls-cert',
        'affected-services': [
            'dds', 'choices', 'storage', 'verification', 'voting'],
        'target-path': '/var/lib/ivxv/service/{service_id}/tls.pem',
        'shared': False,
    },
    'tls-key': {
        'description': 'Service TLS key',
        'db-key': 'tls-key',
        'affected-services': [
            'dds', 'choices', 'storage', 'verification', 'voting'],
        'target-path': '/var/lib/ivxv/service/{service_id}/tls.key',
        'shared': False,
    },
    'dds-token-key': {
        'description': 'Mobile ID identity token',
        'db-key': 'dds-token-key',
        'affected-services': ['choices', 'dds', 'voting'],
        'target-path': '/var/lib/ivxv/service/ticket.key',
        'shared': True,
    },
    'tsp-regkey': {
        'description': 'PKIX TSP registration key',
        'db-key': 'tspreg-key',
        'affected-services': ['voting'],
        'target-path': '/var/lib/ivxv/service/{service_id}/tspreg.key',
        'shared': False,
    },
}
