# IVXV Internet voting framework
"""CLI utilities for collector config validation."""

import json

from ... import CMD_DESCR, CMD_TYPES
from ...command_file import load_collector_cmd_file
from ...lib import IvxvError
from .. import init_cli_util, log


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

    if validate_cfg_consistency(
        validated_cfg,
        election=args["--election"],
        choices=args["--choices"],
        districts=args["--districts"],
        voters=args["--voters"],
    ):
        return 0
    return 1


def validate_cfg_consistency(
    validated_cfg,
    election=None,
    choices=None,
    districts=None,
    voters=None,
):
    """Perform consistency checks for config files."""
    # validate election ID consistency
    if not validate_cfg_election_id(validated_cfg):
        log.info("Election ID consistency check failed")
        return False
    log.info("Election ID consistency check succeeded")

    # validate multiple voters lists consistency
    if len(voters) > 1:
        if not validate_voters_lists_consistency(validated_cfg):
            log.info('Voters lists consistency check failed')
            return False

        log.info('Voters lists consistency check succeeded')

    # determine voterforeignehak, there are no consistency checks
    voterforeignehak = "0000"
    if election:
        for cfg_type, cfg in validated_cfg:
            if cfg_type == "election":
                voterforeignehak = cfg.get("voterforeignehak", voterforeignehak)
                break

    # validate districts and choices/voters lists consistency
    if districts and (choices or voters):
        if not validate_lists_consistency(validated_cfg, voterforeignehak):
            log.info('Voting lists consistency check failed')
            return False

        log.info('Voting lists consistency check succeeded')

    return True


def validate_cfg_election_id(validated_cfg):
    """Validate election ID consistency in config files."""
    # try to detect election ID
    election_id = None
    for cfg_type, cfg in validated_cfg:
        if cfg_type == "election":
            election_id = cfg["identifier"]
            log.info("Detected election ID %r from election config", election_id)
            break

    # validate election ID consistency
    for cfg_type, cfg in validated_cfg:
        if cfg_type in ["election", "technical", "trust"]:
            continue
        cfg_election_id = cfg["election"]
        if not election_id:
            election_id = cfg_election_id
            log.debug("Detected election ID %r from %s config", election_id, cfg_type)
        else:
            if election_id != cfg_election_id:
                log.error(
                    "Election ID %r in %s config does not match with %r",
                    cfg_election_id,
                    cfg_type,
                    election_id,
                )
                return False

    return True


def validate_cfg_files(cfg_files, plain):
    """Load and validate config files."""
    cfg_array = []
    for cfg_type, filepath in cfg_files:
        log.info("Validating %s file %r", CMD_DESCR[cfg_type], filepath)
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
    changeset = 0
    errors = []
    for cfg_type, cfg in cfg_objects:
        if cfg_type != 'voters':
            continue

        if changeset == 0:
            if int(cfg['changeset']) != 0:
                errors.append(
                    f'Invalid changeset {cfg["changeset"]} for initial voter list'
                )
                break
            voters_reg = dict(
                [voter[1], voter[3:5]] for voter in cfg['voters'])
            changeset = 1
            continue

        if int(cfg['changeset']) < changeset:
            errors.append(
                f'Invalid changeset {cfg["changeset"]}, '
                f'expected {changeset} or greater')
            break

        changeset = int(cfg['changeset'])

        # voter list skipping command acts as an empty changeset
        if cfg.get('skip_voter_list'):
            continue

        log.info("Checking voters list changeset #%d consistency", changeset)
        voters_in_patch = set()
        for rec_no, voter in enumerate(cfg['voters']):
            voter_id = voter[1]

            if voter[0] == 'lisamine':
                if (voter_id not in voters_reg
                        and voter_id not in voters_in_patch):
                    voters_reg[voter_id] = voter[3:5]
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
                    # just to test that voter id is valid before removing it
                    voters_reg.pop(voter_id)
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


def validate_lists_consistency(cfg_objects, voterforeignehak):
    """Validate voting lists consistency."""
    districts_cfg = ([
        cfg
        for cfg_type, cfg in cfg_objects
        if cfg_type == 'districts'
    ][0])
    districts = set(districts_cfg['districts'])

    check_failed = False
    changeset_no = 0
    for cfg_type, cfg in cfg_objects:
        errors = []
        if cfg_type == 'choices':
            log.info('Checking districts and choices lists consistency')
            errors = validate_choices_consistency(districts, cfg)
            check_failed = check_failed or bool(errors)
        elif cfg_type == 'voters':
            log.info(
                "Checking districts and voter list changeset #%d consistency",
                changeset_no,
            )
            errors = validate_voters_consistency(districts_cfg, districts, cfg,
                                                 voterforeignehak)
            check_failed = check_failed or bool(errors)
            changeset_no += 1
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


def validate_voters_consistency(
    districts_cfg, districts, voters_cfg, voterforeignehak
):
    """Validate voting lists consistency."""
    errors = []

    if "skip_voter_list" in voters_cfg:
        return errors

    parish_to_district = {}

    for dist in districts_cfg["districts"]:
        for parish in districts_cfg["districts"][dist]["parish"]:
            if parish not in parish_to_district:
                parish_to_district[parish] = []

            parish_to_district[parish].append(dist)

    # check consistency
    for voter in voters_cfg['voters']:
        # if action is "kustutamine" then don't check districts/parish
        if voter[0] == "kustutamine":
            continue
        # Voterlist contains voter EHAK and district no which must be
        # translated into full district identifier
        voter_district = None
        voter_id = voter[1]
        voter_ehak = voter[3]
        voter_district_no = voter[4]

        # Voterlist uses parish FOREIGN, other configs may use some other code
        if voter_ehak == "FOREIGN":
            voter_ehak = voterforeignehak

        # Search district by parish, single parish may be in several districts
        # in this case the district no's must be different and match with voter
        if voter_ehak in parish_to_district:
            for dist in parish_to_district[voter_ehak]:
                if dist.split(".")[1] == voter_district_no:
                    voter_district = dist
                    break
        else:
            errors.append(f'Parish {voter_ehak} not found in districts')

        # Voter must belong to an existing district
        if voter_district is None:
            errors.append(
                f'District for voter {voter_id} in EHAK {voter_ehak} / '
                f'{voter_district_no} cannot be determined')
        elif voter_district not in districts:
            errors.append(
                f'District {voter_district} for voter {voter_id} not found')

    return errors
