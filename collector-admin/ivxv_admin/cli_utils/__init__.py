# IVXV Internet voting framework
"""CLI utilities."""

import logging
import os
import sys
import textwrap

from docopt import docopt

from .. import __version__
from ..config import CONFIG

assert CONFIG  # logging is configured by config module
log = logging.getLogger('ivxv_admin.admin_util')


def init_cli_util(docstr, allow_root=False):
    """
    Initialize command line utility.

    1. Validate command line arguments
    2. Set up logging

    Config file :file:`ivxv-collector-admin.conf` is searched from the
    following locations:

    * current directory

    * :file:`/etc/ivxv`

    * directory specified by environment variable :envvar:`IVXV_ADMIN_CONF`

    * file specified by environment variable :envvar:`IVXV_ADMIN_CONF`
    """
    # validate CLI arguments
    cli_args = docopt(textwrap.dedent(docstr), version=__version__)

    # check user rights
    if not allow_root and not os.getuid():
        log.error('IVXV collector admin utils cannot be run as root')
        sys.exit(1)

    return cli_args


def ask_user_confirmation(question):
    """Ask user confirmation.

    :return: True if user answers "Yes" or False if not.
    """
    while True:
        answer = input(question).upper()
        if answer in 'YN':
            return answer == 'Y'
