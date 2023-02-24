# IVXV Internet voting framework
"""
Common field validator classes.
"""

# pylint: disable=too-few-public-methods

import string

import OpenSSL
from schematics.exceptions import ValidationError
from schematics.types import StringType


class ElectionIdType(StringType):
    """Election ID field validator."""

    def validate(self, value, context=None):
        """Validate field."""
        min_length = 1
        max_length = 28

        if len(value) < min_length or len(value) > max_length:
            raise ValidationError(
                f"Election ID length must be between {min_length} and {max_length}"
            )
        if any(ws in value for ws in list(string.whitespace)):
            raise ValidationError('Election ID contains whitespace')

        return super().validate(value, context)


class CertificateType(StringType):
    """A field that validates input as a PEM-encoded certificate."""

    def validate(self, value, context=None):
        """Validate field."""
        try:
            OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM, value)
        except OpenSSL.crypto.Error as err:
            err_lib, err_func, err_reason = err.args[0][0]
            raise ValidationError(
                f"Error in {err_lib} library {err_func} function: {err_reason}"
            )

        return super().validate(value, context)


class PublicKeyType(StringType):
    """A field that validates input as a PEM-encoded public key."""

    def validate(self, value, context=None):
        """Validate field."""
        try:
            OpenSSL.crypto.load_publickey(OpenSSL.crypto.FILETYPE_PEM, value)
        except OpenSSL.crypto.Error as err:
            err_lib, err_func, err_reason = err.args[0][0]
            raise ValidationError(
                f"Error in {err_lib} library {err_func} function: {err_reason}"
            )

        return super().validate(value, context)
