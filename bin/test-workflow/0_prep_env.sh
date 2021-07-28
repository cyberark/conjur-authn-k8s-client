#!/bin/bash

set -o pipefail

### PLATFORM DETAILS
export CONJUR_OSS_HELM_INSTALLED="${CONJUR_OSS_HELM_INSTALLED:-true}"
export UNIQUE_TEST_ID="$(uuidgen | tr "[:upper:]" "[:lower:]" | head -c 10)"

# PLATFORM is used to differentiate between general Kubernetes platforms (K8s vs. oc), while
# CLUSTER_TYPE is used to differentiate between sub-platforms (for vanilla K8s, KinD vs. GKE)
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  CLUSTER_TYPE="${CLUSTER_TYPE:-kind}"
else
  CLUSTER_TYPE="${CLUSTER_TYPE:-gke}"
fi
export CLUSTER_TYPE

if [[ "$CLUSTER_TYPE" == "oc" ]]; then
  PLATFORM="openshift"
else
  PLATFORM="kubernetes"
fi
export PLATFORM

### DOCKER CONFIG
export USE_DOCKER_LOCAL_REGISTRY="${USE_DOCKER_LOCAL_REGISTRY:-true}"
export DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL:-localhost:5000}"
export DOCKER_REGISTRY_PATH="${DOCKER_REGISTRY_PATH:-localhost:5000}"
export PULL_DOCKER_REGISTRY_URL="${PULL_DOCKER_REGISTRY_URL:-${DOCKER_REGISTRY_URL}}"
export PULL_DOCKER_REGISTRY_PATH="${PULL_DOCKER_REGISTRY_PATH:-${DOCKER_REGISTRY_PATH}}"

### CONJUR AND TEST APP CONFIG
export CONJUR_ACCOUNT="${CONJUR_ACCOUNT:-myConjurAccount}"
export AUTHENTICATOR_ID="${AUTHENTICATOR_ID:-my-authenticator-id}"
export CONJUR_AUTHN_LOGIN_RESOURCE="${CONJUR_AUTHN_LOGIN_RESOURCE:-service_account}"
export CONJUR_AUTHN_LOGIN_PREFIX="${CONJUR_AUTHN_LOGIN_PREFIX:-host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps}"
export CONJUR_VERSION="${CONJUR_VERSION:-5}"
export TEST_APP_NAMESPACE_NAME="${TEST_APP_NAMESPACE_NAME:-app-test}"
export TEST_APP_DATABASE="${TEST_APP_DATABASE:-postgres}"

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
    conjur_service="conjur-oss"
    export CONJUR_NAMESPACE_NAME="${CONJUR_NAMESPACE_NAME:-$conjur_service}"
else
    conjur_service="conjur-master"
    export CONJUR_NAMESPACE_NAME="${CONJUR_NAMESPACE_NAME:-$conjur_service-${UNIQUE_TEST_ID}}"
    export TEST_APP_NAMESPACE_NAME="$TEST_APP_NAMESPACE_NAME-$UNIQUE_TEST_ID"
fi

export CONJUR_APPLIANCE_URL=${CONJUR_APPLIANCE_URL:-https://$conjur_service.$CONJUR_NAMESPACE_NAME.svc.cluster.local}
export SAMPLE_APP_BACKEND_DB_PASSWORD="$(openssl rand -hex 12)"

### PLATFORM SPECIFIC CONFIG
if [[ "$CLUSTER_TYPE" == "gke" ]]; then
    export CONJUR_FOLLOWER_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local"
    export CONJUR_ADMIN_PASSWORD="MySecretP@ss1"
    export CONJUR_APPLIANCE_IMAGE="registry2.itci.conjur.net/conjur-appliance:5.0-stable"
    export CONJUR_FOLLOWER_COUNT=1
    export CONJUR_AUTHN_LOGIN="host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/$CONJUR_NAMESPACE_NAME/service_account/conjur-cluster"
    export STOP_RUNNING_ENV=true
    export DEPLOY_MASTER_CLUSTER=true
    export CONFIGURE_CONJUR_MASTER=true
    export PLATFORM_CONTAINER="platform-container"

    docker build --tag "$PLATFORM_CONTAINER:$CONJUR_NAMESPACE_NAME" \
        --file Dockerfile \
        --build-arg KUBECTL_VERSION="$KUBECTL_VERSION" \
        .
fi
