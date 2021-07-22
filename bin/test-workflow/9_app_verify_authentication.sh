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

# Dump some kubernetes resources and Conjur authentication policy if this
# script exits prematurely
DETAILED_DUMP_ON_EXIT=true

function finish {
  readonly PIDS=(
    "SIDECAR_PORT_FORWARD_PID"
    "INIT_PORT_FORWARD_PID"
    "INIT_WITH_HOST_OUTSIDE_APPS_PORT_FORWARD_PID"
    "SECRETLESS_PORT_FORWARD_PID"
  )

  if [[ "$DETAILED_DUMP_ON_EXIT" == "true" ]]; then
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
  init_pod=$(get_pod_name test-app-summon-init)
  init_pod_with_host_outside_apps=$(get_pod_name test-app-with-host-outside-apps-branch-summon-init)
  secretless_pod=$(get_pod_name test-app-secretless)

  # Routes are defined, but we need to do port-mapping to access them
  oc port-forward "$sidecar_pod" 8081:8080 > /dev/null 2>&1 &
  SIDECAR_PORT_FORWARD_PID=$!
  oc port-forward "$init_pod" 8082:8080 > /dev/null 2>&1 &
  INIT_PORT_FORWARD_PID=$!
  oc port-forward "$secretless_pod" 8083:8080 > /dev/null 2>&1 &
  SECRETLESS_PORT_FORWARD_PID=$!
  oc port-forward "$init_pod_with_host_outside_apps" 8084:8080 > /dev/null 2>&1 &
  INIT_WITH_HOST_OUTSIDE_APPS_PORT_FORWARD_PID=$!

  curl_cmd=curl
  sidecar_url="localhost:8081"
  init_url="localhost:8082"
  secretless_url="localhost:8083"
  init_url_with_host_outside_apps="localhost:8084"
else
  # Test by curling from a pod that is inside the KinD cluster.
  curl_cmd=pod_curl
  init_url="test-app-summon-init.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  init_url_with_host_outside_apps="test-app-with-host-outside-apps-branch-summon-init.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  sidecar_url="test-app-summon-sidecar.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secretless_url="test-app-secretless.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
fi

echo "Waiting for urls to be ready"

check_urls(){
  (
    # $curl_cmd -sS --connect-timeout 3 "$init_url" &&
    # $curl_cmd -sS --connect-timeout 3 "$init_url_with_host_outside_apps" &&
    $curl_cmd -sS --connect-timeout 3 "$sidecar_url" # &&
    # $curl_cmd -sS --connect-timeout 3 "$secretless_url"
  ) > /dev/null
}

bl_retry_constant "${RETRIES}" "${RETRY_WAIT}" check_urls

# echo -e "\nAdding entry to the init app\n"
# $curl_cmd \
#   -d '{"name": "Mr. Init"}' \
#   -H "Content-Type: application/json" \
#   "$init_url"/pet

# echo -e "Adding entry to the init app with host outside apps\n"
# $curl_cmd \
#   -d '{"name": "Mr. Init"}' \
#   -H "Content-Type: application/json" \
#   "$init_url_with_host_outside_apps"/pet

echo -e "Adding entry to the sidecar app\n"
$curl_cmd \
  -d '{"name": "Mr. Sidecar"}' \
  -H "Content-Type: application/json" \
  "$sidecar_url"/pet

# echo -e "Adding entry to the secretless app\n"
# $curl_cmd \
#   -d '{"name": "Mr. Secretless"}' \
#   -H "Content-Type: application/json" \
#   "$secretless_url"/pet

# echo -e "Querying init app\n"
# $curl_cmd "$init_url"/pets

# echo -e "\n\nQuerying init app with hosts outside apps\n"
# $curl_cmd "$init_url_with_host_outside_apps"/pets

echo -e "\n\nQuerying sidecar app\n"
$curl_cmd "$sidecar_url"/pets

# echo -e "\n\nQuerying secretless app\n"
# $curl_cmd "$secretless_url"/pets

DETAILED_DUMP_ON_EXIT=false
