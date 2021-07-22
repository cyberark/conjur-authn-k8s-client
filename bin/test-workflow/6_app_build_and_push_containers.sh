#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"
USE_DOCKER_LOCAL_REGISTRY="${USE_DOCKER_LOCAL_REGISTRY:-true}"
DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL:-localhost:5000}"
PULL_DOCKER_REGISTRY_URL="${PULL_DOCKER_REGISTRY_URL:-localhost:5000}"
CONJUR_OSS_HELM_INSTALLED="${CONJUR_OSS_HELM_INSTALLED:-true}"

source utils.sh

if [[ "$PLATFORM" == "openshift" ]]; then
    docker login -u _ -p $(oc whoami -t) "$DOCKER_REGISTRY_PATH"
fi

announce "Building and pushing test app images."

readonly APPS=(
  "init"
  "sidecar"
)

pushd test_app_summon
  if [[ "$PLATFORM" == "openshift" ]]; then
    echo "Building Summon binaries to include in app image"
    docker build -t test-app-builder -f Dockerfile.builder .

    # retrieve the summon binaries
    id="$(docker create test-app-builder)"
    docker cp "$id":/usr/local/lib/summon/summon-conjur ./tmp.summon-conjur
    docker cp "$id":/usr/local/bin/summon ./tmp.summon
    docker rm --volumes "$id"
  fi


  for app_type in "${APPS[@]}"; do
    # prep secrets.yml
    # NOTE: generated files are prefixed with the test app namespace to allow for parallel CI
    sed "s#{{ TEST_APP_NAME }}#test-summon-$app_type-app#g" ./secrets.template.yml > "tmp.$TEST_APP_NAMESPACE_NAME.secrets.yml"

    dockerfile="Dockerfile"
    if [[ "$PLATFORM" == "openshift" ]]; then
      dockerfile="Dockerfile.oc"
    fi

    echo "Building test app image"
    docker build \
      --build-arg namespace="$TEST_APP_NAMESPACE_NAME" \
      --tag test-app:"$CONJUR_NAMESPACE_NAME" \
      --file "$dockerfile" .

    test_app_image=$(platform_image_for_push "test-$app_type-app")
    docker tag "test-app:$CONJUR_NAMESPACE_NAME" "$test_app_image"

    docker push "$test_app_image"
  done
popd
