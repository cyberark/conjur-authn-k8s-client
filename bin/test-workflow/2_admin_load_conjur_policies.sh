#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"

source utils.sh

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  export CONJUR_ADMIN_PASSWORD="$(get_admin_password)"
fi

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_VERSION
check_env_var CONJUR_ACCOUNT
check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_ADMIN_PASSWORD
check_env_var AUTHENTICATOR_ID
check_env_var CONJUR_NAMESPACE_NAME
check_env_var TEST_APP_DATABASE
check_env_var SAMPLE_APP_BACKEND_DB_PASSWORD

announce "Generating Conjur policy."

prepare_conjur_cli_image() {
  announce "Pulling and pushing Conjur CLI image."

  docker pull cyberark/conjur-cli:"$CONJUR_VERSION"-latest

  cli_app_image="$(platform_image_for_push conjur-cli)"
  docker tag cyberark/conjur-cli:"$CONJUR_VERSION"-latest "$cli_app_image"

  docker push "$cli_app_image"
}

deploy_conjur_cli() {
  announce "Deploying Conjur CLI pod."

  IMAGE_PULL_POLICY='Always'

  cli_app_image=$(platform_image_for_pull conjur-cli)
  sed -e "s#{{ CONJUR_SERVICE_ACCOUNT }}#$(conjur_service_account)#g" ./$PLATFORM/conjur-cli.yml |
    sed -e "s#{{ DOCKER_IMAGE }}#$cli_app_image#g" |
    sed -e "s#{{ IMAGE_PULL_POLICY }}#$IMAGE_PULL_POLICY#g" |
    $cli create -f -

  # Wait until pod appears otherwise $conjur_cli_pod could be empty and we would wait forever
  wait_for_it 300 "has_resource 'app=conjur-cli'"
  conjur_cli_pod=$(get_conjur_cli_pod_name)
  wait_for_it 300 "$cli get pod $conjur_cli_pod -o jsonpath='{.status.phase}'| grep -q Running"
}

ensure_conjur_cli_initialized() {
  announce "Ensure that Conjur CLI pod has a connection with Conjur initialized."

  $cli exec $1 -- bash -c "yes yes | conjur init -a $CONJUR_ACCOUNT -u $CONJUR_APPLIANCE_URL"
  # Flaky with 500 Internal Server Error, mitigate with retry
  wait_for_it 300 "$cli exec $1 -- conjur authn login -u admin -p $CONJUR_ADMIN_PASSWORD"
}

pushd policy > /dev/null
  mkdir -p ./generated > /dev/null

  # NOTE: generated files are prefixed with the test app namespace to allow for parallel CI

  if [[ "$PLATFORM" == "openshift" ]]; then
    is_openshift=true
    is_kubernetes=false
  else
    is_openshift=false
    is_kubernetes=true
  fi

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/cluster-authn-svc-def.template.yml |
    sed "s#{{ CONJUR_NAMESPACE }}#$CONJUR_NAMESPACE_NAME#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".cluster-authn-svc.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/project-authn-def.template.yml |
    sed "s#{{ IS_OPENSHIFT }}#$is_openshift#g" |
    sed "s#{{ IS_KUBERNETES }}#$is_kubernetes#g" |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".project-authn.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/app-identity-def.template.yml |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".app-identity.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/authn-any-policy-branch.template.yml |
    sed "s#{{ IS_OPENSHIFT }}#$is_openshift#g" |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".authn-any-policy-branch.yml
popd > /dev/null

set_namespace "$CONJUR_NAMESPACE_NAME"

announce "Finding or creating a Conjur CLI pod"
conjur_cli_pod="$(get_conjur_cli_pod_name)"
if [ -z "$conjur_cli_pod" ]; then
  prepare_conjur_cli_image
  deploy_conjur_cli
  conjur_cli_pod="$(get_conjur_cli_pod_name)"
fi

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  ensure_conjur_cli_initialized "$conjur_cli_pod"
fi

announce "Loading Conjur policy."

$cli exec "$conjur_cli_pod" -- rm -rf /policy
$cli cp ./policy "$conjur_cli_pod:/policy"

wait_for_it 300 "$cli exec $conjur_cli_pod -- \
  bash -c \"
  conjur_appliance_url=${CONJUR_APPLIANCE_URL}
    CONJUR_ACCOUNT=${CONJUR_ACCOUNT} \
    CONJUR_ADMIN_PASSWORD=${CONJUR_ADMIN_PASSWORD} \
    DB_PASSWORD=${SAMPLE_APP_BACKEND_DB_PASSWORD} \
    TEST_APP_NAMESPACE_NAME=${TEST_APP_NAMESPACE_NAME} \
    TEST_APP_DATABASE=${TEST_APP_DATABASE} \
    /policy/load_policies.sh
  \"
"

$cli exec "$conjur_cli_pod" -- rm -rf ./policy

echo "Conjur policy loaded."
