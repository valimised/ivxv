# IVXV Internet voting framework
"""
Election config validator.
"""

# pylint: disable=no-self-use

from schematics.exceptions import ValidationError
from schematics.models import Model
from schematics.types import (
    BooleanType,
    DateTimeType,
    IntType,
    ListType,
    ModelType,
    StringType,
    URLType,
)

from .fields import CertificateType, ElectionIdType, PublicKeyType
from .schemas import ContainerSchema, OCSPSchema, TSPSchema, protocol_cfg


class ElectionConfigSchema(Model):
    """Validating schema for election config."""
    identifier = ElectionIdType(required=True)
    questions = ListType(ElectionIdType, min_size=1, required=True)

    class ElectionVerificationSchema(Model):
        """Validating schema for election verification config."""
        count = IntType(required=True, min_value=0)
        minutes = IntType(required=True, min_value=0)
        latestonly = BooleanType(default=False)

    verification = ModelType(ElectionVerificationSchema, required=True)

    class ElectionVotingSchema(Model):
        """Validating schema for election voting config."""
        ratelimitstart = IntType(default=0, min_value=0)
        ratelimitminutes = IntType(default=0, min_value=0)

        def validate_ratelimitminutes(self, data, value):
            """Validate rate limit."""
            try:
                if (data['ratelimitstart'] > 0
                        and data['ratelimitminutes'] == 0):
                    raise ValidationError(
                        'ratelimitstart set, but rate limiting disabled')
            except KeyError:
                pass  # error in data structure is catched later
            return value

    voting = ModelType(ElectionVotingSchema)

    class ElectionPeriodSchema(Model):
        """Validating schema for election period config."""
        servicestart = DateTimeType(required=True)
        electionstart = DateTimeType(required=True)
        electionstop = DateTimeType(required=True)
        servicestop = DateTimeType(required=True)

    period = ModelType(ElectionPeriodSchema, required=True)
    voterforeignehak = StringType(regex=r'^[0-9]{1,10}$')
    ignorevoterlist = StringType()

    class VoterListSchema(Model):
        """Validating schema for voter list updating service config."""
        key = PublicKeyType(required=True)

    voterlist = ModelType(VoterListSchema, required=True)

    class VisSchema(Model):
        """Validating schema for VIS service config."""

        url = URLType(required=True)
        ca = ListType(CertificateType)

    vis = ModelType(VisSchema, required=True)

    class AuthSchema(Model):
        """Validating schema for voter authentication config."""

        # FIXME: If service.mid exists, auth.ticket field must exist
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

    identity = StringType(
        required=True, choices=['commonname', 'serialnumber', 'pnoee'])

    class AgeSchema(Model):
        """Validating schema for voters age check config."""
        method = StringType(required=True, choices=['estpic'])
        timezone = StringType(required=True)
        limit = IntType(required=True, min_value=16)

    age = ModelType(AgeSchema)

    vote = ModelType(ContainerSchema, required=True)

    class MIDSchema(Model):
        """Validating schema for Mobile ID config."""
        url = URLType(required=True)
        relyingpartyuuid = StringType(required=True)
        relyingpartyname = StringType(required=True)
        language = StringType(
            required=True, choices=['EST', 'ENG', 'RUS', 'LIT'])
        authmessage = StringType(required=True, max_length=40)
        signmessage = StringType(required=True, max_length=40)
        messageformat = StringType(default='GSM-7', choices=['GSM-7', 'UCS-2'])
        authchallengesize = IntType(default=32, choices=[32, 48, 64])
        statustimeoutms = IntType()
        roots = ListType(CertificateType, required=True)
        intermediates = ListType(CertificateType)
        ocsp = ModelType(OCSPSchema)

        # pylint: disable=no-self-use
        def validate_phonerequired(self, data, value):
            """Validate phone/idcode required field."""
            try:
                if not data['idcoderequired'] and not data['phonerequired']:
                    raise ValidationError('Either idcoderequired or '
                                          'phonerequired must be true')
            except KeyError:
                pass  # error in data structure is catched later
            return value

    mid = ModelType(MIDSchema)

    qualification = ListType(
        protocol_cfg({
            "ocsp": OCSPSchema,
            "ocsptm": OCSPSchema,
            "tsp": TSPSchema,
            "tspreg": TSPSchema,
        }))

    # pylint: disable=unused-argument
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
                    raise ValidationError(f"Value {point1!r} is bigger than {point2!r}")
        except KeyError:
            pass  # error in data structure is catched later
        return value
