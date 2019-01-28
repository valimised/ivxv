# IVXV Internet voting framework
"""Voters list validator."""

import re

from .fields import ElectionIdType

VOTERS_LIST_TYPES = ['algne', 'muudatused']
VOTER_ACTION_TYPES = ['lisamine', 'kustutamine']


def parse_voters_list(list_content):
    """Parse voters list.

    :return: Voters list data
    :rtype: dict

    :raises: ValueError
    """
    lines = list_content.rstrip('\n').split('\n')
    if len(lines) < 3:
        raise ValueError(f'Too few lines in voters list ({len(lines)})')

    # parse header
    data = dict(version=lines[0])
    data['election'] = lines[1]
    data['list_type'] = lines[2]

    # validate header
    if data['version'] != '1':
        raise ValueError(
            'Invalid voters list version "{}" in line 1. Expected value: "1"'
            .format(data['version']))
    ElectionIdType(required=True).validate(data['election'])
    if data['list_type'] not in VOTERS_LIST_TYPES:
        raise ValueError(
            'Unknown voters list type "{}" in line 3. Expected values: {}'
            .format(data['list_type'], ', '.join(VOTERS_LIST_TYPES)))

    # parse list
    is_original_list = data['list_type'] == 'algne'
    data['voters'] = []
    for line_no, line in enumerate(lines, 1):
        if line_no < 4:
            continue
        fields = line.split('\t')
        try:
            validate_voter_record(fields, is_original_list)
        except ValueError as err:
            raise ValueError(f'Line #{line_no}: {err.args[0]}')
        data['voters'].append(fields)

    return data


def validate_voter_record(fields, is_original_list):
    """Validate voter record in voters list."""
    # field count
    if len(fields) != 9:
        raise ValueError(
            'Invalid field count {}, expected 9 fields'.format(len(fields)))
    # field #1 - voter-personalcode
    if not re.match(r'[0-9]{11}$', fields[0]):
        raise ValueError('Invalid ID code "{}"'.format(fields[0]))
    # field #2 - voter-name
    if not fields[1]:
        raise ValueError('Empty person name')
    # field #3 - action
    if fields[2] not in VOTER_ACTION_TYPES:
        raise ValueError('Unknown action "{}". Expected values are {}'
                         .format(fields[2], VOTER_ACTION_TYPES))
    # field #4...7 - station-legacy
    for field_no in 3, 4, 5, 6:
        try:
            int(fields[field_no])
        except ValueError:
            raise ValueError('Invalid district/station number "{}"'
                             .format(fields[field_no]))
    # field #8 - line-no
    if is_original_list and fields[7]:
        try:
            int(fields[7])
        except ValueError:
            raise ValueError('Invalid voter line number "{}"'
                             .format(fields[7]))
    # field #9 - reason
    if is_original_list and fields[8] != '':
        raise ValueError(
            'Invalid reason "{}". Must be empty for original list.'
            .format(fields[8]))
