# IVXV Internet voting framework
"""
Command file loading module collector management service.
"""

import datetime
import json
import logging
import os
import re
import subprocess
import tempfile
import zipfile

import schematics.exceptions
import yaml

from . import (CFG_TYPES, CMD_DESCR, CMD_TYPES, PERMISSION_ELECTION_CONF,
               PERMISSION_TECH_CONF, PERMISSION_USERS_ADMIN,
               RFC3339_DATE_FORMAT_WO_FRACT, USER_ROLES, VOTING_LIST_TYPES)
from .config import CONFIG
from .config_validator.validator_util import validate_cfg
from .db import IVXVManagerDb
from .lib import IvxvError

# create logger
log = logging.getLogger(__name__)


def load_collector_cmd_file(cmd_type, filename, plain=False):
    """Load and validate collector command file.

    :param plain: Load plain file or BDOC container.
    :type plain: bool
    :return: Config as data structure or None on error
    :rtype: dict
    """
    assert cmd_type in CMD_TYPES, f"Invalid command type {cmd_type!r}"

    file_descr = CMD_DESCR[cmd_type]
    if plain:
        cmd_filename = filename
        filename = None
    elif cmd_type in CFG_TYPES:
        cmd_filename = re.compile(rf"(.+\.)?{cmd_type}.yaml$")
    elif cmd_type in VOTING_LIST_TYPES:
        cmd_filename = None
    else:
        assert cmd_type == 'user'
        cmd_filename = 'user.json'

    # load content from file
    cfg = None
    try:
        cfg = load_cfg_file_content(cmd_type, cmd_filename, filename)
    except (IvxvError, OSError) as err:
        log.error('Error while loading config file: %s', err)
    except UnicodeDecodeError as err:
        log.error('Error while decoding config file: %s', err)
    except json.decoder.JSONDecodeError as err:
        log.error('JSON error in config file: %s', err)
    if cfg is None:
        return None

    # validate config file
    log.info('Validating %s', file_descr)
    try:
        cfg = validate_cfg(cfg, cmd_type)
    except ValueError as err:
        log.error(err)
        return None
    except schematics.exceptions.DataError as data_error:
        log_cfg_validation_errors(data_error.errors)
        return None

    # validate election ID in config
    if cmd_type not in ['technical', 'trust', 'user']:
        with IVXVManagerDb() as db:
            election_id = db.get_value('election/election-id')
            if election_id:
                if cmd_type == 'election':
                    cfg_election_id = cfg.get('identifier')
                elif cmd_type in ['choices', 'districts', 'voters']:
                    cfg_election_id = cfg.get('election')
                else:
                    raise NotImplementedError(f"Unknown config type {cmd_type}")
                if election_id != cfg_election_id:
                    log.error(
                        "Election ID %r in config file does not match "
                        "with current election ID %r",
                        cfg_election_id, election_id)
                    return None

    if plain:
        log.info('%s is valid', file_descr.capitalize())
    else:
        log.info('Files in %s package are valid', file_descr)

    return cfg


def load_cfg_file_content(cmd_type, cfg_filename, cmd_filename):
    """Load config file content from command file container.

    :param cfg_filename: Config file name
    :type cfg_filename: str or regexp pattern
    :param cmd_filename: Container file name.
                          Plain config file is used if None.
    :type cmd_filename: str
    :return: Config as data structure; string for voter list; or None on error
    :rtype: dict

    :raises UnicodeDecodeError: if config file is not in UTF-8
    :raises OSError: if some file operation fails
    :raises json.decoder.JSONDecodeError: if JSON decoding fails
    :raises IvxvError: if some internal check fails
    """
    file_descr = CMD_DESCR[cmd_type]

    # load plain config file
    if cmd_filename is None:
        return load_plain_cfg_file(
            cmd_type, os.path.dirname(cfg_filename),
            os.path.basename(cfg_filename))

    # load BDOC command file
    log.info('Loading command file %r (%s)', cmd_filename, file_descr)
    try:
        cmd_file = zipfile.ZipFile(cmd_filename)
    except FileNotFoundError:
        raise IvxvError(f"File {cmd_filename!r} not found")
    except zipfile.BadZipFile:
        raise IvxvError(f"File {cmd_filename!r} is not a valid ZIP container")
    log.debug('Command file loaded')
    log.debug("Files in command file container: %r", cmd_file.namelist())

    # parse contents of command file
    filenames_in_bdoc = cmd_file.namelist()
    for filename in filenames_in_bdoc[:]:
        if filename in [
                'META-INF/manifest.xml', 'META-INF/signatures0.xml',
                'mimetype'
        ]:
            filenames_in_bdoc.remove(filename)
    if cfg_filename is None:
        cfg_filename = get_command_filename(cmd_type, filenames_in_bdoc)
        if cfg_filename is None:
            raise IvxvError("No command filename detected")
        log.debug("Detected command filename %r", cfg_filename)

    if isinstance(cfg_filename, str) and cfg_filename not in filenames_in_bdoc:
        raise IvxvError(
            f'Command file does not contain expected file {cfg_filename} '
            f'for {file_descr}')
    if isinstance(cfg_filename, type(re.compile(''))):
        for filename in filenames_in_bdoc:
            if cfg_filename.match(filename):
                cfg_filename = filename
                break
        else:
            raise IvxvError(
                f'Command file does not contain expected '
                f'file {cfg_filename.pattern} for {file_descr}')

    # create temporary directory to extract container
    with tempfile.TemporaryDirectory() as dirpath:
        log.debug("Extracting command file to directory %r", dirpath)
        cmd_file.extractall(dirpath)
        log.debug('Command file successfully extracted')
        cfg = load_plain_cfg_file(cmd_type, dirpath, cfg_filename)

    return cfg


def ck_json_dict_key_uniqueness(args):
    """Json loader hook to check dict key uniqueness.

    Add strictness for json loader to detect JSON that violates dict key
    uniqueness.
    """
    keys = [_[0] for _ in args]
    for key in keys:
        if keys.count(key) > 1:
            raise IvxvError(f"Duplicate key {key!r} in JSON object")
    return dict(args)


def load_plain_cfg_file(cmd_type, dirpath, filename):
    """Load content from plain config file."""
    log.debug("Reading file %r", os.path.join(dirpath, filename))
    cwd = os.getcwd()
    if dirpath:
        os.chdir(dirpath)
    try:
        with open(filename, 'rb') as fp:
            # possible UnicodeDecodeError must be handled by caller
            file_content = fp.read().decode('utf-8')
            if cmd_type in CFG_TYPES:
                cfg = load_yaml_file(file_content)
            elif cmd_type in ('choices', 'districts', 'user'):
                cfg = json.loads(
                    file_content, object_pairs_hook=ck_json_dict_key_uniqueness)
            else:
                assert cmd_type == 'voters'
                if filename.endswith(".skip.yaml"):
                    cfg = load_yaml_file(file_content)
                else:
                    cfg = file_content
    finally:
        if dirpath:
            os.chdir(cwd)

    return cfg


def get_command_filename(cmd_type, filenames_in_bdoc):
    """Get command file name from list of filenames.

    Get command file for choice list, district list,
    voter list or voter list changeset skip command.

    :return: Command filename or None on error
    :rtype: str
    """
    filename = None

    if cmd_type in ('choices', 'districts'):
        if len(filenames_in_bdoc) > 1:
            log.error(
                "Too many files in %s file: %r",
                CMD_DESCR[cmd_type], filenames_in_bdoc)
        elif not filenames_in_bdoc:
            log.error('Missing %s list in %s file',
                      cmd_type, CMD_DESCR[cmd_type])
        else:
            filename = filenames_in_bdoc[0]

    elif cmd_type == 'voters':
        if len(filenames_in_bdoc) > 2:
            log.error("Too many files in %s file: %r",
                      CMD_DESCR[cmd_type], filenames_in_bdoc)
        elif (
            len(filenames_in_bdoc) == 1 and filenames_in_bdoc[0].endswith(".skip.yaml")
        ):
            filename = filenames_in_bdoc[0]
        elif len(filenames_in_bdoc) < 2:
            log.error('Missing voters list or signature in %s file',
                      CMD_DESCR[cmd_type])
        else:
            sig_filename, utf_filename = sorted(filenames_in_bdoc)
            if (utf_filename.endswith('.utf')
                    and sig_filename == utf_filename[:-4] + '.sig'):
                filename = utf_filename
            else:
                log.error('Voters list and signature file names do not '
                          'match (filenames: %s)',
                          ', '.join([utf_filename, sig_filename]))
                log.error('List file must have ".utf" extension and '
                          'signature file the same base with ".sig" extension')
    else:
        raise NotImplementedError

    return filename


def check_cmd_signature(cmd_type, filename):
    """
    Check collector command file signature.

    :return: Two lists, first one contains signatures by authorized users,
             second one contains all signatures.
    :rtype: tuple

    :raises OSError:
        if reading command file fails or
        if :command:`ivxv-verify-container` command not found
    :raises subprocess.SubprocessError:
        on :command:`ivxv-verify-container` error
    :raises LookupError: on invalid signature line
    :raises IvxvError: on trust root config validation error
    """
    log.debug("Checking command file %r (%s) signature", filename, cmd_type)

    # detect trust root file
    trust_container_filepath = (
        filename if cmd_type == 'trust'
        else os.path.join(CONFIG['active_config_files_path'], 'trust.bdoc'))

    # execute verifier command
    proc = exec_container_verifier(trust_container_filepath, filename)

    # parse command output and create signatures list
    all_signatures = []
    for line in proc.stdout.strip().split('\n'):
        if not re.match(r'.+,.+,[0-9]{11} ', line):
            raise LookupError(f"Invalid signature line: {line}")
        signer, timestamp_str = line.split(' ')
        timestamp = datetime.datetime.strptime(
            timestamp_str, RFC3339_DATE_FORMAT_WO_FRACT).timestamp()
        all_signatures.append([timestamp, signer, line])
    all_signatures.sort()

    # check signers authorization for trust root config
    if cmd_type == 'trust':
        log.debug('Check signers authorization against trust root config')
        cfg = load_collector_cmd_file(cmd_type, filename)
        if cfg is None:
            raise IvxvError('Trust root file is not valid')
        trusted_signers = cfg.get('authorizations', [])
        authorized_signatures = [
            [signature, 'admin']
            for timestamp, signer, signature in all_signatures
            if signer in trusted_signers]
        return authorized_signatures, all_signatures

    # detect permission for command type
    if cmd_type == 'technical':
        permission = PERMISSION_TECH_CONF
    elif cmd_type in CFG_TYPES or cmd_type in VOTING_LIST_TYPES:
        permission = PERMISSION_ELECTION_CONF
    else:
        assert cmd_type == 'user'
        permission = PERMISSION_USERS_ADMIN

    # check signers authorization for other config files
    log.debug(
        'Check signers authorization against collector management database')
    authorized_signatures = []
    with IVXVManagerDb() as db:
        for timestamp, signer, signature in all_signatures:
            try:
                roles = db.get_value(f'user/{signer}')
            except KeyError:
                log.debug('No database record for signer %s', signer)
                continue
            authorized_signatures += [
                [signature, role] for role in roles.split(',')
                if permission in USER_ROLES[role]['permissions']
            ]

    return authorized_signatures, all_signatures


def exec_container_verifier(trust_container_filepath, filepath):
    """Execute container verifier command."""
    cmd = [
        'ivxv-verify-container', '-trust', trust_container_filepath, filepath
    ]
    log.debug("Executing command %r", " ".join(cmd))
    try:
        proc = subprocess.run(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            universal_newlines=True,
        )
    except OSError as err:
        err.strerror = (
            f"Error while executing verifier command {' '.join(cmd)!r}: {err.strerror}"
        )
        raise err

    if proc.returncode == 0:
        return proc

    verifier_errors = {
        64: 'Command was used incorrectly',
        65: 'Failed to open container',
        66: 'Input file did not exist or was not readable',
        74: 'Failed read trust root',
    }
    err_msg = verifier_errors.get(proc.returncode, 'Unhandled error')
    print('Container verifier output:')
    print(proc.stdout)
    print(proc.stderr)
    raise subprocess.SubprocessError(
        f'Failed to execute container verifier: {err_msg}')


def log_cfg_validation_errors(items, prefix="/"):
    """Log validation errors."""
    for field_name, val in items.items():
        if isinstance(val, dict):
            log_cfg_validation_errors(val, f"{prefix}{field_name}/")
        else:
            log.error("Validation error for field %r: %s", prefix + field_name, val)


def load_yaml_file(fp):
    """Load YAML file.

    :return: Content of loaded file
    :rtype: dict

    :raises OSError: if some file operation fails
    """

    def container_constructor_handler(loader, node):
        """Handler for !container constructor."""
        filename = loader.construct_scalar(node)
        if os.path.dirname(filename):
            raise OSError(
                f"Referenced file {filename!r} must be in the same "
                "directory with YAML file."
            )
        with open(filename) as fp:
            content = (yaml.load(fp, yaml.Loader) if filename[-5:] == '.yaml'
                       else fp.read(-1))
        return content

    def timestamp_constructor(loader, node):
        """
        Handler for timestamp fields.

        Returns string field instead of datetime.
        """
        assert loader  # unused variable
        return node.value

    yaml.add_constructor('tag:yaml.org,2002:timestamp', timestamp_constructor)
    yaml.add_constructor('!container', container_constructor_handler)

    return yaml.load(fp, yaml.Loader)
