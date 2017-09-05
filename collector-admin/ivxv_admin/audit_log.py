# IVXV Internet voting framework
"""
Audit log module for collector management service.
"""

import logging

# create logger
log = logging.getLogger('ivxv_audit_log')


AUDIT_EVENTS = {
    # collector status changes
    'INIT': 'Collector service initialized',
    'COLLECTOR_STATUS': 'Collector status changed to {status}',

    # service status changes
    'SERVICE_STATUS': 'Service {service_id} status changed to {status}',

    # config management
    'CONFIG_TRUST_LOAD': 'Trust root config loaded',
    'CONFIG_TECH_LOAD':
    'Collector technical config loaded (version: {version})',
    'CONFIG_ELECTION_LOAD': 'Election config loaded (version: {version})',
    'LIST_CHOICES_LOAD': 'Choices list loaded (version: {version})',
    'LIST_VOTERS_LOAD':
    'Voters list loaded (list number: {list_no}, version: {version})',

    # user management
    'USER_ADD': 'User {user_cn} added (by {admin_cn})',
    'USER_ROLE_ADD': 'Role {role} added to user {user_cn} (by {admin_cn})',
    'USER_ROLE_REMOVE':
    'Role {role} removed from user {user_cn} (by {admin_cn})',

    # election
    'ELECTION_SERVICE_START': 'Collector service started to serve election',
    'ELECTION_SERVICE_STOP': 'Collector service stopped to serve election',
    'ELECTION_START': 'Election period started',
    'ELECTION_STOP': 'Election period stopped',
}


def register_event(event_id, **kw):
    """Register event in audit log."""
    event = AUDIT_EVENTS[event_id].format(**kw)
    log.info(event)
