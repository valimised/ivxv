#!/usr/bin/python3
"""
IVXV Voting List Data Consistency Tool

Usage:
    consistency [options]
    consistency -h | --help
    consistency -v | --version

Options:
    -h --help                    Request for help
    -v --version                 Show version information
    -d --districts <districts>   Districts file
    -l --lvoters <lvoters>       Voters file
    -c --choices <choices>       Choices file
    -m --mapping <mapping>       Districts mapping file
"""

import json
import os
import re

import docopt
import jsonschema

import schema

RE_ELID = re.compile(r'^\w{1,28}$')
RE_ISIKUKOOD = re.compile(r'^\d{11}$')
RE_100UTF8 = re.compile(r'^.{1,100}$', re.UNICODE)
RE_200UTF8 = re.compile(r'^.{1,200}$', re.UNICODE)
RE_100UTF8OPT = re.compile(r'^.{0,100}$', re.UNICODE)
RE_VERSIOON = re.compile(r'^\d{1,2}$')
RE_VALIK = re.compile(r'^\d{1,10}\.\d{1,11}$')
RE_NUMBER10 = re.compile(r'^\d{1,10}$')
RE_BASE64 = re.compile(r'^[a-zA-Z0-9+/=]+$')
RE_BASE64_LINES = re.compile(r'^[a-zA-Z0-9+/=\n]+$')
RE_HEX = re.compile(r'^[a-fA-F0-9]+$')
RE_LINENO = re.compile(r'^\d{0,11}$')
RE_PHONENO = re.compile(r'^\+(\d){7,15}$')
RE_SIGNING_TIME = re.compile(r'^[TZ0-9:-]+$')
RE_NUMBER100 = re.compile(r'^\d{1,100}$')

MAX_VOTE_BASE64_SIZE = 24000


# pylint: disable=C0103
def is_jaoskonna_number_kov_koodiga(p1, p2):
    if not is_jaoskonna_omavalitsuse_kood(p1):
        return False
    if not is_jaoskonna_number_kov_sees(p2):
        return False
    return True


def is_ringkonna_number_kov_koodiga(p1, p2):
    if not is_ringkonna_omavalitsuse_kood(p1):
        return False
    if not is_ringkonna_number_kov_sees(p2):
        return False
    return True


def is_jaoskonna_number_kov_sees(nr):
    return is_jaoskonna_number(nr)


def is_jaoskonna_omavalitsuse_kood(nr):
    return is_omavalitsuse_kood(nr)


def is_ringkonna_number_kov_sees(nr):
    return is_ringkonna_number(nr)


def is_ringkonna_omavalitsuse_kood(nr):
    return is_omavalitsuse_kood(nr)


def is_omavalitsuse_kood(nr):
    return _is_number10(nr)


def is_valimiste_identifikaator(elid):
    return RE_ELID.match(elid) is not None


def is_isikukood(code):
    return RE_ISIKUKOOD.match(code) is not None


def is_ringkonna_number(nr):
    return _is_number10(nr)


def is_jaoskonna_number(nr):
    return _is_number10(nr)


def is_100utf8(sstr):
    return RE_100UTF8.match(sstr) is not None


def is_200utf8(sstr):
    return RE_200UTF8.match(sstr) is not None


def is_100utf8optional(sstr):
    return RE_100UTF8OPT.match(sstr) is not None


def is_nimi(name):
    return is_100utf8(name)


def is_pohjus(sstr):
    return is_100utf8(sstr)


def is_valiku_kood(code):
    return RE_VALIK.match(code) is not None


def is_valiku_nimi(name):
    return is_100utf8(name)


def is_valimisnimekirja_nimi(name):
    return is_100utf8optional(name)


def is_rea_number_voi_tyhi(nr):
    return RE_LINENO.match(nr) is not None


def is_number100(nr):
    return RE_NUMBER100.match(nr) is not None


def _is_number10(nr):
    return RE_NUMBER10.match(nr) is not None


def validate_voter(voter):
    pass


def convert_districts(districts):
    for dist in districts['districts']:
        if dist.startswith("0784"):
            for stat in districts['districts'][dist]['stations']:
                to = "%s\t%s" % ('\t'.join(stat.split('.')),
                                 '\t'.join(dist.split('.')))

                fro = "0784\t%s\t%s" % (stat.split('.')[1],
                                        '\t'.join(dist.split('.')))
                print("%s\t%s" % (fro, to))


def use_mapping(vstat, vdist, mapping):

    rstat = vstat
    rdist = vdist

    if mapping:
        key = "%s\t%s" % (vstat, vdist)
        if key in mapping:
            rstat = mapping[key]['vstat']
            rdist = mapping[key]['vdist']

    return rstat, rdist


def with_mapping(fname):

    if fname is None:
        return None

    mapping = {}

    with open(fname) as mapf:
        for line in mapf.readlines():
            tokens = line.rstrip('\n').split('\t')
            assert len(tokens) == 8, 'Invalid mapping'
            key = "%s.%s\t%s.%s" % (tokens[0], tokens[1], tokens[2], tokens[3])
            mapping[key] = {
                'vstat': "%s.%s" % (tokens[4], tokens[5]),
                'vdist': "%s.%s" % (tokens[6], tokens[7])
            }
    return mapping


def consistency_voters_districts(voters,
                                 districts,
                                 mapping=None,
                                 warnempty=False):

    # Each voter is in an existing station

    distvot = set()

    for el in voters:
        vdist = voters[el]['district']
        vstat = voters[el]['station']

        if (vdist not in districts['districts']):
            print('Warning! Missing district %s' % vdist)
            continue

        vstat, vdist = use_mapping(vstat, vdist, mapping)

        assert vstat in districts['districts'][vdist]['stations'], (
            'Missing station %s in district %s' % (vstat, vdist))

        distvot.add("%s.%s" % (vdist, vstat))

    # In each station there is at least one voter

    if warnempty:
        for ddist in districts['districts']:
            for dstat in districts['districts'][ddist]['stations']:
                key = "%s.%s" % (ddist, dstat)
                if key not in distvot:
                    print("Warning - no voters in %s" % key)


def consistency_choices_districts(choices, districts):

    # Each choice is in an existing district
    for el in choices['choices']:
        assert el in districts['districts'], (
            'Missing district %s in choices' % el)

    # In each district there is at least one choice
    for el in districts['districts']:
        assert choices['choices'][el], 'Missing choices in district %s' % el


def load_and_validate(data, schema_filename):
    with open(schema_filename) as sif:
        with open(data) as dif:
            schema_data = json.load(sif)
            retval = json.load(dif)
            jsonschema.validate(retval, schema_data)

    return retval


def load_districts(arg):
    if arg is None:
        print("All districts-related checks are supressed")
        return None

    print("Validating districts schema")
    return load_and_validate(arg, 'ivxv.districts.schema')


def load_choices(arg):
    if arg is None:
        print("All choices-related checks are supressed")
        return None

    print("Validating choices schema")
    return load_and_validate(arg, 'ivxv.choices.schema')


def load_voterfiles(arg):

    retval = []

    for voter_f in arg.split(','):

        print("Validating voterlist schema %s" % voter_f)

        actions = {
            'file': voter_f,
            'elid': None,
            'type': None,
            'kustutamine': {},
            'lisamine': {}
        }

        with open(voter_f) as dif:
            line1 = dif.readline()
            assert line1 == '1\n', 'Invalid voterlist version'
            actions['elid'] = dif.readline().rstrip('\n')
            line3 = dif.readline().rstrip('\n')
            assert line3 in ['algne', 'muudatused'], 'Invalid voterlist type'
            actions['type'] = line3
            for line in dif.readlines():
                tokens = line.rstrip('\n').split('\t')
                assert len(tokens) == 9, len(tokens)

                idd = tokens[0]

                action = tokens[2]
                voter = {
                    'id': idd,
                    'name': tokens[1],
                    'station': "%s.%s" % (tokens[3], tokens[4]),
                    'district': "%s.%s" % (tokens[5], tokens[6]),
                    'lineno': tokens[7],
                    'reason': tokens[8]
                }

                validate_voter(voter)

                assert idd not in actions[action], (
                    'Multiple similar actions for the same voter, %s' % idd)

                if line3 == 'algne':
                    assert action == 'lisamine', (
                        'Only additions in initial list')
# TODO - this check is importan but cannot function this way
#                if action == 'kustutamine':
#                    assert idd in actions['lisamine'], (
#                        'deletion/addition in wrong order %s' % idd)

                actions[action][idd] = voter

        retval.append(actions)

        print("\t%d additions" % len(actions['lisamine']))
        print("\t%d deletions" % len(actions['kustutamine']))

    print("%d voterfiles loaded" % len(retval))
    return retval


def load_voters(arg, districts=None):
    if arg is None:
        print("All voters-related checks are supressed")
        return None

    voter_db = {}
    voterfiles = load_voterfiles(arg)

    for voter_f in voterfiles:
        if districts is not None:
            print("Validate voters vs. districts %s" % voter_f['file'])
            assert voter_f['elid'] == districts['election'], (
                'Election identifiers do not match')
            consistency_voters_districts(voter_f['lisamine'], districts,
                                         with_mapping(ARGS['--mapping']))

            consistency_voters_districts(voter_f['kustutamine'], districts,
                                         with_mapping(ARGS['--mapping']))

    if len(voterfiles) > 1:
        assert voterfiles[0]['type'] == 'algne'
        for voter_f in voterfiles[1:]:
            assert voter_f['type'] == 'muudatused'

    for voter_f in voterfiles:

        print("processing %s" % voter_f['file'])

        aclist = voter_f['kustutamine']
        for idd in aclist:
            assert idd in voter_db, 'Voter must be in DB for removal'

            sd_a = "%s.%s" % (aclist[idd]['station'], aclist[idd]['district'])
            sd_b = "%s.%s" % (voter_db[idd]['station'],
                              voter_db[idd]['district'])

            assert sd_a == sd_b, (
                'Wrong station number for the voter %s %s != %s' %
                (idd, sd_a, sd_b))

            del voter_db[idd]

        aclist = voter_f['lisamine']
        for idd in aclist:
            assert idd not in voter_db, 'Voter must not be in DB for addition'
            voter_db[idd] = aclist[idd]

        print("%d voters in DB" % len(voter_db))

    if districts:
        consistency_voters_districts(voter_db, districts,
                                     with_mapping(ARGS['--mapping']), True)

    return voter_db


def is_voterlist_list(arg):

    files = arg.split(',')
    for el in files:
        if not os.path.isfile(el):
            return False

    return True


if __name__ == "__main__":
    ARGS = docopt.docopt(
        __doc__, version='IVXV Voting List Data Consistency Tool 1.0')
    SCHEMA = schema.Schema({
        '--help': schema.Or(False),
        '--version': schema.Or(False),
        '--districts': schema.Or(None, os.path.isfile),
        '--lvoters': schema.Or(None, is_voterlist_list),
        '--choices': schema.Or(None, os.path.isfile),
        '--mapping': schema.Or(None, os.path.isfile)
    })
    try:
        ARGS = SCHEMA.validate(ARGS)
    except schema.SchemaError as err:
        exit(err)

    DISTRICTS = load_districts(ARGS['--districts'])
    CHOICES = load_choices(ARGS['--choices'])
    VOTERS = load_voters(ARGS['--lvoters'], DISTRICTS)

    if DISTRICTS is not None and CHOICES is not None:
        print("Validate choices vs. districts")
        consistency_choices_districts(CHOICES, DISTRICTS)

    print('Validation ended')
