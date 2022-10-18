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
  if [ $? -eq 0 ]; then
    announce "Test PASSED!!!!"
  else
    announce "Test FAILED!!!! Displaying Kubernetes Resources"
    dump_kubernetes_resources "$TEST_APP_NAMESPACE_NAME"
    dump_kubernetes_resources "$CONJUR_NAMESPACE_NAME"
    dump_local_docker_logs
    dump_authentication_policy
  fi

  readonly PIDS=(
    "SIDECAR_PORT_FORWARD_PID"
    "SIDECAR_JWT_PORT_FORWARD_PID"
    "SECRETLESS_PORT_FORWARD_PID"
    "SECRETLESS_JWT_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_STANDALONE_PID"
    "SECRETS_PROVIDER_K8S_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_K8S_JWT_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_P2F_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_P2F_JWT_PORT_FORWARD_PID"
    "SECRETS_PROVIDER_ROTATION_PORT_FORWARD_PID"
  )

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
bl_retry_constant "${RETRIES}" "${RETRY_WAIT}" check_test_curl

echo "Waiting for pods to become available"

# restore array of apps to run
declare -a install_apps=($(split_on_comma_delimiter $INSTALL_APPS))

if [[ "$PLATFORM" == "openshift" ]]; then
  # Routes are defined, but we need to do port-mapping to access them.
  if [[ " ${install_apps[*]} " =~ " summon-sidecar " ]]; then
    sidecar_pod=$(get_pod_name test-app-summon-sidecar)
    oc port-forward "$sidecar_pod" 8081:8080 > /dev/null 2>&1 &
    SIDECAR_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " summon-sidecar-jwt " ]]; then
    sidecar_jwt_pod=$(get_pod_name test-app-summon-sidecar-jwt)
    oc port-forward "$sidecar_jwt_pod" 8081:8080 > /dev/null 2>&1 &
    SIDECAR_JWT_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secretless-broker " ]]; then
    secretless_pod=$(get_pod_name test-app-secretless)
    oc port-forward "$secretless_pod" 8083:8080 > /dev/null 2>&1 &
    SECRETLESS_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secretless-broker-jwt " ]]; then
    secretless_jwt_pod=$(get_pod_name test-app-secretless-jwt)
    oc port-forward "secretless_jwt_pod" 8083:8080 > /dev/null 2>&1 &
    SECRETLESS_JWT_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secrets-provider-standalone " ]]; then
    secrets_provider_standalone_pod=$(get_pod_name test-app-secrets-provider-standalone)
    oc port-forward "$secrets_provider_standalone_pod" 8084:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_STANDALONE_PID=$!
  fi
  
  if [[ " ${install_apps[*]} " =~ " secrets-provider-k8s " ]]; then
    secrets_provider_k8s_pod=$(get_pod_name test-app-secrets-provider-k8s)
    oc port-forward "$secrets_provider_k8s_pod" 8086:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_K8S_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secrets-provider-k8s-jwt " ]]; then
    secrets_provider_k8s_jwt_pod=$(get_pod_name test-app-secrets-provider-k8s-jwt)
    oc port-forward "$secrets_provider_k8s_jwt_pod" 8086:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_K8S_JWT_PORT_FORWARD_PID=$!
  fi
  
  if [[ " ${install_apps[*]} " =~ " secrets-provider-p2f " ]]; then
    secrets_provider_p2f_pod=$(get_pod_name test-app-secrets-provider-p2f)
    oc port-forward "$secrets_provider_p2f_pod" 8087:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_P2F_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secrets-provider-p2f-jwt " ]]; then
    secrets_provider_p2f_jwt_pod=$(get_pod_name test-app-secrets-provider-p2f-jwt)
    oc port-forward "$secrets_provider_p2f_jwt_pod" 8087:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_P2F_JWT_PORT_FORWARD_PID=$!
  fi

  if [[ " ${install_apps[*]} " =~ " secrets-provider-rotation " ]]; then
    secrets_provider_rotation_pod=$(get_pod_name test-app-secrets-provider-rotation)
    oc port-forward "$secrets_provider_rotation_pod" 8088:8080 > /dev/null 2>&1 &
    SECRETS_PROVIDER_ROTATION_PORT_FORWARD_PID=$!
  fi

  curl_cmd=curl
  sidecar_url="localhost:8081"
  secretless_url="localhost:8083"
  secrets_provider_standalone_url="localhost:8084"
  secrets_provider_k8s_url="localhost:8086"
  secrets_provider_p2f_url="localhost:8087"
  secrets_provider_rotation_url="localhost:8088"
else
  # Test by curling from a pod that is inside the KinD cluster.
  curl_cmd=pod_curl
  sidecar_url="test-app-summon-sidecar.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secretless_url="test-app-secretless-broker.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_standalone_url="test-app-secrets-provider-standalone.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_k8s_url="test-app-secrets-provider-k8s.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_p2f_url="test-app-secrets-provider-p2f.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
  secrets_provider_rotation_url="test-app-secrets-provider-rotation.$TEST_APP_NAMESPACE_NAME.svc.cluster.local:8080"
fi

echo "Waiting for urls to be ready"

check_url(){
  ( $curl_cmd -sS --connect-timeout 3 "$1" ) > /dev/null
}

# declare associative arrays of app urls and pet names
declare -A app_urls
app_urls[summon-sidecar]="$sidecar_url"
app_urls[summon-sidecar-jwt]="$sidecar_url"
app_urls[secretless-broker]="$secretless_url"
app_urls[secretless-broker-jwt]="$secretless_url"
app_urls[secrets-provider-standalone]="$secrets_provider_standalone_url"
app_urls[secrets-provider-k8s]="$secrets_provider_k8s_url"
app_urls[secrets-provider-k8s-jwt]="$secrets_provider_k8s_url"
app_urls[secrets-provider-p2f]="$secrets_provider_p2f_url"
app_urls[secrets-provider-p2f-jwt]="$secrets_provider_p2f_url"
app_urls[secrets-provider-rotation]="$secrets_provider_rotation_url"

declare -A app_pets
app_pets[summon-sidecar]="Mr. Sidecar"
app_pets[summon-sidecar-jwt]="Mr. Sidecar JWT"
app_pets[secretless-broker]="Mr. Secretless"
app_pets[secretless-broker-jwt]="Mr. Secretless JWT"
app_pets[secrets-provider-standalone]="Mr. Standalone"
app_pets[secrets-provider-k8s]="Mr. Provider"
app_pets[secrets-provider-k8s-jwt]="Mr. JWT Provider"
app_pets[secrets-provider-p2f]="Mr. FileProvider"
app_pets[secrets-provider-p2f-jwt]="Mr. JWT FileProvider"
app_pets[secrets-provider-rotation]="Mr. Rotation"

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


if [[ " ${install_apps[*]} " =~ " secrets-provider-rotation " ]]; then
  # Test secrets rotation
  echo "Testing secrets provider with rotation"
  # Change secret values in conjur
  "$cli" exec "$(get_conjur_cli_pod_name)" -n "$CONJUR_NAMESPACE_NAME" -- bash -c "\
    conjur authn login -u admin -p $(get_admin_password) && \
    conjur variable values add 'test-secrets-provider-rotation-app-db/counter' '1'" > /dev/null
  # Check if SP picks up the new values
  sleep 10
  pod_name=$(get_pod_name test-app-secrets-provider-rotation)
  content="$($cli exec "$pod_name" -c test-app -- cat /mounted/secrets/dummy.yaml)"
  if [[ "$content" == "counter: 1" ]]; then
    echo "Success: Secrets Provider picked up new values"
  else
    echo "Failure: Secrets Provider did not pick up new values"
    echo "Secret file content:
      $content"
    echo "Secrets Provider logs:"
    "$cli" logs "$pod_name" -c cyberark-secrets-provider-for-k8s
    exit 1
  fi
fi
