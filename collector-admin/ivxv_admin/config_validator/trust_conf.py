# IVXV Internet voting framework
"""
Collector trust root config validator.
"""

from schematics.models import Model
from schematics.types import ListType, ModelType, StringType

from .schemas import ContainerSchema


class TrustRootSchema(Model):
    """Validating schema for trust root config."""
    container = ModelType(ContainerSchema, required=True)
    authorizations = ListType(StringType, required=True)
