import os

import yaml
from pydantic import BaseModel
from pathlib import Path


def image_name_for_branch(name: str, branch: str) -> str:
    """
    Generate the image name for a given branch.

    :param name: Base name of the image
    :param branch: Branch name
    :return: Image name with branch suffix, or base name if branch is 'prod'
    """
    return f"{name}_{branch}" if branch != "prod" else name


class Credentials(BaseModel):
    """
    Model to store credentials for Docker registry access.
    """

    push_username: str
    pull_username: str
    push_password: str
    pull_password: str


def load_config(config_path: str) -> dict:
    """
    Load the YAML configuration file.

    :param config_path: Path to the configuration file
    :return: Configuration dictionary
    """
    if os.path.exists(config_path):
        with open(config_path, "r") as file:
            return yaml.safe_load(file)
    else:
        return {"provider": "azure", "azure": {"registry": "", "branches": {}}}


def get_root_absolute_path() -> Path:
    """
    Get the absolute path to the project root.

    :return: Absolute path to the project root
    """
    current_path = Path(__file__).resolve()
    while current_path.name != "PRISM":
        current_path = current_path.parent

    return current_path