#!/bin/bash

set -eo pipefail

. utils.sh

if [[ "$PLATFORM" == "kubernetes" ]] && ! is_minienv; then
  check_env_var "DOCKER_REGISTRY_URL"
fi

# TODO: consider getting rid of USE_DOCKER_LOCAL_REGISTRY in favour of always using
#  DOCKER_REGISTRY_PATH which when empty would default to DOCKER_REGISTRY_URL.
if ! (( [[ "$PLATFORM" == "kubernetes" ]] && is_minienv ) \
    || [[ "$USE_DOCKER_LOCAL_REGISTRY" == "true" ]]); then
  check_env_var "DOCKER_REGISTRY_PATH"
fi

check_env_var "CONJUR_NAMESPACE"
check_env_var "TEST_APP_NAMESPACE_NAME"
check_env_var "CONJUR_ACCOUNT"
check_env_var "CONJUR_ADMIN_PASSWORD"
check_env_var "AUTHENTICATOR_ID"
check_env_var "TEST_APP_DATABASE"
check_env_var "CONJUR_AUTHN_LOGIN_RESOURCE"
check_env_var "PULL_DOCKER_REGISTRY_URL"
check_env_var "PULL_DOCKER_REGISTRY_PATH"

export CONJUR_APPLIANCE_URL="${CONJUR_APPLIANCE_URL:-https://conjur-oss.$CONJUR_NAMESPACE.svc.cluster.local}"

# For annotation-based Kubernetes authentication, the host ID to be used
# for authenticating is an application name.
export CONJUR_AUTHN_LOGIN_PREFIX="host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps"

# Create the random database password
export SAMPLE_APP_BACKEND_DB_PASSWORD=$(openssl rand -hex 12)

ensure_env_database
