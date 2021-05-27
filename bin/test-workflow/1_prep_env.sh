#!/bin/bash

set -eo pipefail

export DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL:-localhost:5000}"
export DOCKER_REGISTRY_PATH="${DOCKER_REGISTRY_PATH:-localhost:5000}"
export PULL_DOCKER_REGISTRY_URL="${PULL_DOCKER_REGISTRY_URL:-localhost:5000}"
export PULL_DOCKER_REGISTRY_PATH="${PULL_DOCKER_REGISTRY_PATH:-localhost:5000}"
export CONJUR_NAMESPACE="${CONJUR_NAMESPACE:-conjur-oss}"
export TEST_APP_NAMESPACE_NAME="${TEST_APP_NAMESPACE_NAME:-app-test}"
export CONJUR_ACCOUNT="${CONJUR_ACCOUNT:-myConjurAccount}"
export AUTHENTICATOR_ID="${AUTHENTICATOR_ID:-my-authenticator-id}"
export TEST_APP_DATABASE="${TEST_APP_DATABASE:-postgres}"
export CONJUR_AUTHN_LOGIN_RESOURCE="${CONJUR_AUTHN_LOGIN_RESOURCE:-service_account}"
export CONJUR_APPLIANCE_URL="${CONJUR_APPLIANCE_URL:-https://conjur-oss.$CONJUR_NAMESPACE.svc.cluster.local}"
export CONJUR_AUTHN_LOGIN_PREFIX="${CONJUR_AUTHN_LOGIN_PREFIX:-host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps}"
export CONJUR_VERSION="${CONJUR_VERSION:-5}"
export PLATFORM="${PLATFORM:-kubernetes}"  # default to kubernetes if env var not set
export CONJUR_OSS_HELM_INSTALLED="${CONJUR_OSS_HELM_INSTALLED:-true}"
export USE_DOCKER_LOCAL_REGISTRY="${USE_DOCKER_LOCAL_REGISTRY:-false}"

export CONJUR_ADMIN_PASSWORD="$(kubectl exec \
            --namespace "$CONJUR_NAMESPACE" \
            deploy/conjur-oss \
            --container conjur-oss \
            -- conjurctl role retrieve-key "$CONJUR_ACCOUNT":user:admin | tail -1)"

# Create the random database password
export SAMPLE_APP_BACKEND_DB_PASSWORD=$(openssl rand -hex 12)
