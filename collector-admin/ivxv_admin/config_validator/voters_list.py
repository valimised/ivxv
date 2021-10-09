# IVXV Internet voting framework
"""Voters list validator."""

import re

import dateutil.parser
from schematics.models import Model
from schematics.types import IntType, StringType

from .fields import ElectionIdType

ISOPARSER = dateutil.parser.isoparser("T")


class VoterListChangesetSkipSchema(Model):
    """Validating schema for voter list changeset skipping command."""

    election = ElectionIdType(required=True)
    skip_voter_list = StringType(regex=r"^[^ ]+ [^ ]+$", required=True)
    changeset = IntType(required=True)


def parse_voters_list(list_content):
    """Parse voters list.

    :return: Voters list data
    :rtype: dict

    :raises ValueError:
    """
    lines = list_content.split('\n')
    if len(lines) < 4:
        raise ValueError(f'Too few lines in voters list ({len(lines)})')

    # parse header
    data = dict(version=lines[0].rstrip('\r'))
    data['election'] = lines[1].rstrip('\r')
    try:
        data['election'].encode('ASCII')
    except UnicodeEncodeError:
        raise ValueError('Election ID contains non-ASCII characters')
    data['changeset'] = lines[2].rstrip('\r')
    data['period'] = lines[3].rstrip('\r')
    timestamps = data['period'].split('\t')
    if len(timestamps) != 2:
        raise ValueError('Period does not contain two tab-separated fields')
    data['period_from'] = timestamps[0]
    data['period_to'] = timestamps[1].rstrip('\r')

    # validate header
    # version-no = "2"
    if data["version"] != "2":
        raise ValueError(
            f"Invalid voters list version {data['version']!r} in line 1. "
            "Expected value: 2"
        )
    ElectionIdType(required=True).validate(data['election'])
    # changeset = integer
    if not data['changeset'].isdigit():
        raise ValueError(
            f"Unknown voters list changeset {data['changeset']!r} in line 3. "
            "Must be an integer"
        )
    # period_from, period_to = RFC 3339 timestamp
    try:
        # FIXME: Accepts all ISO 8601 timestamps, not only RFC 3339.
        ISOPARSER.isoparse(data['period_from'])
        ISOPARSER.isoparse(data['period_to'])
    except ValueError as err:
        raise ValueError(f'Period contains invalid timestamp: {err.args[0]}')

    # parse list
    is_original_list = data['changeset'] == '0'
    data['voters'] = []
    for line_no, line in enumerate(lines[:-1], 1):
        if '\r' in line:
            raise ValueError(f'Line #{line_no}: Invalid character <CR>')
        if line_no < 5:
            continue
        fields = line.split('\t')
        try:
            validate_voter_record(fields, is_original_list)
        except ValueError as err:
            raise ValueError(f'Line #{line_no}: {err.args[0]}')
        data['voters'].append(fields)

    if lines[-1] != '':
        raise ValueError(f'Line #{len(lines)}: Must end with <LF> character')

    return data


def validate_voter_record(fields, is_original_list):
    """Validate voter record in voters list."""
    # field count
    try:
        voter_personalcode, voter_name, action, adminunit_code, no_district = fields
    except ValueError:
        raise ValueError(f"Invalid field count {len(fields)}, expected 5 fields")
    # voter-personalcode = 11DIGIT
    if not re.match(r"[0-9]{11}$", voter_personalcode):
        raise ValueError(f"Invalid voter-personalcode {voter_personalcode!r}")
    # voter-name = 1*100UTF-8-CHAR
    if not voter_name:
        raise ValueError("voter-name is empty")
    if len(voter_name) > 100:
        raise ValueError(f"voter-name lenght {len(voter_name)} exceeds 100 chars")
    # action = "lisamine" | "kustutamine"
    if action not in ["lisamine", "kustutamine"]:
        raise ValueError(
            f"Unknown action {action!r}. Must be 'lisamine' or 'kustutamine'"
        )
    if is_original_list and action != "lisamine":
        raise ValueError(f"Action {action!r} is not allowed in initial list")
    # adminunit-code = 1*4UTF-8-CHAR | "FOREIGN"
    if not adminunit_code:
        raise ValueError("Missing adminunit-code")
    if len(adminunit_code) > 4 and adminunit_code != "FOREIGN":
        raise ValueError(f"adminunit-code {adminunit_code!r} is longer than 4 chars")
    # no-district = 1*10DIGIT
    if not re.match(r"[0-9]{1,10}$", no_district):
        raise ValueError(f"Invalid no-district {no_district!r}")
