# IVXV Internet voting framework
"""Collector Management Service."""

__version__ = '1.8.2'
DEB_PKG_VERSION = '1.8.2'

#: Management daemon data
MANAGEMENT_DAEMON_PORT = 8080
MANAGEMENT_DAEMON_URL = f"http://localhost:{MANAGEMENT_DAEMON_PORT}/"

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
#: ``mobile_id`` - can communicate with Mobile ID service
#: communicate with other services;
SERVICE_TYPE_PARAMS = {
    'backup': {
        'main_service': False,
        'require_config': False,
        'require_tls': False,
        'tspreg': False,
        'mobile_id': False,
    },
    'choices': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': True,
    },
    'log': {
        'main_service': False,
        'require_config': False,
        'require_tls': False,
        'tspreg': False,
        'mobile_id': False,
    },
    'mid': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': True,
    },
    'votesorder': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': False,
    },
    'proxy': {
        'main_service': True,
        'require_config': True,
        'require_tls': False,
        'tspreg': False,
        'mobile_id': False,
    },
    'smartid': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': True,
    },
    'storage': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': False,
    },
    'verification': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': False,
        'mobile_id': False,
    },
    'voting': {
        'main_service': True,
        'require_config': True,
        'require_tls': True,
        'tspreg': True,
        'mobile_id': True,
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
    'mid-token-key': {
        'description': 'Mobile-ID/Smart-ID identity token',
        'db-key': 'mid-token-key',
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
    'ivxv-log': f'ivxv-log_{DEB_PKG_VERSION}_all.deb',
    'ivxv-mid': f'ivxv-mid_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-votesorder': f'ivxv-votesorder_{DEB_PKG_VERSION}_amd64.deb',
    'ivxv-smartid': f'ivxv-smartid_{DEB_PKG_VERSION}_amd64.deb',
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
    "COLLECTOR_INIT": "Initialize Collector",
    "COLLECTOR_RESET":
    f"Reset Collector (state: {COLLECTOR_STATE_NOT_INSTALLED!r})",
    "COLLECTOR_STATE_CHANGE":
    "Collector state changed from {last_state!r} to {state!r}",
    # command loading events
    "CMD_LOAD": "Load command {cmd_type!r} version {version!r}",
    "CMD_LOADED": "Command {cmd_type!r} is loaded, version {version!r}",
    "CMD_REMOVED": "Command {cmd_type!r} is removed, version {version!r}",
    # voter list downloading events
    "VOTER_LIST_DOWNLOADED": "Downloaded voter list changeset #{changeset_no}",
    "VOTER_LIST_DOWNLOAD_FAILED":
    "Failed to download voter list changeset #{changeset_no}",
    # user permission management events
    "PERMISSION_SET": "Add permission {permission!r} to user {user_cn!r}",
    "PERMISSION_RESET": "Reset user {user_cn!r} permissions",
    # election start/stop times registering
    "SET_ELECTION_TIME": "Election {period!r} timestamp set to {timestamp}",
    # service management events
    "SERVICE_REGISTER":
    f"Add {{service_type}} service (state: {SERVICE_STATE_NOT_INSTALLED!r})",
    "SERVICE_CONFIG_APPLY": 'Applied {cfg_descr} version {cfg_version!r}',
    "SERVICE_STATE_CHANGE": 'Service state changed from {last_state!r} to {state!r}',
    "SECRET_INSTALL": "{secret_descr} loaded to service",
}
