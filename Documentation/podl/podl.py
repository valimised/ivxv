#!/usr/bin/env python3

# IVXV Internet voting framework

"""PO translator with Deepl

Translate PO files with Deepl

Usage:
    podl --input <input> --api <api> [options]
    podl -h | --help
    podl -v | --version

Options:
    -h --help               Request for help
    -v --version            Show version information
    -i --input <input>      Input file or directory
    -s --source <source>    Source language [default: ET]
    -t --target <target>    Target language [default: EN-GB]
    -f --force              Actually use API
    --fuzzy                 Override fuzzy translations [default: False]
    -a --api <api>          Deepl API token
"""

import os
import deepl
import polib

from docopt import docopt
from schema import Or, Schema, SchemaError


class Translator:

    def __init__(self, api, source, target, force):
        self.__translator = deepl.Translator(api)
        self.__source_lang = source
        self.__target_lang = target
        self.__force = force

    def translate(self, source_msg):
        if self.__force:
            return str(self.__translator.translate_text(
                source_msg,
                source_lang = self.__source_lang,
                target_lang = self.__target_lang
            ))
        return None


def process_file(filename, translator, fuzzy):
    po = polib.pofile(filename)
    print(f"File {filename} is {po.percent_translated()}% translated")
    for entry in po.untranslated_entries():
        if not entry.msgstr:
            newmsg = translator.translate(entry.msgid)
            if newmsg is not None:
                entry.msgstr = newmsg
                po.save(filename)
    if fuzzy:
        for entry in po.fuzzy_entries():
            newmsg = translator.translate(entry.msgid)
            if newmsg is not None:
                entry.msgstr = newmsg
                po.save(filename)
    po.save(filename)


if __name__ == '__main__':
    ARGS = docopt(__doc__, version='Deepl PO translator 1.0')
    SCHEMA = Schema({
        '--input': Or(os.path.isdir, os.path.isfile),
        '--source': str,
        '--target': str,
        '--api': str,
        '--force': bool,
        '--fuzzy': bool,
        '--help': Or(False),
        '--version': Or(False)
    })
    try:
        ARGS = SCHEMA.validate(ARGS)
    except SchemaError as err:
        exit(err)

    INPUT = ARGS['--input']
    DEEPL = Translator(
        ARGS['--api'], ARGS['--source'], ARGS['--target'], ARGS['--force'])

    if os.path.isfile(INPUT):
        process_file(INPUT, DEEPL, ARGS['--fuzzy'])
    elif os.path.isdir(INPUT):
        for root, dirs, files in os.walk(INPUT):
            for file in files:
                if file.endswith(".po"):
                    file_path = os.path.join(root, file)
                    process_file(file_path, DEEPL, ARGS['--fuzzy'])
    else:
        pass
