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

    # validate multiple voters lists consistency
    if len(args['--voters']) > 1:
        if not validate_voters_lists_consistency(validated_cfg):
            log.info('Voters lists consistency check failed')
            return 1

        log.info('Voters lists consistency check succeeded')

    # validate districts and choices/voters lists consistency
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


def validate_voters_lists_consistency(cfg_objects):
    """Validate voters lists consistency."""
    check_failed = False
    voters_reg = None
    file_no = 0
    errors = []
    for cfg_type, cfg in cfg_objects:
        if cfg_type != 'voters':
            continue

        file_no += 1
        if file_no == 1:
            if cfg['list_type'] != 'algne':
                errors.append(
                    f'Invalid type "{cfg["list_type"]}" '
                    'for initial voters list')
                break
            voters_reg = dict(
                [voter[0], voter[3:7]] for voter in cfg['voters'])
            continue

        log.info('Checking voters list patch #%d consistency', file_no - 1)
        if cfg['list_type'] != 'muudatused':
            errors.append(
                f'Invalid type "{cfg["list_type"]}" '
                'for voters list patch')
            break
        voters_in_patch = set()
        for rec_no, voter in enumerate(cfg['voters'], start=1):
            voter_id = voter[0]

            if voter[2] == 'lisamine':
                if (voter_id not in voters_reg
                        and voter_id not in voters_in_patch):
                    voters_reg[voter_id] = voter[3:7]
                else:
                    errors.append(
                        f'Record #{rec_no}: Adding voter ID {voter_id} '
                        'that is already in voters list')
                voters_in_patch.add(voter_id)
            else:
                if voter_id in voters_in_patch:
                    errors.append(
                        f'Record #{rec_no}: Removing ID {voter_id} '
                        'that is added with this patch')
                try:
                    district = voters_reg.pop(voter_id)
                    if district != voter[3:7]:
                        errors.append(
                            f'Record #{rec_no}: Removing voter ID {voter_id} '
                            f'from invalid district {voter[3:7]}. '
                            f'Voter is registered in district {district}')
                except KeyError:
                    errors.append(
                        f'Record #{rec_no}: Removing voter ID {voter_id} '
                        'that is not in voters list')

        if errors:
            break

    for error in errors:
        log.error(error)

    check_failed = check_failed or bool(errors)

    return not check_failed


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

    for district in districts_cfg['districts']:
        district_stations = districts_cfg['districts'][district]
        for station in district_stations['stations']:
            uniq_station = '.'.join([station, district])
            if uniq_station in stations:
                errors.append(f'Duplicate record for station {station} '
                              f'in district {district}')
            else:
                stations.add(uniq_station)

    # check consistency
    for voter in voters_cfg['voters']:

        # In each station there must be at least one voter voters
        voter_district = '.'.join(voter[5:7])
        if voter_district not in districts:
            errors.append(
                f'Voter {voter[0]} in non-existing district {voter_district}')

        # Each voter must be in an existing station
        voter_station = '.'.join(voter[3:5])
        uniq_station = '.'.join(voter[3:7])
        if uniq_station not in stations:
            errors.append(f'Voter {voter[0]} in non-existing station '
                          f'{voter_station} in district {voter_district}')

    return errors
