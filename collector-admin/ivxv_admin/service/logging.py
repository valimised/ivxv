# IVXV Internet voting framework
"""Logging helper for microservice management."""

import logging

# create logger
log = logging.getLogger('.'.join(__name__.split('.')[:-1]))


class ServiceLogger(logging.LoggerAdapter):
    """Logger adapter for service object.

    Include service ID for logged messages.
    Include message level for levels other than INFO.
    """

    def process(self, msg, kwargs):
        """
        Process the logging message and keyword arguments passed in to
        a logging call to insert contextual information.
        """
        return f"SERVICE {self.extra['service_id']}: {msg}", kwargs

    def log(self, level, msg, *args, **kwargs):
        """
        Delegate a log call to the underlying logger, after adding
        contextual information from this adapter instance.
        """
        if self.isEnabledFor(level):
            if level != logging.INFO:
                msg = f"{logging.getLevelName(level).upper()}: {msg}"
            msg, kwargs = self.process(msg, kwargs)
            self.logger.log(level, msg, *args, **kwargs)
