#!/bin/bash

set -o pipefail

### PLATFORM DETAILS
export CONJUR_OSS_HELM_INSTALLED="${CONJUR_OSS_HELM_INSTALLED:-true}"
export UNIQUE_TEST_ID="$(uuidgen | tr "[:upper:]" "[:lower:]" | head -c 10)"

# PLATFORM is used to differentiate between general Kubernetes platforms (kubernetes, openshift), while
# CONJUR_PLATFORM is used to differentiate between sub-platforms (kind, gke, jenkins, openshift) for the Conjur deployment
# APP_PLATFORM serves the same purpose as CONJUR_PLATFORM, but for the test app deployment (kind, gke, openshift)
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  CONJUR_PLATFORM="${CONJUR_PLATFORM:-kind}"
else
  CONJUR_PLATFORM="${CONJUR_PLATFORM:-gke}"
fi
export CONJUR_PLATFORM

if [[ "$CONJUR_PLATFORM" == "openshift" ]]; then
  PLATFORM="${PLATFORM:-openshift}"
else
  PLATFORM="${PLATFORM:-kubernetes}"
fi
export PLATFORM

if [[ "$CONJUR_PLATFORM" == "kind" ]]; then
  RUN_CLIENT_CONTAINER="false"
else
  RUN_CLIENT_CONTAINER="true"
fi

if [[ "$CONJUR_PLATFORM" != "kind" ]]; then
  if [[ "$CONJUR_PLATFORM" != "jenkins" ]]; then
    APP_PLATFORM="$CONJUR_PLATFORM"
  elif [[ "$PLATFORM" == "kubernetes" ]]; then
    APP_PLATFORM="gke"
  elif [[ "$PLATFORM" == "openshift" ]]; then
    APP_PLATFORM="openshift"
  fi
fi
export APP_PLATFORM

### DOCKER CONFIG
export USE_DOCKER_LOCAL_REGISTRY="${USE_DOCKER_LOCAL_REGISTRY:-true}"
export DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL:-localhost:5000}"
export DOCKER_REGISTRY_PATH="${DOCKER_REGISTRY_PATH:-localhost:5000}"
export PULL_DOCKER_REGISTRY_URL="${PULL_DOCKER_REGISTRY_URL:-${DOCKER_REGISTRY_URL}}"
export PULL_DOCKER_REGISTRY_PATH="${PULL_DOCKER_REGISTRY_PATH:-${DOCKER_REGISTRY_PATH}}"
export PLATFORM_CONTAINER="platform-container"

### CONJUR AND TEST APP CONFIG
export CONJUR_ACCOUNT="${CONJUR_ACCOUNT:-myConjurAccount}"
export AUTHENTICATOR_ID="${AUTHENTICATOR_ID:-my-authenticator-id}"
export CONJUR_AUTHN_LOGIN_RESOURCE="${CONJUR_AUTHN_LOGIN_RESOURCE:-service_account}"
export CONJUR_AUTHN_LOGIN_PREFIX="${CONJUR_AUTHN_LOGIN_PREFIX:-host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps}"
export CONJUR_VERSION="${CONJUR_VERSION:-5}"
export TEST_APP_DATABASE="${TEST_APP_DATABASE:-postgres}"
export TEST_APP_REPO="${TEST_APP_REPO:-cyberark/demo-app}"
export TEST_APP_TAG="${TEST_APP_TAG:-latest}"
export INSTALL_APPS="${INSTALL_APPS:-summon-sidecar,secretless-broker,secrets-provider-init}"

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  conjur_service="conjur-oss"
  if [[ "$PLATFORM" == "openshift" ]]; then
    export CONJUR_NAMESPACE_NAME="${CONJUR_NAMESPACE_NAME:-$conjur_service-${UNIQUE_TEST_ID}}"
    export HELM_RELEASE="${HELM_RELEASE:-conjur-oss-${UNIQUE_TEST_ID}}"
    export TEST_APP_NAMESPACE_NAME="${TEST_APP_NAMESPACE_NAME:-app-test-$UNIQUE_TEST_ID}"
  else
    export CONJUR_NAMESPACE_NAME="${CONJUR_NAMESPACE_NAME:-$conjur_service}"
    export TEST_APP_NAMESPACE_NAME="${TEST_APP_NAMESPACE_NAME:-app-test}"
  fi
else
  export TEST_APP_NAMESPACE_NAME="${TEST_APP_NAMESPACE_NAME:-app-test-$UNIQUE_TEST_ID}"
  export CONJUR_APPLIANCE_IMAGE="${CONJUR_APPLIANCE_IMAGE:-registry2.itci.conjur.net/conjur-appliance:5.0-stable}"
  export CONJUR_ADMIN_PASSWORD="MySecretP@ss1"

  if [[ "$CONJUR_PLATFORM" == "gke" ]]; then
    conjur_service="conjur-master"
  else
    conjur_service="conjur-authentication"
  fi
  export CONJUR_NAMESPACE_NAME="${CONJUR_NAMESPACE_NAME:-$conjur_service-${UNIQUE_TEST_ID}}"
fi

export CONJUR_APPLIANCE_URL=${CONJUR_APPLIANCE_URL:-https://$conjur_service.$CONJUR_NAMESPACE_NAME.svc.cluster.local}
export SAMPLE_APP_BACKEND_DB_PASSWORD="$(openssl rand -hex 12)"

### PLATFORM SPECIFIC CONFIG
if [[ "$CONJUR_PLATFORM" == "gke" ]]; then
  export CONJUR_FOLLOWER_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local"
  export CONJUR_FOLLOWER_COUNT=1
  export CONJUR_AUTHN_LOGIN="host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/$CONJUR_NAMESPACE_NAME/service_account/conjur-cluster"
  export STOP_RUNNING_ENV=true
  export DEPLOY_MASTER_CLUSTER=true
  export CONFIGURE_CONJUR_MASTER=true
elif [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
  export HOST_IP="${HOST_IP:-$(curl http://169.254.169.254/latest/meta-data/public-ipv4)}"
  export CONJUR_MASTER_PORT="${CONJUR_MASTER_PORT:-40001}"
  export CONJUR_FOLLOWER_PORT="${CONJUR_FOLLOWER_PORT:-40002}"
  export CONJUR_APPLIANCE_URL="https://${HOST_IP}:${CONJUR_MASTER_PORT}"
  export CONJUR_FOLLOWER_URL="https://${HOST_IP}:${CONJUR_FOLLOWER_PORT}"
  export CONJUR_ACCOUNT="demo"

  docker build --tag "custom-certs" \
    --file Dockerfile.jq \
    .
fi

if [[ "$RUN_CLIENT_CONTAINER" == "true" ]]; then
  docker build --tag "$PLATFORM_CONTAINER:$CONJUR_NAMESPACE_NAME" \
      --file Dockerfile \
      --build-arg KUBECTL_VERSION="$KUBECTL_VERSION" \
      --build-arg OPENSHIFT_CLI_URL="$OPENSHIFT_CLI_URL" \
      .
fi
