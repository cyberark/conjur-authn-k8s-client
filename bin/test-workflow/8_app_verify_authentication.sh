#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME

rm -rf bash-lib
git clone https://github.com/cyberark/bash-lib.git

init_bash_lib

RETRIES=150
# Seconds
RETRY_WAIT=2

function finish {
  exit_code=$?

  readonly PIDS=(
    "SIDECAR_PORT_FORWARD_PID"
    "SECRETLESS_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_STANDALONE_PID"
    "SECRETS_PROVIDER_INIT_PORT_FORWARD_PID"
  )

  # Upon error, dump some kubernetes resources and Conjur authentication policy
  if [ $exit_code -ne 0 ]; then
    dump_kubernetes_resources
    dump_authentication_policy
  fi

  set +u

  echo -e "\n\nStopping all port-forwarding"
  for pid in "${PIDS[@]}"; do
    if [ -n "${!pid}" ]; then
      # Kill process, and swallow any errors
      kill "${!pid}" > /dev/null 2>&1
    fi
  done

  if [ $exit_code -eq 0 ]; then
    announce "Test PASSED!!!!"
  else
    announce "Test FAILED!!!!"
  fi
}
trap finish EXIT

announce "Validating that the deployments are functioning as expected."

set_namespace "$TEST_APP_NAMESPACE_NAME"

deploy_test_curl() {
  $cli delete --ignore-not-found pod/test-curl
  $cli create -f ./"$PLATFORM"/test-curl.yml
}

check_test_curl() {
  pods_ready "test-curl"
}

pod_curl() {
  kubectl exec test-curl -- curl "$@"
}

echo "Deploying a test curl pod"
deploy_test_curl
echo "Waiting for test curl pod to become available"
bl_retry_constant "${RETRIES}" "${RETRY_WAIT}"  check_test_curl
  
echo "Waiting for pods to become available"

if [[ "$PLATFORM" == "openshift" ]]; then
  sidecar_pod=$(get_pod_name test-app-summon-sidecar)
  secretless_pod=$(get_pod_name test-app-secretless)
  secrets_provider_standalone_pod=$(get_pod_name test-app-secrets-provider-standalone)
  secrets_provider_init_pod=$(get_pod_name test-app-secrets-provider-init)

  # Routes are defined, but we need to do port-mapping to access them
  oc port-forward "$sidecar_pod" 8081:8080 > /dev/null 2>&1 &
  SIDECAR_PORT_FORWARD_PID=$!
  oc port-forward "$secretless_pod" 8083:8080 > /dev/null 2>&1 &
  SECRETLESS_PORT_FORWARD_PID=$!
  oc port-forward "$secrets_provider_standalone_pod" 8084:8080 > /dev/null 2>&1 &
  SECRETS_PROVIDER_STANDALONE_PID=$!
  oc port-forward "$secrets_provider_init_pod" 8086:8080 > /dev/null 2>&1 &
  SECRETS_PROVIDER_INIT_PORT_FORWARD_PID=$!

  curl_cmd=curl
  sidecar_url="localhost:8081"
  secretless_url="localhost:8083"
  secrets_provider_standalone_url="localhost:8084"
  secrets_provider_init_url="localhost:8086"
else
  # Test by curling from a pod that is inside the KinD cluster.
  curl_cmd=pod_curl
  sidecar_url="test-app-summon-sidecar.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secretless_url="test-app-secretless-broker.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_standalone_url="test-app-secrets-provider-standalone.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_init_url="test-app-secrets-provider-init.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
fi

echo "Waiting for urls to be ready"

check_url(){
  ( $curl_cmd -sS --connect-timeout 3 "$1" ) > /dev/null
}

# restore array of apps to run
declare -a install_apps=($(split_on_comma_delimiter $INSTALL_APPS))

# declare associative arrays of app urls and pet names
declare -A app_urls
app_urls[summon-sidecar]="$sidecar_url"
app_urls[secretless-broker]="$secretless_url"
app_urls[secrets-provider-standalone]="$secrets_provider_standalone_url"
app_urls[secrets-provider-init]="$secrets_provider_init_url"

declare -A app_pets
app_pets[summon-sidecar]="Mr. Sidecar"
app_pets[secretless-broker]="Mr. Secretless"
app_pets[secrets-provider-standalone]="Mr. Standalone"
app_pets[secrets-provider-init]="Mr. Provider"

# check connection to each installed test app
for app in "${install_apps[@]}"; do
  bl_retry_constant "${RETRIES}" "${RETRY_WAIT}" check_url "${app_urls[$app]}"
done

# add pet to and query pets from each installed test app
for app in "${install_apps[@]}"; do
  echo -e "\nAdding entry with $app app\n"
  $curl_cmd \
    -d "{\"name\": \"${app_pets[$app]}\"}" \
    -H "Content-Type: application/json" \
    "${app_urls[$app]}"/pet

  echo -e "\n\nQuerying $app app\n"
  $curl_cmd "${app_urls[$app]}"/pets
done
