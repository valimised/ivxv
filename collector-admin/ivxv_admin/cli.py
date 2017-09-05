# IVXV Internet voting framework
"""
Helpers to implement CLI utilities for collector management service.
"""

import logging
import os
import sys
import textwrap

from docopt import docopt

from .ivxv_pkg import VERSION


def init_cli_util(docstr, allow_root=False):
    """
    Initialize command line utility.

    1. Validate command line arguments
    2. Set up logging

    Config file is searched "ivxv-collector-admin.conf" from the following
    locations:

    * current directory

    * /etc/ivxv

    * directory specified by environment variable IVXV_ADMIN_CONF

    * file specified by environment variable IVXV_ADMIN_CONF
    """
    # validate CLI arguments
    cli_args = docopt(textwrap.dedent(docstr), version=VERSION)

    # set up logging

    # check user rights
    if not allow_root and not os.getuid():
        log = logging.getLogger('ivxv_admin.admin_util')
        log.error('IVXV collector admin utils cannot be run as root')
        sys.exit(1)

    return cli_args
