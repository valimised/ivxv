#!/usr/bin/python

"""Data Validation Tool

Usage:
    validator -d <data> -s <schema>
    validator -h | --help
    validator -v | --version

Options:
    -h --help               Request for help
    -v --version            Show version information
    -d --data <data>        Data file
    -s --schema <schema>    Schema file
"""

import os

import docopt
import schema

import json
import jsonschema


if __name__ == "__main__":

    ARGS = docopt.docopt(__doc__, version='Data Validation Tool 1.0')
    SCHEMA = schema.Schema({
        '--help': schema.Or(False),
        '--version': schema.Or(False),
        '--data': os.path.isfile,
        '--schema': os.path.isfile
    })
    try:
        ARGS = SCHEMA.validate(ARGS)
    except schema.SchemaError as err:
        exit(err)

    DATA_F = ARGS['--data']
    SCHEMA_F = ARGS['--schema']

    with open(SCHEMA_F) as sif:
        with open(DATA_F) as dif:
            schema = json.load(sif)
            data = json.load(dif)
            jsonschema.validate(data, schema)

    print('Validation ended')
