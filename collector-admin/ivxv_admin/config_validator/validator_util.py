# IVXV Internet voting framework
"""Config validator utility for collector config files and voting lists."""

from .choices_list import ChoicesListSchema
from .districts_list import DistrictsListSchema
from .election_conf import ElectionConfigSchema
from .tech_conf import CollectorTechnicalConfigSchema
from .trust_conf import TrustRootSchema
from .user_management import UserManagementCommandSchema
from .voters_list import parse_voters_list


def validate_cfg(cfg, schema_name):
    """Validate collector config.

    :param cfg: Configuration data structure
    :type cfg: dict

    :raises: ValueError
    :raises: schematics.exceptions.DataError
    """
    # validate voters list
    if schema_name == 'voters':
        return parse_voters_list(cfg)

    if not isinstance(cfg, dict):
        raise ValueError('Configuration data is not dictionary')

    # detect config schema
    schemas = {
        'trust': TrustRootSchema,
        'technical': CollectorTechnicalConfigSchema,
        'election': ElectionConfigSchema,
        'choices': ChoicesListSchema,
        'districts': DistrictsListSchema,
        'user': UserManagementCommandSchema,
    }
    validator = schemas[schema_name](cfg.copy())
    validator.validate()

    return validator.to_primitive()
