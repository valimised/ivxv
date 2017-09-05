# IVXV Internet voting framework
"""
CLI utilities for IVXV Collector Management Service.
"""

import hashlib
import json
import logging
import logging.config
import os
import random
import re
import shutil

import OpenSSL

from jinja2 import Environment, PackageLoader

from . import (COMMAND_TYPES, CONFIG_TYPES,
               SERVICE_SECRET_TYPES, VOTING_LIST_TYPES)
from . import (COLLECTOR_STATE_CONFIGURED, COLLECTOR_STATE_FAILURE,
               COLLECTOR_STATE_INSTALLED, COLLECTOR_STATE_PARTIAL_FAILURE)
from . import (SERVICE_STATE_CONFIGURED, SERVICE_STATE_FAILURE,
               SERVICE_STATE_INSTALLED)
from . import lib
from .audit_log import register_event
from .cli import init_cli_util
from .command_file import (
    load_cmd_file_content, check_cmd_signature, load_collector_command_file)
from .config import CONFIG
from .db import IVXVManagerDb
from .service_config import Service, exec_remote_cmd

ACTIVE_CONFIG_FILES_PATH = CONFIG.get('active_config_files_path')
PERMISSIONS_PATH = CONFIG.get('permissions_path')

# JSON formatting options
JSON_DUMP_ARGS = dict(indent=2, sort_keys=True)

# Config value names for Management Service data directories
MANAGEMENT_PATH_PARAM_NAMES = [
    'ivxv_admin_data_path',
    'admin_ui_data_path',
    'permissions_path',
    'command_files_path',
    'active_config_files_path',
    'file_upload_path',
    'exported_votes_path',
    'deb_pkg_path',
    'ivxv_db_path',
]

# create logger
log = logging.getLogger(__name__)


def ivxv_create_data_dirs_util():
    """
    Create IVXV Collector Management Service data directories.

    NOTE: Directory owners and permissions are not set by this utility!

    Usage: ivxv-create-data-dirs
    """
    init_cli_util(ivxv_create_data_dirs_util.__doc__, allow_root=True)

    # config parameter names for directories
    for conf_var in MANAGEMENT_PATH_PARAM_NAMES:
        dirname = CONFIG.get(conf_var)
        if os.path.exists(dirname):
            log.info('Path "%s" already exist', dirname)
        else:
            log.info('Creating data directory "%s"', dirname)
            os.mkdir(dirname)
        if not os.path.isdir(dirname):
            log.error('Path "%s" is not a directory', dirname)
            return 1


def _check_db_directory():
    """Check database directory."""
    db_file_path = CONFIG.get('ivxv_db_file_path')
    db_path = os.path.dirname(db_file_path)
    if not os.path.exists(db_path):
        log.error('Database directory "%s" does not exist', db_path)
        return False
    return db_file_path


def ivxv_collector_init_util():
    """
    Initialize IVXV Collector.

    Usage: ivxv-collector-init [--force]

    Options:
        --force     Don't ask user confirmation
    """
    args = init_cli_util(ivxv_collector_init_util.__doc__)

    # ask confirmation
    if not args['--force']:
        if not lib.ask_user_confirmation(
                'Do You want to initialize IVXV Collector (Y=yes) ?'):
            return 1

    # initialize data directories
    for conf_var in MANAGEMENT_PATH_PARAM_NAMES:
        dirpath = CONFIG.get(conf_var)
        if not os.path.exists(dirpath):
            log.info('Creating directory "%s"', dirpath)
            os.mkdir(dirpath)
        if not os.path.isdir(dirpath):
            log.error('Path "%s" is not a directory', dirpath)
            return 1
        if conf_var not in ['ivxv_admin_data_path', 'deb_pkg_path',
                            'active_config_files_path']:
            lib.clean_directory(dirpath)
        for filename in os.listdir(CONFIG['active_config_files_path']):
            pattern = (r'(choices|election|technical|trust|voters-[0-9]{2})'
                       r'\.bdoc$')
            if re.match(pattern, filename):
                os.unlink(os.path.join(CONFIG['active_config_files_path'],
                                       filename))
    _init_management_datafiles()

    # initialize management database
    _init_management_database()


def database_util():
    """
    Add, remove or modify key/value pairs in
    IVXV Collector Management Service database.

    WARNING!
        Use this utility only for testing purposes!
        Never change database in production systems!

    Usage: ivxv-db <key> <value>
           ivxv-db --del <key>
    """
    args = init_cli_util(database_util.__doc__)

    # check database directory
    if not _check_db_directory():
        return 1

    # process record
    db = IVXVManagerDb(for_update=True)
    if args['--del']:
        db.rm_value(args['<key>'])
        log.info('Database value "%s" successfully removed', args['<key>'])
    else:
        db.set_value(args['<key>'], args['<value>'])
        log.info('Database value "%s" set to "%s"',
                 args['<key>'], args['<value>'])
    db.close()


def database_reset_util():
    """
    Reset IVXV Collector Management Service database.

    Usage: ivxv-db-reset [--force]

    Options:
        --force     Don't ask user confirmation
    """
    args = init_cli_util(database_reset_util.__doc__)

    # check database directory
    db_file_path = _check_db_directory()
    if not db_file_path:
        return 1

    # ask confirmation
    if not args['--force']:
        if not lib.ask_user_confirmation(
                'Do You want to reset IVXV management database (Y=yes) ?'):
            return 1

    # initialize database
    _init_management_database()

    # initialize data files
    _init_management_datafiles()

    # initialize data directories
    lib.clean_directory(CONFIG.get('file_upload_path'))


def _init_management_database():
    """Initialize management database."""
    log.debug('Initializing IVXV management database')

    db = IVXVManagerDb(open_db=False)
    db.reset()
    db.close()

    log.info('New management database is created with default values')
    register_event('INIT')


def _init_management_datafiles():
    """Initialize management data files."""
    shutil.copy(
        '/usr/lib/python3/dist-packages/ivxv_admin/templates/stats.json',
        CONFIG.get('admin_ui_data_path'))


def database_dump_util():
    """
    Dump IVXV Collector Management Service database.

    Usage: ivxv-db-dump
    """
    init_cli_util(database_dump_util.__doc__)

    log.info('Dumping IVXV management database')

    # check database directory
    if not _check_db_directory():
        return 1

    db = IVXVManagerDb()
    db.dump()
    db.close()


def users_list_util():
    """User listing utility for collector administrator interface."""
    init_cli_util("""
        List IVXV Collector Management Service registered users.

        Usage: ivxv-users-list
    """)

    # read users from database
    db = IVXVManagerDb()
    users = db.get_all_values('user')
    db.close()

    # output users
    user_no = 0
    for user_cn, permissions in sorted(users.items()):
        user_no += 1
        print('%d. %s   Permissions: %s' % (user_no, user_cn, permissions))

    if not user_no:
        print('No users defined')


def command_load_util():
    """
    Load command to IVXV Collector Management Service.

    Usage: ivxv-cmd-load [--validate-only] <type> FILE

    Options:
        <type>              Command type. Possible values are:
                            - election: election config
                            - technical: collector technical config
                            - trust: trust root config
                            - choices: choices list
                            - voters: voters list
                            - user: user account and role(s)
        --validate-only     Validate only.
    """
    args = init_cli_util(command_load_util.__doc__)

    # validate CLI arguments
    cmd_type = args['<type>'].lower()
    if cmd_type not in COMMAND_TYPES:
        log.error('Invalid command type "%s". Possible values are: %s',
                  cmd_type, ', '.join(COMMAND_TYPES))
        return 1
    conf_filename = args['FILE']
    conf_ext = conf_filename.split('.')[-1]

    # don't allow to load choices list more than once
    if cmd_type == 'choices':
        db = IVXVManagerDb()
        choices_list_version = db.get_value('list/choices')
        db.close()
        if choices_list_version:
            log.error('Choices list is already loaded (version: %s)',
                      choices_list_version)
            return 1
    # don't allow to load technical config if trust root config is not loaded
    if cmd_type == 'technical':
        db = IVXVManagerDb()
        trust_conf = db.get_value('config/trust')
        db.close()
        if not trust_conf:
            log.error(
                'Trust root must be loaded before technical configuration')
            return 1

    # check config permissions
    conf_signatures, all_signatures = None, None
    try:
        conf_signatures, all_signatures = check_cmd_signature(cmd_type,
                                                              args['FILE'])
    except FileNotFoundError as err:
        log.error('Binary for verifying container signature not found: %s',
                  err.strerror)
    except LookupError as err:
        log.error('Failed to verify config file signatures: %s', err)
    except OSError as err:
        log.error(err)

    if None in (conf_signatures, all_signatures):
        # failed to check command signature
        # Don't stop if only validating is requested
        if not args['--validate-only']:
            return 1
    else:
        for signature in all_signatures:
            log.info('Config file is signed by: %s', signature[2])
        if not conf_signatures:
            log.error('No signatures by authorized users')
            return 1
        config_version, role = conf_signatures[0]
        log.info('User %s with role "%s" is authorized to execute "%s" '
                 'commands', config_version.split(' ')[0], role, cmd_type)
        log.info('Using signature "%s" as config file version', config_version)
        config_timestamp = config_version.split(' ')[1]

    # load config
    log.info('Loading command %s from file %s', cmd_type, conf_filename)
    config_data = load_collector_command_file(cmd_type, args['FILE'])
    if config_data is None:
        return 1

    if args['--validate-only']:
        return

    # reset database on trust root config loading
    if cmd_type == 'trust':
        log.info('Resetting collector management database')
        db = IVXVManagerDb(open_db=False)
        db.reset()

    # connect to management service database
    db = IVXVManagerDb(for_update=True)
    active_config_path = None
    db_key = 'config' if cmd_type in CONFIG_TYPES else 'list'

    # detect order number for voters list
    if cmd_type == 'voters':
        voter_list_no = lib.detect_voters_list_order_no(db) + 1
        active_config_path = os.path.join(
            ACTIVE_CONFIG_FILES_PATH,
            '{}-{:02}.{}'.format(cmd_type, voter_list_no, conf_ext))
        db_key += '/{}{:02}'.format(cmd_type, voter_list_no)
    else:
        db_key += '/' + cmd_type

    # write config file to admin ui data path
    target_filepath = os.path.join(
        CONFIG.get('command_files_path'), '{}-{}.{}'.format(
            cmd_type, config_timestamp, conf_ext))
    log.debug('Copying file %s to %s', conf_filename, target_filepath)
    shutil.copy(conf_filename, target_filepath)
    log.info('%s config file loaded successfully', cmd_type.capitalize())

    # register config file version
    if cmd_type in CONFIG_TYPES.keys() or cmd_type in VOTING_LIST_TYPES.keys():
        db.set_value(db_key, config_version)
        if cmd_type == 'voters':
            db.set_value(db_key + '-loaded', '')

    # register user permissions
    if cmd_type == 'trust':  # initial permissions
        log.info('Resetting user permissions')
        for user_cn in db.get_all_values('user'):
            db.rm_value('user/' + user_cn)
        for user_cn in config_data['authorizations']:
            db.set_value('user/' + user_cn, 'admin')
        lib.populate_user_permissions(db)
    elif cmd_type == 'user':  # permissions from user management command
        user_cn = config_data['cn']
        log.info('Resetting user permissions')
        for existing_user_cn in db.get_all_values('user'):
            if existing_user_cn == user_cn:
                db.rm_value('user/' + user_cn)
        db.set_value('user/' + user_cn, ','.join(sorted(config_data['roles'])))

    # register election start/stop times
    if cmd_type == 'election':
        config_data = load_cmd_file_content(
            cmd_type, cmd_type + '.yaml', args['FILE'], CONFIG_TYPES[cmd_type])
        db.set_value('election/election-id', config_data['identifier'])
        for key in ['servicestart', 'electionstart',
                    'electionstop', 'servicestop']:
            db.set_value('election/' + key, config_data['period'][key])

    # create symlink to active config directory
    if cmd_type in CONFIG_TYPES.keys() or cmd_type in VOTING_LIST_TYPES.keys():
        if not active_config_path:
            active_config_path = os.path.join(
                ACTIVE_CONFIG_FILES_PATH,
                '{}.{}'.format(cmd_type, conf_ext))
        try:
            os.remove(active_config_path)
        except FileNotFoundError:
            pass
        os.symlink(target_filepath, active_config_path)

    # close database
    db.close()


def config_apply_util():
    """
    Apply IVXV collector config to IVXV services.

    Usage: ivxv-config-apply [--type=<type>] ... [<service-id>] ...

    Options:
        --type=<type>    Config type. Possible values are:
                         - election: election config file
                         - technical: collector technical config file
                         - choices: choices list
                         - voters: voters list
    """
    args = init_cli_util(config_apply_util.__doc__)
    # config types in applying order
    config_types_default = ['technical', 'election', 'choices', 'voters']

    # validate CLI arguments
    if not args['--type']:
        args['--type'] = config_types_default[:]
    for config_type in args['--type']:
        if config_type not in COMMAND_TYPES or config_type == 'trust':
            log.error('Invalid config type specified: --type=%s', config_type)
            return 1

    # connect to management service database
    db = IVXVManagerDb(for_update=True)

    # load technical config
    tech_config_version = db.get_value('config/technical')
    if not tech_config_version:
        log.error('Technical config is not loaded')
        return 1
    log.info('Technical config is signed by %s', tech_config_version)
    tech_conf_filename = os.path.join(ACTIVE_CONFIG_FILES_PATH,
                                      'technical.bdoc')
    tech_config = load_collector_command_file('technical', tech_conf_filename)
    if tech_config is None:
        return 1
    assert isinstance(tech_config, dict)

    # load election config
    election_config_version = db.get_value('config/election')
    if 'election' in args['--type']:
        if election_config_version:
            log.info(
                'Election config is signed by %s', election_config_version)
        else:
            log.info('Election config is not loaded')

    # generate database records for undefined services
    lib.gen_service_record_defaults(db, tech_config)
    lib.gen_host_record_defaults(db, tech_config)
    db.close()

    # generate service list from tech config
    services = lib.generate_service_list(tech_config['network'],
                                         args['<service-id>'])
    if services is None:
        return 1
    services = _prepare_service_list(services, args['<service-id>'])

    # apply config
    results = dict([i, {}] for i in args['--type'])
    for config_type in config_types_default:
        if config_type in args['--type']:
            # disallow applying if required services aren't operational
            if config_type in ['choices', 'storage']:
                list_services = lib.get_services(
                    include_types=['choices', 'storage'],
                    service_status=[SERVICE_STATE_CONFIGURED])
                if not list_services:
                    log.error(
                        'Choices and/or storage services are not configured')
                    return 1
            results[config_type] = _apply_config_to_services(
                services=services,
                config_type=config_type,
                tech_config=tech_config,
                tech_config_version=tech_config_version,
                election_config_version=election_config_version,
            )

    # output results
    results_summary = []
    for config_type in config_types_default:
        if results.get(config_type):
            log.info('Results for %s:', lib.config_type_verbose(config_type))
            for service_id, result in sorted(results[config_type].items()):
                log.info('  Service %s: %s',
                         service_id, 'success' if result else 'FAILED')
                results_summary.append(result)

    log.info('%d configuration packages successfully applied',
             results_summary.count(True))

    if False in results_summary:
        log.error('Failed to apply %d configuration packages',
                  results_summary.count(False))
        return 1


def _prepare_service_list(services, service_ids):
    """
    Prepare service list for applying config to services,
    return services that are in service_ids list.
    """
    services = [[service_type, service_data]
                for service_type, service_data in services
                if not service_ids or service_data['id'] in service_ids]

    # move log service record to beginning of list
    # to configure log collector(s) before other services
    for service_block in services[:]:
        if service_block[0] == 'log':
            services.remove(service_block)
            services.insert(0, service_block)

    return services


def _apply_config_to_services(services, config_type, tech_config,
                              tech_config_version, election_config_version):
    """
    Helper to apply config to services.

    :param services: Service list
    :type services: list
    :param config_type: Config types to apply
    :type config_type: str
    :param tech_config: Technical configuration from config file
    :type tech_config: dict
    :param services: List of services, grouped by service type
    :type services: list
    :param tech_config_version: Technical configuration version
    :type tech_config_version: str
    :param election_config_version: Election configuration version
    :type election_config_version: str

    :returns: dict with results for every service
    """
    # get list of registered services that require config update
    db = IVXVManagerDb()
    service_conf_status = lib.get_service_conf_status(db, tech_config)
    hostnames = sorted(set([service_data['address'].split(':')[0]
                            for _, service_data in services]))
    host_state = dict([
        [hostname, db.get_value('host/{}/state'.format(hostname))]
        for hostname in hostnames])
    db.close()

    # apply config to services
    def report_applying_result(config_success):
        """Report process result"""
        config_descr = (config_type + ' ' +
                        'config' if config_type in CONFIG_TYPES else 'list')
        results[service_id] = config_success
        if config_success:
            log.info('Service %s: %s config applied successfully',
                     service_id, config_descr)
        else:
            log.error('Service %s: failed to apply %s config',
                      service_id, config_descr)

    # apply config
    results = {}
    applied_lists = []
    for service_type, service_data in services:
        service_id = service_data['id']
        if not service_conf_status[service_id][config_type]:
            continue

        service_data['service_type'] = service_type  # used by Service()
        service = Service(service_id, service_data)

        # apply technical config
        if config_type == 'technical':
            # initialize service host if required
            if not host_state[service.hostname]:
                log.info('Initializing service host "%s"', service.hostname)
                success = service.init_service_host()
                if not success:
                    results[service_id] = False
                    continue
                db = IVXVManagerDb(for_update=True)
                db.set_value('host/{}/state'.format(service.hostname),
                             'REGISTERED')
                db.close()
                host_state[service.hostname] = 'REGISTERED'

            # apply config
            log.info('Service %s: Applying technical config', service_id)
            report_applying_result(
                service.apply_tech_conf(tech_config_version, tech_config))

        # apply election config
        if config_type == 'election' and service_type != 'log':
            log.info('Service %s: Applying elections config', service_id)
            report_applying_result(
                service.apply_election_conf(election_config_version))

        # apply choices list
        if (config_type == 'choices' and service_type == 'choices' and
                'choices' not in applied_lists):
            log.info('Service %s: Applying choices list', service_id)
            report_applying_result(service.apply_list('choices'))
            applied_lists.append('choices')

        # apply voters list
        if (config_type == 'voters' and service_type == 'voting' and
                'voters' not in applied_lists):
            for list_no in service_conf_status[service_id]['voters']:
                log.info('Service %s: Applying voters list #%d',
                         service_id, list_no)
                report_applying_result(service.apply_list('voters', list_no))
                applied_lists.append('voters')

    return results


def status_util():
    """
    Output IVXV collector state.

    Usage: ivxv-status [--json] [--filter=<filter-type> ...]
                       [--service=<service-id> ...]

    Options:
        --json                  Output full data in JSON format.
                                Note: filters have no effect in JSON output.
        --filter=<filter-type>  Filter output by section. Possible values are:
                                  * collector - collector status;
                                  * election - election data;
                                  * config - versions of loaded config;
                                  * list - versions of loades lists;
                                  * service - service information;
                                  * ext - external service information;
                                  * storage - storage information;
                                  * smart - output only relevant data;
                                  * all - output all data;
                                [Default: smart].
        --service=<service-id>  Filter output by service ID.
                                Note: This filter conflicts other section
                                filters than "smart" or "service".
    """
    args = init_cli_util(status_util.__doc__)
    filters = args['--filter']

    # validate CLI args
    data_sections = [
        'collector', 'election', 'config', 'list', 'service', 'ext', 'storage']
    for arg in filters:
        if arg not in data_sections + ['all', 'smart']:
            log.error('Invalid section: %s', arg)
            return 1
    if 'all' in filters:
        filters = data_sections
    if args['--service']:
        if filters != ['smart'] and filters != 'service':
            log.error('Service ID can only used with '
                      '"smart" or "service" section filters')
            return 1
        filters = ['service']
    if 'smart' in filters:
        filters = ['smart', 'collector', 'storage']

    # generate status data
    db = IVXVManagerDb()
    status = lib.generate_collector_status(db)
    db.close()

    # output JSON
    if args['--json']:
        print(json.dumps(status, indent=4, sort_keys=True))
        return

    # apply smart filtering
    if 'smart' in filters:
        filters = _config_status_filters(filters, status)

    # apply service ID filters
    if args['--service']:
        for network_id in list(status['network'].keys()):
            for service_id in list(status['network'][network_id].keys()):
                if service_id not in args['--service']:
                    del status['network'][network_id][service_id]
            if not status['network'][network_id]:
                del status['network'][network_id]

    # output plain text
    tmpl_env = Environment(loader=PackageLoader('ivxv_admin', 'templates'))
    template = tmpl_env.get_template('ivxv_status.jinja')
    output = template.render(status, sections=filters)
    output = re.sub(r'\n\n+', '\n\n', output)
    print(output.strip())


def _config_status_filters(filters, status):
    """Reconfigure list of status filters depending on actual status."""
    if (status['config']['trust'] or
            status['config']['technical'] or
            status['config']['election']):
        filters += ['config']
    if status['config']['election']:
        filters += ['election']
    if (status['list']['choices-loaded'] or
            status['list']['choices'] or
            status['list'].get('voters01') or
            status['list'].get('voters01-loaded')):
        filters += ['list']
    if status.get('service'):
        filters += ['service']
    return filters


def export_votes_util():
    """
    Export collected votes from IVXV voting service.

    Usage: ivxv-votes-export <output-file>
    """
    args = init_cli_util(export_votes_util.__doc__)

    services = lib.get_services(
        include_types=['voting'],
        require_collector_status=[COLLECTOR_STATE_CONFIGURED])
    if not services:
        return 1
    service_items = list(services.items())

    # get collected votes archive from voting service host
    while service_items:
        service_id, service_data = service_items.pop(
            random.randint(0, len(service_items) - 1))
        service = Service(service_id, service_data)
        if service.get_collected_votes(args['<output-file>']):
            break
    else:
        log.error('Failed to export collected votes')
        return 1

    log.info('Collected votes archive is written to %s',
             args['<output-file>'])


def import_secret_data_file_util():
    """
    Import secret data files to services.

    This utility loads file that contains secret data to services.

    Supported secret types are:

        tls-cert - TLS certificate for service.

                Certificate (and key) is used for securing
                communication between services and service instances.

        tls-key - TLS key for service.

                Key is used together with service certificate.

        tsp-regkey - PKIX TSP registration key for voting services.

                Key is used for signing Time Stamp Protocol requests.

                Key file must be in PEM format and
                must be not password protected.

        dds-token-key - Mobile ID identity token for
                        choices, dds and voting services.

                Key file must be 32 bytes long.

    Usage: ivxv-secret-import [--service=<service-id>] <secret-type> <keyfile>
    """
    # validate CLI arguments
    args = init_cli_util(import_secret_data_file_util.__doc__)
    secret_type = args['<secret-type>'].lower()
    filepath = args['<keyfile>']
    if secret_type not in SERVICE_SECRET_TYPES:
        log.error('Invalid secret type "%s"', secret_type)
        log.info('Supported secret types are: %s',
                 ', '.join(SERVICE_SECRET_TYPES.keys()))
        return 1
    secret_descr = SERVICE_SECRET_TYPES[secret_type]['description']

    # load file
    log.debug('Loading %s file', secret_descr)
    try:
        with open(filepath, 'rb') as fp:
            file_content = fp.read()
    except (FileNotFoundError, PermissionError) as err:
        log.error('Unable to load file "%s": %s', filepath, err.strerror)
        return 1

    # validate file
    if secret_type in ('tls-cert', 'tls-key') and not args['--service']:
        log.error('Service ID must be specified with %s', secret_descr)
        return 1
    if secret_type == 'dds-token-key' and len(file_content) != 32:
        log.error('%s file is not 32 bytes long (actual size: %d bytes)',
                  secret_descr, len(file_content))
        return 1
    if secret_type == 'tsp-regkey':
        try:
            OpenSSL.crypto.load_privatekey(OpenSSL.crypto.FILETYPE_PEM,
                                           file_content)
        except OpenSSL.crypto.Error:
            log.error('Unable to load %s from file "%s"',
                      secret_descr, filepath)
            return 1

    # calculate file checksum
    file_checksum = hashlib.sha256(file_content).hexdigest()

    # generate list of voting services that are in required state
    services_affected = SERVICE_SECRET_TYPES[secret_type]['affected-services']
    services = lib.get_services(
        include_types=services_affected,
        require_collector_status=[COLLECTOR_STATE_INSTALLED,
                                  COLLECTOR_STATE_CONFIGURED,
                                  COLLECTOR_STATE_FAILURE,
                                  COLLECTOR_STATE_PARTIAL_FAILURE],
        service_status=[SERVICE_STATE_INSTALLED,
                        SERVICE_STATE_CONFIGURED,
                        SERVICE_STATE_FAILURE])
    if not services:
        return 1

    # copy key to service hosts
    # FIXME avoid multiple copying of shared secret to single host
    services_updated = []
    for service_id, service_data in sorted(services.items()):
        if args['--service'] and service_id != args['--service']:
            continue
        if (service_data.get(SERVICE_SECRET_TYPES[secret_type]['db-key']) ==
                file_checksum):
            log.info('Service %s already contains specified %s',
                     service_id, secret_descr)
            continue

        service = Service(service_id, service_data)
        if not service.load_secret_file(secret_type, filepath, file_checksum):
            return 1
        services_updated.append(service_id)

    if args['--service'] and args['--service'] not in services_updated:
        log.error('%s was not imported to service %s',
                  secret_descr, args['--service'])
        return 1

    if args['--service']:
        log.info('%s is imported to service %s',
                 secret_descr, args['--service'])
    else:
        log.info('%s is imported to services', secret_descr)


def initialize_logmonitor_util():
    """
    Initialize IVXV Log Monitor.

    This utility exports collected IVXV log files from Log Collector
    Service to Log Monitor and initializes log analyzis in Log Monitor.

    Usage: ivxv-logmonitor-copy-log [--force]

    Options:
        --force     Don't ask user confirmation
    """
    # validate CLI arguments
    args = init_cli_util(initialize_logmonitor_util.__doc__)

    # ask confirmation
    if not args['--force']:
        if not lib.ask_user_confirmation(
                'Do You want to initialize IVXV Log Monitor (Y=yes) ?'):
            return 1

    # detect Log Monitor location
    config = load_collector_command_file(
        'technical',
        os.path.join(CONFIG['active_config_files_path'], 'technical.bdoc'))
    if config is None:
        log.error('Cannot load %s', CONFIG_TYPES['technical'])
        return 1
    logmonitor_config = config.get('logging')
    try:
        logmonitor_address = logmonitor_config[0]['address']
    except (TypeError, IndexError, KeyError):
        log.error('Log monitor config is not defined')
        return 1
    log.info('Using address "%s" for log monitor', logmonitor_address)
    logmon_account = 'logmon@{}'.format(logmonitor_address)

    # detect Log Collectors
    db = IVXVManagerDb()
    services = db.get_all_values('service')
    db.close()
    log_collectors = [[sid, service]
                      for sid, service in services.items()
                      if service['service-type'] == 'log']
    try:
        log_collector = random.choice(log_collectors)
    except IndexError:
        log.error(
            'Log Collector service is not defined in technical configuration')
        return 1
    service = Service(log_collector[0], log_collector[1])
    log.info('Using log collector service %s', service.service_id)

    # checking access to Log Monitor account
    cmd = ['ssh', logmon_account, 'true']
    proc = exec_remote_cmd(cmd)
    if proc.returncode:
        log.error('Cannot access to Log Monitor (%s)', logmon_account)
        return 1

    # export log file from Log Collector to Log Monitor
    source_path = '{}@{}:/var/log/ivxv.log'.format(
        service.service_account_name, service.hostname)
    target_path = 'logmon@{}:/var/log/ivxv-log/ivxv.log'.format(
        logmonitor_address)
    cmd = ['scp', '-3', '-C', source_path, target_path]
    log.info('Copying collected log file '
             'from Log Collector Service to Log Monitor')
    proc = exec_remote_cmd(cmd)
    if proc.returncode:
        log.error('Failed to copy log file')
        return 1
