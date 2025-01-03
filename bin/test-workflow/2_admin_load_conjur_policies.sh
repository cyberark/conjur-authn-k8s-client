#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"

source utils.sh

# Upon error, dump kubernetes resources in the Conjur Namespace
trap dump_conjur_namespace_upon_error EXIT

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  export CONJUR_ADMIN_PASSWORD="$(get_admin_password)"
fi

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var TEST_APP_NAMESPACE_LABEL
check_env_var CONJUR_ACCOUNT
check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_ADMIN_PASSWORD
check_env_var AUTHENTICATOR_ID
check_env_var CONJUR_NAMESPACE_NAME
check_env_var TEST_APP_DATABASE
check_env_var SAMPLE_APP_BACKEND_DB_PASSWORD
TEST_JWT_FLOW="${TEST_JWT_FLOW:-false}"

announce "Generating Conjur policy."

prepare_conjur_cli_image() {
  announce "Pulling and pushing Conjur CLI image."

  docker pull cyberark/conjur-cli:8

  cli_app_image="$(platform_image_for_push conjur-cli $CONJUR_NAMESPACE_NAME)"
  docker tag cyberark/conjur-cli:8 "$cli_app_image"

  docker push "$cli_app_image"
}

deploy_conjur_cli() {
  announce "Deploying Conjur CLI pod."

  IMAGE_PULL_POLICY='Always'

  cli_app_image="$(platform_image_for_pull conjur-cli $CONJUR_NAMESPACE_NAME)"
  sed -e "s#{{ CONJUR_SERVICE_ACCOUNT }}#$(conjur_service_account)#g" ./$PLATFORM/conjur-cli.yml |
    sed -e "s#{{ DOCKER_IMAGE }}#$cli_app_image#g" |
    sed -e "s#{{ IMAGE_PULL_POLICY }}#$IMAGE_PULL_POLICY#g" |
    "$cli" create -f -

  # Wait until pod appears otherwise $conjur_cli_pod could be empty and we would wait forever
  wait_for_it 300 "has_resource 'app=conjur-cli'"
  conjur_cli_pod=$(get_conjur_cli_pod_name)
  wait_for_it 300 "$cli get pod $conjur_cli_pod -o jsonpath='{.status.phase}'| grep -q Running"
}

ensure_conjur_cli_initialized() {
  announce "Ensure that Conjur CLI pod has a connection with Conjur initialized."

  "$cli" exec "$1" -- sh -c "echo y | conjur init -a '$CONJUR_ACCOUNT' -u '$CONJUR_APPLIANCE_URL' --self-signed --force"
  # Flaky with 500 Internal Server Error, mitigate with retry
  wait_for_it 300 "$cli exec $1 -- conjur login -i admin -p '$CONJUR_ADMIN_PASSWORD'"
}

pushd policy > /dev/null
  mkdir -p ./generated > /dev/null
  chmod 777 ./generated || true

  # NOTE: generated files are prefixed with the test app namespace to allow for parallel CI

  if [[ "$PLATFORM" == "openshift" ]]; then
    is_openshift=true
    is_kubernetes=false
    if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
      oc adm policy add-scc-to-user anyuid system:serviceaccount:$CONJUR_NAMESPACE_NAME:conjur-oss
    fi
  else
    is_openshift=false
    is_kubernetes=true
  fi

  validator_id="${VALIDATOR_ID:-validator}"
  validator_namespace="${VALIDATOR_NAMESPACE_NAME:-$CONJUR_NAMESPACE_NAME}"
  app_validator_id="${APP_VALIDATOR_ID:-app-validator}"
  app_validator_namespace="${APP_VALIDATOR_NAMESPACE_NAME:-$TEST_APP_NAMESPACE_NAME}"

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/authenticator-policy.template.yml > ./generated/"$TEST_APP_NAMESPACE_NAME".authenticator-policy.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/app-identities-policy.template.yml |
    sed "s#{{ VALIDATOR_ID }}#$validator_id#g" |
    sed "s#{{ VALIDATOR_NAMESPACE_NAME }}#$validator_namespace#g" |
    sed "s#{{ APP_VALIDATOR_ID }}#$app_validator_id#g" |
    sed "s#{{ APP_VALIDATOR_NAMESPACE_NAME }}#$app_validator_namespace#g" |
    sed "s#{{ IS_OPENSHIFT }}#$is_openshift#g" |
    sed "s#{{ IS_KUBERNETES }}#$is_kubernetes#g" |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g" |
    sed "s#{{ TEST_APP_NAMESPACE_LABEL }}#$TEST_APP_NAMESPACE_LABEL#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".app-identities-policy.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g" ./templates/app-identities-policy-jwt.template.yml |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g"  > ./generated/"$TEST_APP_NAMESPACE_NAME".app-identities-policy-jwt.yml

  sed "s#{{ AUTHENTICATOR_ID }}#$AUTHENTICATOR_ID#g"  ./templates/app-grants.template.yml |
    sed "s#{{ TEST_APP_NAMESPACE_NAME }}#$TEST_APP_NAMESPACE_NAME#g" > ./generated/"$TEST_APP_NAMESPACE_NAME".app-grants.yml
popd > /dev/null

if [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
  JWKS_URI="NONE"
  ISSUER="NONE"
  docker compose -f "temp/conjur-intro-$UNIQUE_TEST_ID/docker-compose.yml" \
    run --rm \
    -v "${PWD}/policy":/tmp/policy \
    -w /src/cli \
    --entrypoint /bin/sh \
    client -c "
      conjur_appliance_url='${CONJUR_APPLIANCE_URL}' \
      CONJUR_ACCOUNT='${CONJUR_ACCOUNT}' \
      CONJUR_ADMIN_PASSWORD='${CONJUR_ADMIN_PASSWORD}' \
      DB_PASSWORD='${SAMPLE_APP_BACKEND_DB_PASSWORD}' \
      TEST_APP_NAMESPACE_NAME='${TEST_APP_NAMESPACE_NAME}' \
      TEST_APP_DATABASE='${TEST_APP_DATABASE}' \
      AUTHENTICATOR_ID='${AUTHENTICATOR_ID}' \
      JWKS_URI='${JWKS_URI}'\
      ISSUER='${ISSUER}'\
      /tmp/policy/load_policies.sh
    "
else
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

  # Avoid using kubectl cp because it requires the `tar` command to be
  # installed on the source and destination pods. Instead, use `kubectl exec`
  # to write the policy file to the destination pod.

  "$cli" exec "$conjur_cli_pod" -- rm -rf /tmp/policy
  "$cli" exec "$conjur_cli_pod" -- mkdir -p /tmp/policy/generated

  set -- "generated/$TEST_APP_NAMESPACE_NAME.authenticator-policy.yml" \
    "generated/$TEST_APP_NAMESPACE_NAME.app-identities-policy.yml" \
    "generated/$TEST_APP_NAMESPACE_NAME.app-identities-policy-jwt.yml" \
    "app-policy.yml" \
    "generated/$TEST_APP_NAMESPACE_NAME.app-grants.yml" \
    "load_policies.sh"

  for policy_file in "$@"; do
    "$cli" exec -i "$conjur_cli_pod" -- sh -c "cat - > /tmp/policy/$policy_file" < "${PWD}/policy/$policy_file"
  done

  "$cli" exec "$conjur_cli_pod" -- chmod +x /tmp/policy/load_policies.sh

  announce "Extracting openid configuration"
  JWKS_URI=$($cli get --raw /.well-known/openid-configuration | jq '.jwks_uri')
  ISSUER=$($cli get --raw /.well-known/openid-configuration | jq '.issuer')
  announce "JWKS URI of this cluster is $JWKS_URI and Issuer is $ISSUER"

  wait_for_it 300 "$cli exec $conjur_cli_pod -- \
    sh -c \"
      conjur_appliance_url='${CONJUR_APPLIANCE_URL}' \
      CONJUR_ACCOUNT='${CONJUR_ACCOUNT}' \
      CONJUR_ADMIN_PASSWORD='${CONJUR_ADMIN_PASSWORD}' \
      DB_PASSWORD='${SAMPLE_APP_BACKEND_DB_PASSWORD}' \
      TEST_APP_NAMESPACE_NAME='${TEST_APP_NAMESPACE_NAME}' \
      TEST_APP_DATABASE='${TEST_APP_DATABASE}' \
      AUTHENTICATOR_ID='${AUTHENTICATOR_ID}' \
      JWKS_URI='${JWKS_URI}'\
      ISSUER='${ISSUER}'\
      /tmp/policy/load_policies.sh
    \"
  "

  "$cli" exec "$conjur_cli_pod" -- rm -rf ./tmp/policy
fi

echo "Conjur policy loaded."
