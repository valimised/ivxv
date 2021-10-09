# IVXV Internet voting framework
"""CLI utilities for config applying."""

import json
import os
import shutil
import sys

from ... import (CFG_TYPES, CMD_TYPES, SERVICE_STATE_CONFIGURED,
                 SERVICE_TYPE_PARAMS, lib)
from ...command_file import load_collector_cmd_file
from ...config import cfg_path
from ...db import IVXVManagerDb
from ...lib.lockfile import PidLocker
from ...service import generate_service_list, get_service_cfg_state
from ...service.service import Service
from .. import init_cli_util, log

#: config types in applying order
CFG_TYPES_DEFAULT = ("technical", "election", "choices", "districts", "voters")


def main():
    """Apply loaded config to IVXV services."""
    args = init_cli_util("""
    Apply loaded IVXV Collector config to services.

    Usage: ivxv-config-apply [--type=<type>] ... [<service-id>] ...

    Options:
        --type=<type>    Config type. Possible values are:
                         - election: election config file
                         - technical: collector technical config file
                         - choices: choices list
                         - districts: districts list
                         - voters: voters list
    """)

    # validate CLI arguments
    for cfg_type in args['--type']:
        if cfg_type not in CMD_TYPES or cfg_type == 'trust':
            log.error('Invalid config type specified: --type=%s', cfg_type)
            return 1

    cfg_types = args['--type'] or list(CFG_TYPES_DEFAULT)
    service_ids = args['<service-id>']

    # load config data
    with IVXVManagerDb() as db:
        cfg_data = get_cfg_data(db)
    try:
        tech_cfg = load_tech_cfg(
            cfg_data['technical']['version'],
            cfg_data['election']['version'],
            'election' in cfg_types)
    except lib.IvxvError as err:
        log.error(err)
        return 1

    # remove choices/voters list version record from config data
    # if list is not loaded and also not specified in CLI args
    if not args['--type']:
        if not cfg_data.get('choices', {}).get('version'):
            cfg_types.remove('choices')
        if not cfg_data.get('districts', {}).get('version'):
            cfg_types.remove('districts')
        if not cfg_data.get('voters0000', {}).get('version'):
            cfg_types.remove('voters')

    # generate service list from tech config
    services = generate_service_list(tech_cfg['network'], service_ids)
    if services is None:
        return 1
    services = prepare_service_list(services, service_ids)

    # create pidfile to avoid running of multiple instances
    pidfile_path = cfg_path('ivxv_admin_data_path', 'ivxv-config-apply.pid')
    log.debug("Creating pidfile %r", pidfile_path)
    try:
        PidLocker(pidfile_path)
    except IOError:
        log.error("Creating pidfile %r failed. Is another %s running?",
                  pidfile_path, os.path.basename(sys.argv[0]))
        sys.exit(1)

    # apply config
    apply_result = apply_cfg(
        cfg_types=cfg_types,
        services=services,
        tech_cfg=tech_cfg,
        cfg_data=cfg_data)
    return int(not apply_result)


def get_cfg_data(db):
    """Load config versions from database.

    :rtype: dict
    """
    # get config versions
    cfg_data = {
        'technical': {
            'version': db.get_value('config/technical') or None
        },
        'election': {
            'version': db.get_value('config/election') or None
        },
        'choices': {
            'version': db.get_value('list/choices') or None
        },
        'districts': {
            'version': db.get_value('list/districts') or None
        },
    }
    if not cfg_data['choices']['version']:
        del cfg_data['choices']
    if not cfg_data['districts']['version']:
        del cfg_data['districts']
    try:
        for changeset_no in range(10_000):
            cfg_data[f"voters{changeset_no:04}"] = {
                "version": db.get_value(f"list/voters{changeset_no:04}") or None
            }
    except KeyError:
        pass

    # get config applying report filenames
    for cmd_type in cfg_data:
        cfg_filepath = lib.get_loaded_cfg_file_path(cmd_type)
        cfg_data[cmd_type]["state_file"] = (
            cfg_filepath.replace(os.path.splitext(cfg_filepath)[1], ".json")
            if cfg_filepath
            else None
        )

    return cfg_data


def load_tech_cfg(tech_cfg_ver, election_cfg_ver, has_election_cfg):
    """Load technical config.

    * Load technical config
    * Log election config state
    * Generate database records for undefined services

    :return: technical config
    :rtype: dicts

    :raises IvxvError: if technical config cannot be loaded
    """
    # load technical config
    if not tech_cfg_ver:
        raise lib.IvxvError(
            'Technical config is not loaded to management service')
    log.info('Technical config is signed by %s', tech_cfg_ver)
    tech_cfg = load_collector_cmd_file(
        'technical', cfg_path('active_config_files_path', 'technical.bdoc'))
    if tech_cfg is None:
        raise lib.IvxvError('Error while loading technical config')

    # log election config state
    if has_election_cfg:
        if election_cfg_ver:
            log.info('Election config is signed by %s', election_cfg_ver)
        else:
            log.info('Election config is not loaded to management service')

    return tech_cfg


def prepare_service_list(services, service_ids):
    """Prepare service list for applying config to services.

    Return services that are in service_ids list.
    """
    services = [[service_type, service_data]
                for service_type, service_data in services
                if not service_ids or service_data['id'] in service_ids]

    # reorder services to satisfy dependencies
    for service_block in services[:]:
        # move log service record to beginning of list
        # to configure log collector(s) before other services
        if service_block[0] == 'log':
            services.remove(service_block)
            services.insert(0, service_block)

        # move proxy service records to end of list to ensure proxied services
        # are started and initial health checks can succeed
        if service_block[0] == 'proxy':
            services.remove(service_block)
            services.append(service_block)

    return services


def apply_cfg(cfg_types, services, tech_cfg, cfg_data):
    """Apply config to services."""
    results = dict([i, {}] for i in cfg_types)
    for cfg_type in CFG_TYPES_DEFAULT:
        if cfg_type not in cfg_types:
            continue
        # disallow applying choices/districts/voters list
        # if required services aren't operational
        if cfg_type in ["choices", "districts", "voters"]:
            list_services = lib.get_services(
                include_types=['choices', 'storage'],
                service_state=[SERVICE_STATE_CONFIGURED])
            if not list_services:
                log.error('Choices and/or storage services are not configured')
                return False
        result_by_service = apply_cfg_to_services(
            services=services,
            cfg_type=cfg_type,
            tech_cfg=tech_cfg,
            cfg_data=cfg_data,
        )
        results[cfg_type] = result_by_service

        # update config state file
        if (cfg_type in CFG_TYPES and result_by_service
                and False not in result_by_service.values()):
            state_filepath = cfg_data[cfg_type]['state_file']
            with open(state_filepath) as fp:
                cfg_state = json.load(fp)
            cfg_state['completed'] = True

            tmp_filepath = f'{state_filepath}.tmp'
            with open(tmp_filepath, 'x') as fp:
                json.dump(cfg_state, fp, indent=4, sort_keys=True)
            shutil.move(tmp_filepath, state_filepath)

    return aggregate_results(results)


def apply_cfg_to_services(services, cfg_type, tech_cfg, cfg_data):
    """Helper to apply config to services.

    :param services: Service list
    :type services: list
    :param cfg_type: Config type to apply
    :type cfg_type: str
    :param tech_cfg: Technical configuration from config file
    :type tech_cfg: dict
    :param services: List of services, grouped by service type
    :type services: list
    :param cfg_data: Loaded config/list files data (version, filenames)
    :type cfg_data: dict

    :return: dict with results for every service
    """
    # get list of registered services that require config update
    with IVXVManagerDb() as db:
        service_cfg_state = get_service_cfg_state(db, tech_cfg)
        hostnames = sorted(
            set(
                service_data['address'].split(':')[0]
                for _, service_data in services
            ))
        host_state = dict(
            [hostname, db.get_value(f'host/{hostname}/state')]
            for hostname in hostnames)

    # apply config to services
    def report_applying_result():
        """Report process result."""
        cfg_descr = '{} {}'.format(
            cfg_type, 'config' if cfg_type in CFG_TYPES else 'list')
        results[service_id] = cfg_success
        if cfg_success:
            log.info('Service %s: %s config applied successfully',
                     service_id, cfg_descr)
            service.register_event(
                'SERVICE_CONFIG_APPLY',
                params={'cfg_descr': cfg_descr, 'cfg_version': cfg_version})
        else:
            log.error('Service %s: failed to apply %s', service_id, cfg_descr)

    # apply config
    results = {}
    applied_lists = []
    attempt_no = 0
    for service_type, service_data in services:
        service_id = service_data['id']
        if not service_cfg_state[service_id][cfg_type]:
            continue

        service_data['service_type'] = service_type  # used by Service()
        service = Service(service_id, service_data)

        # apply technical config
        if cfg_type == 'technical':
            # initialize service host if required
            if not host_state[service.hostname]:
                log.info("Initializing service host %r", service.hostname)
                success = service.init_service_host()
                if not success:
                    results[service_id] = False
                    continue
                with IVXVManagerDb(for_update=True) as db:
                    db.set_value(f'host/{service.hostname}/state',
                                 'REGISTERED')
                host_state[service.hostname] = 'REGISTERED'

            # apply config
            log.info('Service %s: Applying technical config', service_id)
            cfg_version = cfg_data[cfg_type]['version']
            attempt_no = service.load_apply_state(
                cfg_data[cfg_type]['state_file'], attempt_no)
            cfg_success = service.apply_tech_cfg(cfg_version, tech_cfg)
            report_applying_result()
            service.update_apply_state()

        # apply election config
        elif (cfg_type == 'election'
              and SERVICE_TYPE_PARAMS[service_type]['require_config']):
            log.info('Service %s: Applying elections config', service_id)
            cfg_version = cfg_data[cfg_type]['version']
            attempt_no = service.load_apply_state(
                cfg_data[cfg_type]['state_file'], attempt_no)
            cfg_success = service.apply_election_cfg(cfg_version)
            report_applying_result()
            service.update_apply_state()

        # apply choices/districts list
        elif (cfg_type in ["choices", "districts"] and service_type == "choices"
              and cfg_type not in applied_lists):
            log.info("Service %s: Applying %s list", service_id, cfg_type)
            cfg_version = cfg_data[cfg_type]['version']
            attempt_no = service.load_apply_state(
                cfg_data[cfg_type]['state_file'], attempt_no)
            cfg_success = service.apply_list(cfg_type)
            report_applying_result()
            service.update_apply_state(completed=True)
            applied_lists.append(cfg_type)

        # apply voters lists
        elif (cfg_type == 'voters' and service_type == 'voting'
              and 'voters' not in applied_lists):
            for changeset_no in service_cfg_state[service_id]["voters"]:
                log.info(
                    "Service %s: Applying voter list changeset #%d",
                    service_id,
                    changeset_no,
                )
                voters_list_id = f"voters{changeset_no:04}"
                cfg_version = cfg_data[voters_list_id]['version']
                attempt_no = service.load_apply_state(
                    cfg_data[voters_list_id]['state_file'])
                cfg_success = service.apply_list(cfg_type, changeset_no)
                report_applying_result()
                service.update_apply_state(completed=True)
                applied_lists.append('voters')

    return results


def aggregate_results(results):
    """Aggregate and output collector config applying results for services.

    :rtype: bool
    :return: True if no error detected
    """
    results_summary = []
    for cfg_type in CFG_TYPES_DEFAULT:
        if results.get(cfg_type):
            log.info('Results for %s:', lib.cfg_type_verbose(cfg_type))
            for service_id, result in sorted(results[cfg_type].items()):
                log.info('  Service %s: %s',
                         service_id, 'success' if result else 'FAILED')
                results_summary.append(result)

    log.info('%d configuration packages successfully applied',
             results_summary.count(True))

    if False in results_summary:
        log.error('Failed to apply %d configuration packages',
                  results_summary.count(False))

    return False not in results_summary
