# IVXV Internet voting framework
"""
Command file handling module collector management service.
"""

import datetime
import json
import logging
import os
import re
import subprocess
import tempfile
import zipfile

import yaml
import schematics.exceptions

from . import (COMMAND_TYPES, CONFIG_TYPES, RFC3339_DATE_FORMAT_WO_FRACT,
               ROLES, VOTING_LIST_TYPES)
from . import (PERMISSION_ELECTION_CONF, PERMISSION_TECH_CONF,
               PERMISSION_USERS_ADMIN)
from . import config_validator
from .config import CONFIG
from .db import IVXVManagerDb

# create logger
log = logging.getLogger(__name__)


def load_collector_command_file(cmd_type, filename):
    """
    Load collector command file.

    Allowed command file types are:

        - collectors config package (trust, election, technical);

        - voting list (choices, voters).
    """
    assert cmd_type in COMMAND_TYPES, 'Invalid command type "%s"' % cmd_type

    if cmd_type in CONFIG_TYPES:
        cmd_filename = cmd_type + '.yaml'
        file_descr = CONFIG_TYPES[cmd_type]
    elif cmd_type in VOTING_LIST_TYPES:
        cmd_filename = None
        file_descr = VOTING_LIST_TYPES[cmd_type]
    elif cmd_type == 'user':
        assert cmd_type == 'user'
        cmd_filename = 'user.json'
        file_descr = 'user managment command'

    # load content from file
    config = load_cmd_file_content(cmd_type, cmd_filename,
                                   filename, file_descr)

    # FIXME don't validate voters list. there are some errors
    #       in validation that need to be fixed in the future.
    if cmd_type == 'voters':
        return config

    # validate config file
    log.info('Validating %s', file_descr)
    try:
        config = config_validator.validate_config(config, cmd_type)
    except ValueError as err:
        log.error(err)
        return
    except schematics.exceptions.DataError as data_error:
        _log_config_validation_errors(data_error.errors)
        return
    log.info('Files in %s package are valid', file_descr)

    return config


def load_cmd_file_content(cmd_type, cmd_filename, bdoc_filename, file_descr):
    """Load content of collector command file."""
    # load command file
    try:
        log.info('Loading command file "%s" (%s)', bdoc_filename, file_descr)
        cmd_file = zipfile.ZipFile(bdoc_filename)
    except FileNotFoundError:
        log.info('File "%s" not found', bdoc_filename)
        return
    except zipfile.BadZipFile:
        log.info('File "%s" is not ZIP container', bdoc_filename)
        return
    log.info('Command file loaded')
    log.debug('Files in command file container: %s',
              ' '.join(cmd_file.namelist()))

    # parse contents of command file
    filenames_in_bdoc = cmd_file.namelist()
    for filename in filenames_in_bdoc[:]:
        if filename in ['META-INF/manifest.xml', 'META-INF/signatures0.xml',
                        'mimetype']:
            filenames_in_bdoc.remove(filename)
    if cmd_filename is None:
        cmd_filename = _get_cmd_filename(cmd_type, filenames_in_bdoc)
        log.debug('Detected command filename: %s', cmd_filename)

    if cmd_filename not in filenames_in_bdoc:
        log.error('Command file does not contain expected file %s for %s',
                  cmd_filename, file_descr)
        return

    # create temporary directory to extract container
    with tempfile.TemporaryDirectory() as dirname:
        log.debug('Extracting command file to directory %s', dirname)
        cmd_file.extractall(dirname)
        log.info('Command file successfully extracted')

        log.info('Reading files from command file')
        cwd = os.getcwd()
        os.chdir(dirname)
        with open(cmd_filename) as fp:
            if cmd_type in CONFIG_TYPES:
                config = load_yaml_file(fp)
            elif cmd_type in ('choices', 'user'):
                config = json.load(fp)
            else:
                config = fp.read()
        os.chdir(cwd)

    return config


def _get_cmd_filename(cmd_type, filenames_in_bdoc):
    """
    Check filenames in voting list command package
    and return voting list filename.
    """
    if cmd_type == 'choices':
        if len(filenames_in_bdoc) > 1:
            log.error(
                'Too many files in choices command file: %s',
                filenames_in_bdoc)
            return
        elif not filenames_in_bdoc:
            log.error('Missing choices list in choices command file')
            return
        return filenames_in_bdoc[0]

    if cmd_type == 'voters':
        if len(filenames_in_bdoc) > 2:
            log.error(
                'Too many files in voters command file: %s',
                filenames_in_bdoc)
            return
        elif len(filenames_in_bdoc) < 2:
            log.error(
                'Missing voters list or signature in voters command file')
            return

        cmd_filename, signature_filename = sorted(filenames_in_bdoc)
        if cmd_filename + '.signature' != signature_filename:
            log.error('Voters list and signature file names does not match '
                      '(filenames: %s)',
                      ', '.join([cmd_filename, signature_filename]))
            log.error('Signature file name must be list file '
                      'with ".signature" suffix')
            return

        return cmd_filename

    raise NotImplementedError('Unknown command type: %s' % cmd_type)


def check_cmd_signature(cmd_type, filename):
    """
    Check collector command file signature.

    :returns: Two lists, first one contains signatures by authorized users,
              second one contains all signatures.
    """
    log.debug('Checking command file %s (%s) signature', filename, cmd_type)

    # detect trust root file
    trust_container_filepath = os.path.join(CONFIG['active_config_files_path'],
                                            'trust.bdoc')
    if cmd_type == 'trust' and not os.path.exists(trust_container_filepath):
        trust_container_filepath = filename

    try:
        open(trust_container_filepath)
    except OSError as err:
        err.strerror = "Trust root not found: %s" % trust_container_filepath
        raise err

    # execute verifier command
    cmd = ['ivxv-container-verifier', '-trust', trust_container_filepath,
           filename]
    proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                          universal_newlines=True)

    if proc.returncode:
        verifier_errors = {
            64: 'Command was used incorrectly',
            65: 'Failed to open container',
            66: 'Input file did not exist or was not readable',
            74: 'Failed read trust root',
        }
        err_msg = verifier_errors.get(proc.returncode, 'Unhandled error')
        raise subprocess.SubprocessError(': '.join([err_msg, proc.stderr]))

    # parse command output and create signatures list
    all_signatures = []
    for line in proc.stdout.strip().split('\n'):
        if not re.match(r'.+,.+,[0-9]{11} ', line):
            raise LookupError('Invalid signature line: %s' % line)
        signer, timestamp_str = line.split(' ')
        timestamp = datetime.datetime.strptime(
            timestamp_str, RFC3339_DATE_FORMAT_WO_FRACT).timestamp()
        all_signatures.append([timestamp, signer, line])
    all_signatures = sorted(all_signatures)

    # check signers authorization for trust root config
    if cmd_type == 'trust':
        log.debug('Check signers authorization against trust root config')
        config = load_collector_command_file(cmd_type, filename)
        trusted_signers = config.get('authorizations', [])
        authorized_signatures = [
            [signature, 'admin']
            for timestamp, signer, signature in all_signatures
            if signer in trusted_signers]
        return authorized_signatures, all_signatures

    # detect permission for command type
    if cmd_type == 'technical':
        permission = PERMISSION_TECH_CONF
    elif cmd_type in CONFIG_TYPES or cmd_type in VOTING_LIST_TYPES:
        permission = PERMISSION_ELECTION_CONF
    else:
        assert cmd_type == 'user'
        permission = PERMISSION_USERS_ADMIN

    # check signers authorization for other config files
    log.debug(
        'Check signers authorization against collector management database')
    authorized_signatures = []
    db = IVXVManagerDb()
    for timestamp, signer, signature in all_signatures:
        try:
            roles = db.get_value('user/{}'.format(signer))
        except KeyError:
            log.debug('No database record for signer %s', signer)
            continue
        authorized_signatures += [[signature, role]
                                  for role in roles.split(',')
                                  if permission in ROLES[role]['permissions']]
    db.close()

    return authorized_signatures, all_signatures


def load_yaml_file(fp):
    """Load YAML file."""
    def container_constructor_handler(loader, node):
        """Handler for !container constructor."""
        filename = loader.construct_scalar(node)
        if os.path.dirname(filename):
            raise AssertionError('Referenced file "{}" must be in the same '
                                 'directory with YAML file.'.format(filename))
        with open(filename) as fp:
            content = (yaml.load(fp) if filename[-5:] == '.yaml'
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

    return yaml.load(fp)


def _log_config_validation_errors(items, prefix='/'):
    """Log validation errors."""
    for field_name, val in items.items():
        if isinstance(val, dict):
            _log_config_validation_errors(val, prefix + field_name + '/')
        else:
            log.error('Validation error for field "%s%s": %s',
                      prefix, field_name, val)
