# IVXV Internet voting framework
"""
Choices list validator.
"""

from schematics.exceptions import DataError
from schematics.models import Model
from schematics.types import DictType, StringType

from .fields import ElectionIdType


class ChoicesListSchema(Model):
    """Validating schema for choices list."""
    election = ElectionIdType(required=True)
    choices = DictType(DictType(DictType(StringType)))

    def validate(self, partial=False, convert=True, app_data=None, **kwargs):
        """Validate model."""
        super().validate(partial, convert, app_data, **kwargs)

        choices = []
        for district_choices in self.choices.values():
            for choice in district_choices.values():
                for choice_id in choice.keys():
                    if choice_id in choices:
                        raise DataError(
                            {'choices': f'Duplicate choice ID: {choice_id}'})
                    choices.append(choice_id)
