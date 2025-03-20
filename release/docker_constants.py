from dataclasses import dataclass

@dataclass
class Image:
    """
    Represents a Docker image configuration for building and deployment.

    Attributes:
        key (str): The environment var to store the final image name.
        name (str): The name of the image, used for tagging and identification.
        dockerfile_path (str): The path to the Dockerfile, relative to the context path.
        context_path (str): The build context path for Docker, typically the directory containing the Dockerfile and necessary files.
    """

    key: str
    name: str
    dockerfile_path: str
    context_path: str


images_to_build = [
    Image(
        key="prism_backend",
        name="prism_backend",
        dockerfile_path="Dockerfile",
        context_path=".",
    ),
    Image(
        key="prism_frontend",
        name="prism_frontend",
        dockerfile_path="Dockerfile",
        context_path="frontend",
    ),
    Image(
        key="keycloak",
        name="keycloak",
        dockerfile_path="Dockerfile",
        context_path="keycloak",
    )
]


images_to_pull_from_private = [
    "grobid",
    "traefik",
]
