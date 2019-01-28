# IVXV Internet voting framework
"""CLI utilities for collector config validation."""

import json

from .. import init_cli_util, log
from ... import CMD_DESCR, CMD_TYPES
from ...command_file import load_collector_cmd_file
from ...lib import IvxvError


def main():
    """Validate IVXV collector config."""
    args = init_cli_util("""
    Validate IVXV collector config files.

    Validate single config files. Also validate voting lists consistency if
    multiple lists are provided.

    Usage:
        ivxv-config-validate [--plain] [--trust=<trust-file>]
            [--technical=<technical-file>] [--election=<election-file>]
            [--choices=<choices-file>] [--districts=<districts-file>]
            [--voters=<voters-file> ...]

    Options:
        --plain     Validate plain config file (Default: BDOC container)
    """)
    plain = args['--plain']

    # validate CLI arguments
    cfg_files = []
    for cfg_type in CMD_TYPES:
        files = args.get(f'--{cfg_type}')
        if isinstance(files, list):
            cfg_files += [[cfg_type, filepath] for filepath in files]
        elif files:
            cfg_files.append([cfg_type, files])
    if not cfg_files:
        log.error('Config file is not specified')
        return 1

    # validate single files
    try:
        validated_cfg = validate_cfg_files(cfg_files, plain)
    except IvxvError:
        return 1
    log.info('Config files are valid')

    # validate lists consistency
    if args['--districts'] and (args['--choices'] or args['--voters']):
        if not validate_lists_consistency(validated_cfg):
            log.info('Voting lists consistency check failed')
            return 1

        log.info('Voting lists consistency check succeeded')

    return 0


def validate_cfg_files(cfg_files, plain):
    """Load and validate config files."""
    cfg_array = []
    for cfg_type, filepath in cfg_files:
        log.info('Validating %s file %s', CMD_DESCR[cfg_type], filepath)
        try:
            cfg = load_collector_cmd_file(cfg_type, filepath, plain)
        except json.decoder.JSONDecodeError:
            cfg = None
        if cfg is None:
            raise IvxvError
        cfg_array.append([cfg_type, cfg])

    return cfg_array


def validate_lists_consistency(cfg_objects):
    """Validate voting lists consistency."""
    districts_cfg = ([
        cfg
        for cfg_type, cfg in cfg_objects
        if cfg_type == 'districts'
    ][0])
    districts = set(districts_cfg['districts'])

    check_failed = False
    for cfg_type, cfg in cfg_objects:
        errors = []
        if cfg_type == 'choices':
            log.info('Checking districts and choices lists consistency')
            errors = validate_choices_consistency(districts, cfg)
            check_failed = check_failed or bool(errors)
        elif cfg_type == 'voters':
            log.info('Checking districts and voters lists consistency')
            errors = validate_voters_consistency(districts, districts_cfg, cfg)
            check_failed = check_failed or bool(errors)
        for error in errors:
            log.error(error)

    return not check_failed


def validate_choices_consistency(districts, choices_cfg):
    """Validate districts and choices lists consistency."""
    choices = set(choices_cfg['choices'])
    errors = []

    # Each choice is in an existing district
    choices_wo_district = choices.difference(districts)
    if choices_wo_district:
        errors.append(
            'The following choices have no matching district: {}'
            .format(','.join(sorted(choices_wo_district))))

    # In each district there must be at least one choice
    districts_wo_choices = districts.difference(choices)
    if districts_wo_choices:
        errors.append(
            'The following districts have no choice defined: {}'
            .format(','.join(sorted(districts_wo_choices))))

    return errors


def validate_voters_consistency(districts, districts_cfg, voters_cfg):
    """Validate voting lists consistency."""
    errors = []

    # generate stations list
    stations = set()
    for district_stations in districts_cfg['districts'].values():
        for station in district_stations['stations']:
            if station in stations:
                errors.append(f'Duplicate record for station {station}')
            else:
                stations.add(station)

    # check consistency
    for voter in voters_cfg['voters']:
        # Each voter must be in an existing station
        voter_station = '.'.join(voter[3:5])
        if voter_station not in stations:
            errors.append(f'Voter {voter[0]} have no station {voter_station}')

        # In each station there must be at least one voter voters
        voter_district = '.'.join(voter[5:7])
        if voter_district not in districts:
            errors.append(
                f'Voter {voter[0]} have no district {voter_district}')

    return errors
