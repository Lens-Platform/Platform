#! /usr/bin/env sh

set -e

# build the docker file
GIT_COMMIT=$(git rev-list -1 HEAD) && \
DOCKER_BUILDKIT=1 docker build --tag test/authentication_handler_service --build-arg "REVISION=${GIT_COMMIT}" .
