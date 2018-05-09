#!/bin/bash -e

DOCKERHUB_IMAGE='cyberark/conjur-kubernetes-authenticator'
VERSION_TAG="$(<VERSION)"

docker tag "$(cat AUTHENTICATOR_TAG)" "$DOCKERHUB_IMAGE"
docker tag "$(cat AUTHENTICATOR_TAG)" "$DOCKERHUB_IMAGE:$VERSION_TAG"

docker push $DOCKERHUB_IMAGE
docker push "$DOCKERHUB_IMAGE:$VERSION_TAG"
