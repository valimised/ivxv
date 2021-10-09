# IVXV Internet voting framework
"""Config validator utility for collector config files and voting lists."""

import json
import os

import jsonschema
import pkg_resources

from .choices_list import ChoicesListSchema
from .districts_list import DistrictsListSchema
from .election_conf import ElectionConfigSchema
from .tech_conf import CollectorTechnicalConfigSchema
from .trust_conf import TrustRootSchema
from .user_management import UserManagementCommandSchema
from .voters_list import VoterListChangesetSkipSchema, parse_voters_list


def validate_cfg(cfg, schema_name):
    """Validate collector config.

    :param cfg: Configuration data structure
    :type cfg: dict

    :raises ValueError:
    :raises schematics.exceptions.DataError:
    """
    # validate choices and districts list with jsonschema
    if schema_name in ['choices', 'districts']:
        jsonschema_src = pkg_resources.resource_string(
            'ivxv_admin',
            os.path.join('jsonschema', f'ivxv.{schema_name}.schema'))
        schema = json.loads(jsonschema_src.decode('UTF-8'))
        try:
            jsonschema.validate(instance=cfg, schema=schema)
        except jsonschema.exceptions.ValidationError as err:
            raise ValueError(err)

    # validate voters list
    if schema_name == "voters" and isinstance(cfg, str):
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
        "voters": VoterListChangesetSkipSchema,
    }
    validator = schemas[schema_name](cfg.copy())
    validator.validate()

    return validator.to_primitive()
