# IVXV Internet voting framework
"""
Validator for collector config files and voting lists.
"""

# pylint: disable=invalid-name,no-self-use,unused-argument

import re
import string

from schematics.models import Model
from schematics.types import (BooleanType, DateTimeType, DictType, IntType,
                              ListType, ModelType, PolyModelType, StringType,
                              URLType)
from schematics.exceptions import DataError, ValidationError

# don't use "from . import *" as this module can be used without namespace
from ivxv_admin import ROLES


# fields
# FIXME: constrainments on schemas passed to wrapper aren't enforced.
# fields can  be missing from the conf, type constrainments are ignored,
# without raising any errors (e.g missing storage.conf.ca,
# invalid URL in qualification.*.url)
def protocol_conf(mapping, **kwargs):
    """
    Return an alternative protocol configuration type.

    Alternative protocol configurations allow only one from a selection of
    alternatives to be used. The model looks like this::

        protocol: <protocol name>
        conf: <configuration for the chosen protocol>

    schematics has PolyModelType which allows "conf" to be different models,
    but it must select the model to use based solely on the contents of "conf",
    which cannot be done with this construction. So create a new wrapper model
    for each protocol configuration model which also includes the "protocol"
    field and let PolyModelType choose from those.
    """

    models = []
    for protocol, model in mapping.items():
        wrapper = type(model.__name__ + "Wrapper", (Model,), {
            "protocol": StringType(required=True, choices=[protocol]),
            "conf": ModelType(model, required=True),
        })
        mapping[protocol] = wrapper
        models.append(wrapper)

    def _claim(_, data):
        return mapping.get(data.get("protocol"))

    return PolyModelType(models, claim_function=_claim, **kwargs)


class PEMType(StringType):
    """A field that validates input as PEM-encoding."""

    MESSAGES = {
        "pem": "Not a well-formed PEM encoding.",
        "block": "Wrong block type '{0}', want '{1}'."
    }

    PEM_REGEX = re.compile(r"^-----BEGIN ([A-Z ]+)-----\n"
                           r"([a-zA-Z0-9+/\n]+)(=(\n?=)?)?\n"
                           r"-----END \1-----\s*$")

    def __init__(self, block_type, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.block_type = block_type

    def validate_pem(self, value):
        """Validates the PEM and block type."""
        match = PEMType.PEM_REGEX.match(value)
        if not match:
            raise ValidationError(self.messages['pem'])
        if match.group(1) != self.block_type:
            raise ValidationError(self.messages['block'].format(
                match.group(1), self.block_type))


class CertificateType(PEMType):
    """A field that validates input as a PEM-encoded certificate."""
    def __init__(self, *args, **kwargs):
        super().__init__("CERTIFICATE", *args, **kwargs)


class PublicKeyType(PEMType):
    """A field that validates input as a PEM-encoded public key."""
    def __init__(self, *args, **kwargs):
        super().__init__("PUBLIC KEY", *args, **kwargs)


class ElectionIdType(StringType):
    """
    Election ID field validator.
    """

    def validate(self, value, context=None):
        """Validate field."""
        min_length = 1
        max_length = 28

        if len(value) < min_length or len(value) > max_length:
            raise ValidationError(
                'ID length must be between {} and {}'
                .format(min_length, max_length))
        if any(ws in value for ws in list(string.whitespace)):
            raise ValidationError('Election ID contains whitespace')

        return super().validate(value, context)


# schemas

class OCSPSchema(Model):
    """Validating schema for OCSP config."""
    url = URLType(required=True)
    responders = ListType(CertificateType)


class OCSPSchemaNoURL(Model):
    """Validating schema for OCSP config."""
    responders = ListType(CertificateType)


class BDocSchema(Model):
    """Validating schema for BDoc config."""
    bdocsize = IntType(required=True, min_value=1)
    filesize = IntType(required=True, min_value=1)
    roots = ListType(CertificateType, required=True)
    intermediates = ListType(CertificateType)
    checktimemark = BooleanType(required=True)
    ocsp = ModelType(OCSPSchema)


class DummySchema(Model):
    """Validating schema for Dummy container config."""
    trusted = ListType(StringType)


class ContainerSchema(Model):
    """Validating schema for signed container config."""
    bdoc = ModelType(BDocSchema)
    dummy = ModelType(DummySchema)


class TrustRootSchema(Model):
    """Validating schema for trust root config."""
    container = ModelType(ContainerSchema, required=True)
    authorizations = ListType(StringType, required=True)


class ServiceSchema(Model):
    """Validating schema for subservice config."""
    id = StringType(regex=r'.+@.+', required=True)
    address = StringType(regex=r'.+:[0-9]+', required=True)
    peeraddress = StringType(regex=r'.+:[0-9]+')


class ServicesSchema(Model):
    """Validating schema for subservices config."""
    proxy = ListType(ModelType(ServiceSchema))
    dds = ListType(ModelType(ServiceSchema))
    voting = ListType(ModelType(ServiceSchema))
    choices = ListType(ModelType(ServiceSchema))
    verification = ListType(ModelType(ServiceSchema))
    storage = ListType(ModelType(ServiceSchema))
    log = ListType(ModelType(ServiceSchema))


class CollectorTechnicalConfigSchema(Model):
    """Validating schema for collector technical config."""
    debug = BooleanType(default=False)

    class VoterListSchema(Model):
        """Validating schema for voter list updating service config."""
        key = PublicKeyType(required=True)

    voterlist = ModelType(VoterListSchema, required=True)

    class AuthSchema(Model):
        """Validating schema for voter authentication config."""

        # FIXME: If service.dds exists, auth.ticket field must exist
        class TicketAuthSchema(Model):
            """Validating schema for ticket authentication config."""

        ticket = ModelType(TicketAuthSchema)

        class TLSAuthSchema(Model):
            """Validating schema for TLS authentication config."""
            roots = ListType(CertificateType, required=True)
            intermediates = ListType(CertificateType)
            ocsp = ModelType(OCSPSchema)

        tls = ModelType(TLSAuthSchema)

    auth = ModelType(AuthSchema, required=True)

    identity = StringType(required=True,
                          choices=['commonname', 'serialnumber'])

    class AgeSchema(Model):
        """Validating schema for voters age check config."""
        method = StringType(required=True, choices=['estpic'])
        timezone = StringType(required=True)
        limit = IntType(required=True, min_value=16)

    age = ModelType(AgeSchema)

    vote = ModelType(ContainerSchema, required=True)

    class FilterSchema(Model):
        """Validating schema for connection filter config."""
        class TLSFilterSchema(Model):
            """Validating schema for TLS connection filter config."""
            handshaketimeout = IntType(required=True, min_value=0)

        tls = ModelType(TLSFilterSchema, required=True)

        class CodecFilterSchema(Model):
            """Validating schema for codec connection filter config."""
            rwtimeout = IntType(required=True, min_value=0)

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
        port = IntType(default=514)
        protocol = StringType(required=True, choices=['udp', 'tcp'],
                              default='udp')

    logging = ListType(ModelType(LogServerSchema))

    class DDSSchema(Model):
        """Validating schema for DDS config."""
        url = URLType(required=True)
        language = StringType(required=True,
                              choices=['EST', 'ENG', 'RUS', 'LIT'])
        servicename = StringType(required=True, max_length=20)
        authmessage = StringType(required=True, max_length=40)
        signmessage = StringType(required=True, max_length=40)
        roots = ListType(CertificateType, required=True)
        intermediates = ListType(CertificateType)
        ocsp = ModelType(OCSPSchemaNoURL, required=True)

    dds = ModelType(DDSSchema)

    class FileStorageServiceSchema(Model):
        """Validating schema for file storage service config."""
        wd = StringType(required=True)

    class EtcdStorageServiceSchema(Model):
        """Validating schema for etcd storage service config."""
        ca = CertificateType(required=True)
        conntimeout = IntType(required=True, min_value=0)
        optimeout = IntType(required=True, min_value=0)

    storage = protocol_conf({
        "file": FileStorageServiceSchema,
        "etcd": EtcdStorageServiceSchema,
    }, required=True)

    class TSPSchema(Model):
        """Validating schema for timestamp protocol config."""
        url = URLType(required=True)
        signers = ListType(CertificateType, required=True)
        delaytime = IntType(required=True, min_value=0)

    qualification = ListType(protocol_conf({
        "ocsp": OCSPSchema,
        "ocsptm": OCSPSchema,
        "tsp": TSPSchema,
        "tspreg": TSPSchema,
    }))

    class StatsServiceSchema(Model):
        """Validating schema for stats service config."""
        url = URLType(required=True)
        tlscert = CertificateType(required=True)

    stats = ModelType(StatsServiceSchema)


class ElectionConfigSchema(Model):
    """Validating schema for election config."""
    identifier = ElectionIdType(required=True)
    questions = ListType(ElectionIdType, min_size=1, required=True)

    class ElectionVerificationSchema(Model):
        """Validating schema for election verification config."""
        count = IntType(required=True, min_value=0)
        minutes = IntType(required=True, min_value=0)

    verification = ModelType(ElectionVerificationSchema, required=True)

    class ElectionPeriodSchema(Model):
        """Validating schema for election period config."""
        servicestart = DateTimeType(required=True)
        electionstart = DateTimeType(required=True)
        electionstop = DateTimeType(required=True)
        servicestop = DateTimeType(required=True)

    period = ModelType(ElectionPeriodSchema, required=True)
    ignorevoterlist = StringType()

    def validate_questions(self, data, value):
        """Validate question field."""
        if value and len(value) > len(set(value)):
            raise ValidationError('Election questions must be unique')
        return value

    def validate_period(self, data, value):
        """Validate election period."""
        try:
            for point1, point2 in [['servicestart', 'electionstart'],
                                   ['electionstart', 'electionstop'],
                                   ['electionstop', 'servicestop']]:
                if data['period'][point1] >= data['period'][point2]:
                    raise ValidationError('Value "%s" is bigger than "%s"' %
                                          (point1, point2))
        except KeyError:
            pass  # error in data structure is catched later
        return value


class ChoicesListSchema(Model):
    """Validating schema for choices list."""
    election = ElectionIdType(required=True)
    choices = DictType(DictType(DictType(StringType)))


class VotersListSchema(Model):
    """Validating schema for voters list."""
    election = ElectionIdType(required=True)
    version = StringType(required=True, regex='^1$')
    list_type = StringType(required=True, regex='^(algne|muudatused)$')

    class VoterArraySchema(ListType):
        """Validation type for voter array."""

        def validate(self, value, context=None):
            """Validate field."""
            if not re.match(r'[0-9]{11}$', value[0]):
                raise ValidationError('Invalid ID code')
            if not re.match(r'.+ .+$', value[1]):
                raise ValidationError('Invalid person name')
            if not re.match(r'(lisamine|kustutamine)$', value[2]):
                raise ValidationError('Invalid action')
            if not re.match(r'.+$', value[3]):
                raise ValidationError('Invalid station part (1)')
            if not re.match(r'.+$', value[4]):
                raise ValidationError('Invalid station part (2)')
            if not re.match(r'[0-9]*$', value[5]):
                raise ValidationError('Invalid line number')
            return super().validate(value, context)

    voters = ListType(VoterArraySchema(StringType, min_size=7, max_size=7),
                      required=True)


class UserManagementCommandSchema(Model):
    """Validating schema for user management command."""
    action = StringType(required=True, regex='user-permissions')
    cn = StringType(required=True, regex='.+,.+,[0-9]{11}')
    roles = ListType(StringType(), required=True)

    def validate(self, partial=False, convert=True, app_data=None, **kwargs):
        """Validate model."""
        super().validate(partial, convert, app_data, **kwargs)

        # role list checks
        if not isinstance(self.roles, list):
            raise DataError({'roles': 'Value is not a list'})
        # pylint: disable=not-an-iterable
        for role in self.roles:
            if role not in ROLES:
                raise DataError({'roles': 'Unknown role "%s"' % role})
        if len(self.roles) != len(set(self.roles)):
            raise DataError({'roles': 'Duplicate roles'})
        # pylint: disable=unsupported-membership-test
        if 'none' in self.roles and len(self.roles) > 1:
            raise DataError(
                {'roles': 'Role "none" can\'t be used with orher roles'})


def validate_config(config, schema):
    """Validate collector config."""
    if schema == 'voters':
        lines = config.split('\n')
        data = dict(version=lines.pop(0))
        data['election'] = lines.pop(0)
        data['list_type'] = lines.pop(0)
        data['voters'] = [line.strip().split('\t') for line in lines]
    else:
        data = config

    if not isinstance(data, dict):
        raise ValueError('Configuration data is not dictionary')

    # detect config schema
    schemas = {
        'trust': TrustRootSchema,
        'technical': CollectorTechnicalConfigSchema,
        'election': ElectionConfigSchema,
        'choices': ChoicesListSchema,
        'voters': VotersListSchema,
        'user': UserManagementCommandSchema,
    }
    validator = schemas[schema](data)
    validator.validate()

    return validator.to_primitive()


def main(args):
    """
    Validate config file data structure.

    This utility is used to validate plain YAML file.

    Usage: config_validator <type> FILE

    Options:
        <type>              Config type. Possible values are:
                            - election: election config
                            - technical: collector technical config
                            - trust: trust root config
    """
    import logging
    log = logging.getLogger(__name__)
    logging.basicConfig(level=logging.DEBUG)

    config_type = args['<type>']
    if config_type not in ['election', 'technical', 'trust']:
        log.error('Invalid config type "%s"', args['<type>'])
        return 1

    from ivxv_admin.command_file import load_yaml_file
    with open(args['FILE']) as fp:
        log.info('Loading config file %s', args['FILE'])
        try:
            config = load_yaml_file(fp)
        except FileNotFoundError as err:
            log.error('File %s not found', err.filename)
            return 1
    log.info('Config file is validated as YAML')

    validate_config(config, config_type)
    log.info('Config file %s is a valid %s config', args['FILE'], config_type)


if __name__ == '__main__':
    from docopt import docopt
    exit(main(docopt(main.__doc__)))
