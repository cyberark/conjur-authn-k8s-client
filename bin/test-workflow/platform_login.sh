#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

source utils.sh

if [[ "$CONJUR_PLATFORM" == "gke" || "$APP_PLATFORM" == "gke" ]]; then
  check_env_var GCLOUD_SERVICE_KEY
  check_env_var GCLOUD_CLUSTER_NAME
  check_env_var GCLOUD_ZONE
  check_env_var GCLOUD_PROJECT_NAME
fi

if [[ "$CONJUR_PLATFORM" == "openshift" || "$APP_PLATFORM" == "openshift" ]]; then
  check_env_var CONJUR_PLATFORM
  check_env_var APP_PLATFORM
  check_env_var OPENSHIFT_URL
  check_env_var OPENSHIFT_USERNAME
  check_env_var OPENSHIFT_PASSWORD
  check_env_var DOCKER_REGISTRY_PATH
fi

function main {
  if [[ "$CONJUR_PLATFORM" == "gke" || "$APP_PLATFORM" == "gke" ]]; then
    gcloud auth activate-service-account \
      --key-file "$GCLOUD_SERVICE_KEY"
    gcloud container clusters get-credentials "$GCLOUD_CLUSTER_NAME" \
      --zone "$GCLOUD_ZONE" \
      --project "$GCLOUD_PROJECT_NAME"
    docker login "$DOCKER_REGISTRY_URL" \
      -u oauth2accesstoken \
      -p "$(gcloud auth print-access-token)"
  elif [[ "$CONJUR_PLATFORM" == "openshift" || "$APP_PLATFORM" == "openshift" ]]; then
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
