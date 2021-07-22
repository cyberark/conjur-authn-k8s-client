#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

function main {
  if [[ "$CLUSTER_TYPE" == "gke" ]]; then
    gcloud auth activate-service-account \
      --key-file "$GCLOUD_SERVICE_KEY"
    gcloud container clusters get-credentials "$GCLOUD_CLUSTER_NAME" \
      --zone "$GCLOUD_ZONE" \
      --project "$GCLOUD_PROJECT_NAME"
    docker login "$DOCKER_REGISTRY_URL" \
      -u oauth2accesstoken \
      -p "$(gcloud auth print-access-token)"
  elif [[ "$CLUSTER_TYPE" == "oc" ]]; then
    oc login "$OPENSHIFT_URL" \
      --username="$OPENSHIFT_USERNAME" \
      --password="$OPENSHIFT_PASSWORD" \
      --insecure-skip-tls-verify=true
    docker login \
      -u _ -p "$(oc whoami -t)" \
      "$DOCKER_REGISTRY_PATH"
  fi
}

main
