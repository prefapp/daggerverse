from typing import Annotated
from dagger import Doc, function, File
import jsonschema
import yaml
from typing import Annotated

REGISTRY_SCHEMA = {
    "type": "object",
    "required": ["name", "repository", "auth_strategy"],
    "properties": {
        "name": {
            "type": "string"
        },
        "repository": {
            "type": "string"
        },
        "auth_strategy": {
            "type": "string",
            "enum": ["aws_oidc", "azure_oidc", "generic"]
        },
    },
}

FLAVOR_SCHEMA = {
    "type": "object",
    "required": ["dockerfile"],
    "properties": {
        "build_args": {
            "type": "object",
            "patternProperties": {
                ".*": {"type": "string"}
            },
        },
        "dockerfile": {
            "type": "string"
        },
        "extra_registries": {
            "type": "array",
            "items": REGISTRY_SCHEMA
        },
    },
}

SCHEMA = {
    "type": "object",
    "required": ["snapshots", "releases"],
    "properties": {
        "snapshots": {
            "type": "object",
            "required": ["default"],
            "properties": {
                "default": FLAVOR_SCHEMA,
            },
            "patternProperties": {
                ".*": FLAVOR_SCHEMA
            }
        },
        "releases": {
            "type": "object",
            "required": ["default"],
            "properties": {
                "default": FLAVOR_SCHEMA,
            },
            "patternProperties": {
                ".*": FLAVOR_SCHEMA
            }
        },

    }
}


@ function
async def validate(
    self,
    config: Annotated[File, Doc("Configuration file.")],
) -> None:
    """Validate the configuration file."""

    config_data = yaml.safe_load(await config.contents())

    jsonschema.validate(config_data, SCHEMA)

    return
