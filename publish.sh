#!/bin/bash -e

DOCKERHUB_IMAGE='cyberark/conjur-kubernetes-authenticator'
VERSION_TAG="$(<VERSION)"

docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE"
docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE:$VERSION_TAG"

docker push $DOCKERHUB_IMAGE
docker push "$DOCKERHUB_IMAGE:$VERSION_TAG"
