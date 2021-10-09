# IVXV Internet voting framework
"""
Districts list validator.
"""

from schematics.models import Model
from schematics.types import DictType, ListType, ModelType, StringType

from .fields import ElectionIdType


class DistrictsListSchema(Model):
    """Validating schema for districts list."""

    class DistrictSchema(Model):
        """Validating schema for district record config."""
        name = StringType(required=True)
        parish = ListType(StringType, required=True)

    class RegionSchema(Model):
        """Validating schema for region record config."""
        state = StringType()
        county = StringType()
        parish = StringType()

    election = ElectionIdType(required=True)
    districts = DictType(ModelType(DistrictSchema))
    regions = DictType(ModelType(RegionSchema))
    counties = DictType(ListType(StringType))
