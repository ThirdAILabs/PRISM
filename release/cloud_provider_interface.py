from typing import Dict, List, Optional
import json
import re

import docker
from utils import Credentials
from abc import ABC, abstractmethod


class CloudProvider(ABC):
    def authorize_credentials(self, credentials: Credentials):
        self.credentials = credentials

    def get_local_image_digest(self, image_id: str):
        client = docker.from_env()
        image = client.images.get(image_id)
        digest = image.attrs["RootFS"]["Layers"]
        return digest

    def build_image(
        self,
        dockerfile_path: str,
        context_path: str,
        tag: str,
        nocache: bool,
        buildargs: Dict[str, str],
    ) -> str:
        """
        Build a Docker image from the specified path with the given tag.

        :param dockerfile_path: Path to the actual Dockerfile
        :param context_path: Path to the context used to build
        :param tag: Tag for the built image
        :param nocache: Whether to use cache during build
        :param buildargs: Build arguments for the Docker build
        :return: ID of the built image
        """
        print(f"Building image at path: {context_path} with tag: {tag}")
        docker_client = docker.APIClient(base_url="unix://var/run/docker.sock")
        generator = docker_client.build(
            path=context_path,
            dockerfile=dockerfile_path,
            tag=tag,
            rm=True,
            platform="linux/x86_64",
            nocache=nocache,
            buildargs=buildargs,
        )
        image_id: Optional[str] = None
        for chunk in generator:
            for minichunk in chunk.strip(b"\r\n").split(b"\r\n"):
                json_chunk = json.loads(minichunk)
                if "stream" in json_chunk:
                    print(json_chunk["stream"].strip())
                    match = re.search(
                        r"(^Successfully built |sha256:)([0-9a-f]+)$",
                        json_chunk["stream"],
                    )
                    if match:
                        image_id = match.group(2)
                if "errorDetail" in json_chunk:
                    raise RuntimeError(json_chunk["errorDetail"]["message"])
        if not image_id:
            raise RuntimeError(f"Did not successfully build {tag} from {context_path}")

        print(f"\nLocal: Built {image_id}\n")
        print("\n===============================================================\n")

        return image_id

    @abstractmethod
    def push_image(self, image_id: str, tag: str) -> None:
        pass

    @abstractmethod
    def get_image_digest(self, name: str, tag: str) -> List[str]:
        pass

    @abstractmethod
    def delete_image(self, repository: str, tag: str, **kwargs) -> None:
        pass

    @abstractmethod
    def create_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> Dict[str, str]:
        pass

    @abstractmethod
    def update_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> None:
        pass

    @abstractmethod
    def update_image(self, image_id: str, name: str, tag: str) -> None:
        pass

    @abstractmethod
    def get_registry_name(self) -> str:
        pass

    @abstractmethod
    def get_full_image_name(self, base_name: str, branch: str, tag: str) -> str:
        pass
