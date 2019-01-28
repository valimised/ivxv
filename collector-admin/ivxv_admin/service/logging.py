# IVXV Internet voting framework
"""Logging helper for microservice management."""

import datetime
import logging

from .. import RFC3339_DATE_FORMAT_WO_FRACT

# create logger
log = logging.getLogger('.'.join(__name__.split('.')[:-1]))


class ServiceLogger:
    """Logger wrapper for service object.

    Prepend service ID for logged messages.
    Message level is also prepended for levels other than INFO.
    """
    log_prefix = None  #: Prefix for log messages (str)
    storage = None  #: Collection for logged records (list of strings)

    def __init__(self, service_id):
        """Constructor."""
        self.log_prefix = 'SERVICE %s: ' % service_id
        self.storage = []

    def __getattr__(self, name):
        """Get wrapper for logger method."""
        log_method = getattr(log, name)

        def log_method_wrapper(*args):
            """Wrapper for logger method to prepend log message prefix."""
            args = list(args)
            if name != 'info':
                args[0] = '{}: {}'.format(name.upper(), args[0])
            args[0] = self.log_prefix + args[0]
            self.storage.append('{} {}'.format(
                datetime.datetime.now().strftime(RFC3339_DATE_FORMAT_WO_FRACT),
                args[0] % tuple(args[1:])))
            return log_method(*tuple(args))

        return log_method_wrapper
