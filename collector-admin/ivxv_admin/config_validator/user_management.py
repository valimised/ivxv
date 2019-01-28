# IVXV Internet voting framework
"""
Collector user management command validator.
"""

from schematics.exceptions import DataError
from schematics.models import Model
from schematics.types import ListType, StringType

from .. import USER_ROLES


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
            if role not in USER_ROLES:
                raise DataError({'roles': 'Unknown role "%s"' % role})
        if len(self.roles) != len(set(self.roles)):
            raise DataError({'roles': 'Duplicate roles'})
        # pylint: disable=unsupported-membership-test
        if 'none' in self.roles and len(self.roles) > 1:
            raise DataError({
                'roles':
                'Role "none" can\'t be used with other roles'
            })
