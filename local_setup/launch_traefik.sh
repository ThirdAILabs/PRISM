#!/bin/bash

curr_dir=$(dirname "$(realpath "$0")")

ARGS=(
    "--api.dashboard=true"
    "--api.insecure=true"
    "--entrypoints.web.address=:80"
    "--entrypoints.web.forwardedHeaders.insecure=true"
    "--entrypoints.traefik.address=:8080"
    "--providers.file.filename=$curr_dir/traefik_config/dynamic-conf.yaml"
    "--providers.file.watch=true"
    "--log.level=INFO"
    "--providers.http.pollInterval=5s"
)

traefik "${ARGS[@]}"
