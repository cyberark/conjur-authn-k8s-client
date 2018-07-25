#!/bin/bash -e

DOCKERHUB_IMAGE_OLD='cyberark/conjur-kubernetes-authenticator'  # backwards-compatible image name
DOCKERHUB_IMAGE_NEW='cyberark/conjur-authn-k8s-client'
VERSION_TAG="$(<VERSION)"

REDHAT_IMAGE='scan.connect.redhat.com/ospid-1c46a2de-1d88-40e6-a433-7114ad0099cb/conjur-openshift-authenticator-client'

docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE_OLD"
docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE_OLD:$VERSION_TAG"

docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE_NEW"
docker tag conjur-authn-k8s-client:dev "$DOCKERHUB_IMAGE_NEW:$VERSION_TAG"

docker push $DOCKERHUB_IMAGE_OLD
docker push "$DOCKERHUB_IMAGE_OLD:$VERSION_TAG"

docker push $DOCKERHUB_IMAGE_NEW
docker push "$DOCKERHUB_IMAGE_NEW:$VERSION_TAG"

docker tag conjur-authn-k8s-client:dev-redhat "$REDHAT_IMAGE"
docker tag conjur-authn-k8s-client:dev-redhat "$REDHAT_IMAGE:$VERSION_TAG"

docker push "$REDHAT_IMAGE:$VERSION_TAG" || { echo 'Red Hat push FAILED!'; exit 0; }  # you can't push the same tag twice to redhat registry, so ignore errors
