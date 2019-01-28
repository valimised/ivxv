# IVXV Internet voting framework
"""Microservice management module."""

from .. import SERVICE_TYPE_PARAMS, SERVICE_STATE_REMOVED
from .logging import log

#: Path to users SSH public key file.
IVXV_ADMIN_SSH_PUBKEY_FILE = '~/.ssh/id_ed25519.pub'
#: Path to rsyslog config filename.
RSYSLOG_CFG_FILENAME = '/etc/rsyslog.d/ivxv-logging.conf'


def get_service_cfg_state(db, cfg):
    """Get list of services that requires config update.

    Check technical config records against database and generate list of
    services that require config update.

    :param db: Database handler
    :param cfg: Technical config

    :return: Service list with config update params:

             .. code-block:: text

                 {
                     'service-id': {
                         'technical': bool,
                         'election': bool,
                         'choices': bool,
                         'voters': [list-number, ...],
                     },
                     ...
                 }

    :rtype: dict

    """
    assert 'network' in cfg

    # detect config versions
    tech_cfg_ver = db.get_value('config/technical')
    election_cfg_ver = db.get_value('config/election')

    # detect need for list updates
    update_choices_list = (db.get_value('list/choices') !=
                           db.get_value('list/choices-loaded'))
    update_voters_list = []
    for voter_list_no in range(1, 100):
        try:
            if (db.get_value('list/voters%02d' % voter_list_no) !=
                    db.get_value('list/voters%02d-loaded' % voter_list_no)):
                update_voters_list.append(voter_list_no)
        except KeyError:
            break

    # create list of services
    service_list = {}
    for network in cfg['network']:
        for service_type, services in sorted(network['services'].items()):
            for service in services or []:
                db_key_prefix = f'service/{service["id"]}'

                service_tech_cfg_ver = db.get_value(
                    db_key_prefix + '/technical-conf-version')
                service_election_cfg_ver = (
                    election_cfg_ver and
                    db.get_value(db_key_prefix + '/election-conf-version'))

                service_list[service['id']] = {
                    'technical': service_tech_cfg_ver != tech_cfg_ver,
                    'election': service_election_cfg_ver != election_cfg_ver,
                    'choices':
                    service_type == 'choices' and update_choices_list,
                    'voters': update_voters_list,
                }
                if service_list[service['id']]['technical']:
                    log.debug('Service %s have no latest version of '
                              'technical config', service['id'])

                if service_list[service['id']]['election']:
                    log.debug('Service %s have no latest version of '
                              'election config', service['id'])

    return service_list


def generate_service_list(service_networks, service_id_filter=None):
    """Generate service list from collector technical config.

    .. code-block:: text

        [
            [<service_type>, {param: value, ...}],
            ...
        ]

    :param service_networks: Service networks from collector technical config
                             (section 'networks').
    :type service_networks: dict
    :param service_id_filter: Service IDs to include.
                              All services are included if None.
    :type service_id_filter: list

    :return: List of services or None if some of specified services not found.
    """
    service_id_filter = service_id_filter or []
    service_list = []
    service_ids = service_id_filter[:]
    for network in service_networks:
        for service_type, services in sorted(network['services'].items()):
            for service_data in services or []:
                service_id = service_data['id']
                if service_id_filter and service_id not in service_ids:
                    log.debug('Skipping service %s', service_id)
                    continue
                if service_ids:
                    service_ids.remove(service_id)
                service_list.append([service_type, service_data])

    if service_ids:
        log.error('Invalid service ID: %s', ', '.join(service_ids))
        return None

    return service_list


def generate_service_hints(services):
    """Generate configuration hints for services and inject it to service data.

    :param services: Service data
    :type services: dict
    """
    for service_id, params in services.items():
        if params['state'] == SERVICE_STATE_REMOVED:
            continue
        # ordered list of hints
        hints = [
            ['Apply technical config', not params['technical-conf-version']],
        ]
        service_type_params = SERVICE_TYPE_PARAMS[params['service-type']]
        if service_type_params['require_tls']:
            hints += [
                ['Install service TLS key', not params.get('tls-key', True)],
                ['Install service TLS certificate',
                 not params.get('tls-cert', True)],
            ]
        if service_type_params['dds']:
            hints.append(
                ['Install mobile ID identity token key',
                 not params.get('dds-token-key', True)])
        if service_type_params['tspreg']:
            hints.append(
                ['Install TSP registration key',
                 not params.get('tspreg-key', True)])
        if service_type_params['require_config']:
            hints.append(
                ['Apply election config', not params['election-conf-version']])

        for hint, is_relevant in hints:
            if is_relevant:
                services[service_id]['bg_info'] = hint
                break
