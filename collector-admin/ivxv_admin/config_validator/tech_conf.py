# IVXV Internet voting framework
"""
Collector technical config validator.
"""

import re

from schematics.exceptions import ValidationError
from schematics.models import Model
from schematics.types import (BooleanType, IntType, ListType, ModelType,
                              StringType)

from .fields import CertificateType
from .schemas import protocol_cfg


class ServicesSchema(Model):
    """Validating schema for subservices config."""

    class BackupServiceSchema(Model):
        """Validating schema for backup subservice config."""
        id = StringType(regex=r'.+@.+', required=True)
        address = StringType(required=True)

    class ServiceSchema(Model):
        """Validating schema for subservice config."""
        id = StringType(regex=r'.+@.+', required=True)
        address = StringType(regex=r'.+:[0-9]+', required=True)
        peeraddress = StringType(regex=r'.+:[0-9]+')

    proxy = ListType(ModelType(ServiceSchema))
    mid = ListType(ModelType(ServiceSchema))
    smartid = ListType(ModelType(ServiceSchema))
    votesorder = ListType(ModelType(ServiceSchema))
    voting = ListType(ModelType(ServiceSchema))
    choices = ListType(ModelType(ServiceSchema))
    verification = ListType(ModelType(ServiceSchema))
    storage = ListType(ModelType(ServiceSchema))
    log = ListType(ModelType(ServiceSchema))
    backup = ListType(ModelType(BackupServiceSchema), max_size=1)


class CollectorTechnicalConfigSchema(Model):
    """Validating schema for collector technical config."""
    debug = BooleanType(default=False)
    snidomain = StringType(required=True)

    class FilterSchema(Model):
        """Validating schema for connection filter config."""

        class TLSFilterSchema(Model):
            """Validating schema for TLS connection filter config."""
            handshaketimeout = IntType(required=True, min_value=0)
            ciphersuites = ListType(
                StringType(choices=[
                    'TLS_RSA_WITH_AES_128_CBC_SHA',
                    'TLS_RSA_WITH_AES_256_CBC_SHA',
                    'TLS_RSA_WITH_AES_128_GCM_SHA256',
                    'TLS_RSA_WITH_AES_256_GCM_SHA384',
                    'TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA',
                    'TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA',
                    'TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA',
                    'TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA',
                    'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256',
                    'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256',
                    'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384',
                    'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384',
                ]),
                required=True)

        tls = ModelType(TLSFilterSchema, required=True)

        class CodecFilterSchema(Model):
            """Validating schema for codec connection filter config."""
            rwtimeout = IntType(required=True, min_value=0)
            requestsize = IntType(min_value=0)
            logrequests = BooleanType(default=False)

        codec = ModelType(CodecFilterSchema, required=True)

    filter = ModelType(FilterSchema, required=True)

    class SegmentSchema(Model):
        """Validating schema for network segment config."""
        id = StringType(required=True)
        services = ModelType(ServicesSchema, required=True)

    network = ListType(ModelType(SegmentSchema), required=True)

    class LogServerSchema(Model):
        """Validating schema for log collecting server config."""
        address = StringType(required=True)
        port = IntType(default=20514)

    logging = ListType(ModelType(LogServerSchema))

    class FileStorageServiceSchema(Model):
        """Validating schema for file storage service config."""
        wd = StringType(required=True)

    class EtcdStorageServiceSchema(Model):
        """Validating schema for etcd storage service config."""
        ca = CertificateType(required=True)
        conntimeout = IntType(required=True, min_value=0)
        optimeout = IntType(required=True, min_value=0)
        # FIXME: Compare to network.#.services.storage.#.id on first load.
        bootstrap = ListType(StringType(regex=r'.+@.+'), required=True)

    storage = protocol_cfg(
        {
            "file": FileStorageServiceSchema,
            "etcd": EtcdStorageServiceSchema,
        },
        required=True)

    class BackupTimeType(StringType):
        """Field validator for backup time."""

        def validate(self, value, context=None):
            if not re.match(r'[0-9]{2}:[0-9]{2}', value):
                raise ValidationError(f"Value must be in format HH:MM (not {value!r})")
            hour, minute = value.split(':')
            if int(hour) >= 24:
                raise ValidationError(f"Hour must be smaller than 24 (not {value!r})")
            if int(minute) >= 60:
                raise ValidationError(f"Minute must be smaller than 60 (not {value!r})")
            return super().validate(value, context)

    backup = ListType(BackupTimeType)
