#!/bin/bash -e

DOCKERHUB_IMAGE='cyberark/conjur-kubernetes-authenticator'
VERSION_TAG="$(<VERSION)"

REDHAT_IMAGE='scan.connect.redhat.com/ospid-1c46a2de-1d88-40e6-a433-7114ad0099cb/conjur-openshift-authenticator-client'

docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE"
docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE:$VERSION_TAG"

docker tag conjur-authn-k8s-client:dev-redhat "$REDHAT_IMAGE"
docker tag conjur-authn-k8s-client:dev-redhat "$REDHAT_IMAGE:$VERSION_TAG"

docker push $DOCKERHUB_IMAGE
docker push "$DOCKERHUB_IMAGE:$VERSION_TAG"
docker push "$REDHAT_IMAGE:$VERSION_TAG"
