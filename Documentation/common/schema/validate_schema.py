# IVXV Internet voting framework
"""Validate jsonschema example files."""

import json
import logging
import os
import sys

import jsonschema

log = logging.getLogger()
logging.basicConfig(level=logging.INFO)


def main():
    """Main routine."""
    for schema_filename in sorted(os.listdir(".")):
        if schema_filename.endswith(".schema"):
            log.info("Loading schema %s", schema_filename)
            with open(schema_filename) as fd:
                schema = json.load(fd)
            log.info("Schema file %s is valid", schema_filename)

            example_filename = f"{schema_filename}.example"
            log.info("Validating file %s", example_filename)
            with open(example_filename) as fd:
                example = json.load(fd)

            try:
                jsonschema.validate(instance=example, schema=schema)
            except jsonschema.exceptions.ValidationError as err:
                log.error(
                    "Schema validation error while validating %r: %s",
                    example_filename,
                    err,
                )
                return 1
            log.info("File %s is valid", example_filename)

    return 0


if __name__ == "__main__":
    sys.exit(main())
