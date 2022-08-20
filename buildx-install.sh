#!/bin/bash

LATEST=$(wget -qO- https://api.github.com/repos/docker/buildx/releases/latest | jq -r .name)
wget --no-verbose https://github.com/docker/buildx/releases/download/$LATEST/buildx-$LATEST.linux-amd64
chmod 555 buildx-$LATEST.linux-amd64
mkdir -p ~/.docker/cli-plugins
mv buildx-$LATEST.linux-amd64 ~/.docker/cli-plugins/docker-buildx
docker buildx version
