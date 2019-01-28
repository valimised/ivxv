#!/usr/bin/python3

import argparse
import subprocess
import urllib.parse

tbprocessed = []

DELIM = ':'
KIND_DIR = 'DIR'
KIND_SUBTITLE = 'SUBTITLE'
KIND_TITLE = 'TITLE'
KIND_IMPORT = 'IMPORT'
KIND_COMMENT = 'COMMENT'
KIND_EXAMPLE = 'EXAMPLE'
KIND_FUNC = 'FUNC'
KIND_DECL = 'DECL'
KIND_TYPE = 'TYPE'
KIND_TYPEMETHOD = 'TYPEMETHOD'
KIND_TYPEFUNC = 'TYPEFUNC'
KIND_CONTENTS = 'CONTENTS'
KIND_CONSTANTS = 'CONSTANTS'
KIND_VARIABLES = 'VARIABLES'

SUPPORTED = [KIND_DIR, KIND_SUBTITLE, KIND_TITLE, KIND_IMPORT, KIND_COMMENT,
             KIND_EXAMPLE, KIND_FUNC, KIND_DECL, KIND_TYPE,
             KIND_TYPEMETHOD, KIND_TYPEFUNC, KIND_CONTENTS,
             KIND_CONSTANTS, KIND_VARIABLES]


def decorate_title(_, item):
    return "Package %s" % item


def decorate_title2(kind, item):
    return "%s *%s*" % (kind.lower(), item)


def decorate_title3(kind, item):
    return kind.title()


def decorate_import(_, item):
    return 'import "%s"' % item


def decorate_contents(_, item):
    line1 = "\n.. contents::\n    :local:\n"
    line2 = ""
    if len(item) > 0:
        line2 = "    :depth: %s\n" % item
    return "%s%s" % (line1, line2)


def decorate_decl(_, item):
    line1 = "::\n\n%s" % item
    return line1


def decorate_escape(_, item):
    return item.replace('*', '\*').replace('\n', '')


def decorate_urldecode(kind, item):
    return decorate_escape(kind, urllib.parse.unquote_plus(item))


def decorate_comment(_, item):

    def is_header(hdr):
        for el in ['.', ':']:
            if el in hdr:
                return 0
        return len(hdr)

    def header_mode(hdr_mode, hdr_len, line):

        if not hdr_mode:
            return line == 0, 0, None

        if len(line) > 0:
            hdr_len = is_header(line)
            return hdr_len > 0, hdr_len, None

        if hdr_len > 0:
            hdr = "-" * hdr_len
            hdr += '\n'
            return True, 0, hdr
        return True, 0, None

    def pre_mode(pre_mode, line):
        ret = None
        if line.startswith((' ', '\t')) and not pre_mode:
            ret = "\n\n::\n\n"
            return True, ret
        return False, None

    ret = ""
    pre_mod = False
    hdr_mod = False
    hdr_len = 0
    for line in item.split('\n'):
        outline = line.replace('\\', '\\\\')
        pre_mod, pre = pre_mode(pre_mod, outline)
        if pre:
            ret += pre
            hdr_mod = False
            hdr_len = 0

        if not pre_mod:
            hdr_mod, hdr_len, hdr = header_mode(hdr_mod, hdr_len, outline)
            if hdr:
                ret += hdr
        ret += "%s\n" % outline

    return ret


def decorate_todo(kind, item):
    if len(item) > 0:
        raise Exception('Implement todo for "%s" "%s"' % (kind, item))
    return ""


def decorate_nop(kind, item):
    return item


KIND_OUTPUT_EASY = {
    KIND_TITLE: {
        'prefix': '-',
        'suffix': '-',
        'decorator': decorate_title
    },
    KIND_SUBTITLE: {
        'prefix': '',
        'suffix': '='
    },
    KIND_FUNC: {
        'suffix': '-',
        'decorator': decorate_title2
    },
    KIND_TYPEFUNC: {
        'suffix': '`',
        'decorator': decorate_escape
    },
    KIND_TYPEMETHOD: {
        'suffix': '`',
        'decorator': decorate_urldecode
    },
    KIND_TYPE: {
        'suffix': '-',
        'decorator': decorate_title2
    },
    KIND_DIR: {
    },
    KIND_IMPORT: {
        'decorator': decorate_import
    },
    KIND_COMMENT: {
        'decorator': decorate_comment
    },
    KIND_DECL: {
        'decorator': decorate_decl
    },
    KIND_EXAMPLE: {
        'decorator': decorate_todo
    },
    KIND_CONTENTS: {
        'decorator': decorate_contents
    },
    KIND_CONSTANTS: {
        'suffix': '-',
        'decorator': decorate_title3
    },
    KIND_VARIABLES: {
        'suffix': '-',
        'decorator': decorate_title3
    }
}


def get_ix(kind, ix, length):
    ret = ""
    if length > 0 and ix in KIND_OUTPUT_EASY[kind]:
        ret += '\n'
        ret += length * KIND_OUTPUT_EASY[kind][ix]
        ret += '\n'
    return ret


def generate_rst(kind, item):
    if kind is None:
        return ""

    decorator = decorate_todo
    if kind in KIND_OUTPUT_EASY:
        if 'decorator' in KIND_OUTPUT_EASY[kind]:
            decorator = KIND_OUTPUT_EASY[kind]['decorator']
        else:
            decorator = decorate_nop

    real_item = decorator(kind, item)
    length = len(real_item)
    prefix = get_ix(kind, 'prefix', length)
    suffix = get_ix(kind, 'suffix', length)
    return "\n%s%s%s\n" % (prefix, real_item, suffix)


parser = argparse.ArgumentParser(
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    description='Create reStructured Text from godoc')

parser.add_argument('-p', '--package', default='ivxv.ee',
                    help='Package to process')
parser.add_argument('-t', '--template', default='pseudomarkup',
                    help='Directory with compatible template')
parser.add_argument('-r', '--recursive', action='store_true',
                    help='Should we follow subpackages')
parser.add_argument('-s', '--section', action='store_true',
                    help='Generate first section')
parser.add_argument('-v', '--verbose', action='store_true',
                    help='Should we be verbose')
args = parser.parse_args()


tbprocessed.append(args.package)


if args.section:

    print('..  IVXV API')
    print('')
    print('==========')
    print('golang API')
    print('==========')

    print('-----------------------------------')
    print('Package %s' % args.package)
    print('-----------------------------------')


while len(tbprocessed) > 0:
    package = tbprocessed.pop()
    if args.verbose:
        print("Processing %s" % package)
    proc = subprocess.run(["godoc", "-templates", args.template, package],
                          stdout=subprocess.PIPE, universal_newlines=True,
                          check=True)
    markup = proc.stdout

    rstout = ''
    # Ignore occasional error message from godoc
    block = ''
    blockkind = None
    for line in markup.split('\n')[1:]:
        if len(line) == 0:
            if blockkind in [KIND_COMMENT]:
                block += '\n'
            continue
        tokens = line.split(DELIM, 1)
        kind = tokens[0]
        if kind not in SUPPORTED:
            continue
#            raise Exception('Unsupported item "%s"' % kind)

        item = tokens[1]
        if args.recursive and kind == KIND_DIR:
            tbprocessed.append("%s/%s" % (package, item))

        if kind != blockkind:
            rstblock = generate_rst(blockkind, block)
            print(rstblock, end='')
            rstout += rstblock
            block = item
            blockkind = kind
        else:
            block += '\n'
            block += item

    rstblock = generate_rst(blockkind, block)
    print(rstblock, end='')
    rstout += rstblock
