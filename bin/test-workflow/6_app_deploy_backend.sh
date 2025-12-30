#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var SAMPLE_APP_BACKEND_DB_PASSWORD

TIMEOUT="${TIMEOUT:-300}"

# Upon error, dump kubernetes resources in the application Namespace
trap dump_application_namespace_upon_error EXIT

announce "Deploying test app postgres backend for $TEST_APP_NAMESPACE_NAME."

set_namespace "$TEST_APP_NAMESPACE_NAME"

# Delete existing backend resources if they exist
echo "Cleaning up any existing backend deployment"
$cli delete --namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found \
  statefulset test-app-backend

# Delete any created PVCs
$cli delete --namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found \
  pvc -l app=test-app-backend

# Delete existing secrets
$cli delete --namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found \
  secret test-app-backend-certs test-app-backend-password

echo "Creating secrets for test app backend"

# Create TLS certificate secret
$cli --namespace "$TEST_APP_NAMESPACE_NAME" \
  create secret generic \
  test-app-backend-certs \
  --from-file=server.crt=./etc/ca.pem \
  --from-file=server.key=./etc/ca-key.pem

# Create password secret
$cli --namespace "$TEST_APP_NAMESPACE_NAME" \
  create secret generic \
  test-app-backend-password \
  --from-literal=password="$SAMPLE_APP_BACKEND_DB_PASSWORD"

# Select the appropriate manifest based on platform
if [[ "$PLATFORM" == "openshift" ]]; then
  manifest_file="./openshift/postgres-backend.yaml"
else
  manifest_file="./kubernetes/postgres-backend.yaml"
fi

echo "Deploying PostgreSQL backend using $manifest_file"
$cli apply --namespace "$TEST_APP_NAMESPACE_NAME" -f "$manifest_file"

# Wait for the StatefulSet to be ready
echo "Waiting for PostgreSQL backend to be ready (timeout: ${TIMEOUT}s)..."

# First wait for the pod to be created (up to 60 seconds)
echo "Waiting for pod test-app-backend-0 to be created..."
for i in {1..60}; do
  if $cli get pod test-app-backend-0 --namespace "$TEST_APP_NAMESPACE_NAME" &>/dev/null; then
    echo "Pod test-app-backend-0 found"
    break
  fi
  if [[ $i -eq 60 ]]; then
    announce "Failed to deploy PostgreSQL backend - pod not created after 60 seconds"
    $cli get statefulset -n "$TEST_APP_NAMESPACE_NAME"
    $cli describe statefulset test-app-backend -n "$TEST_APP_NAMESPACE_NAME"
    exit 1
  fi
  sleep 1
done

# Now wait for the pod to become ready
ret_val=0
$cli wait --namespace "$TEST_APP_NAMESPACE_NAME" \
  --for=condition=ready \
  --timeout="${TIMEOUT}s" \
  pod/test-app-backend-0 || ret_val=$?

if [[ "$ret_val" != 0 ]]; then
  announce "Failed to deploy PostgreSQL backend"
  $cli version
  $cli get statefulset -n "$TEST_APP_NAMESPACE_NAME"
  $cli describe statefulset test-app-backend -n "$TEST_APP_NAMESPACE_NAME"
  $cli get pods -n "$TEST_APP_NAMESPACE_NAME" -l app=test-app-backend
  $cli describe pod test-app-backend-0 -n "$TEST_APP_NAMESPACE_NAME" || true
  $cli logs test-app-backend-0 -n "$TEST_APP_NAMESPACE_NAME" || true
  exit $ret_val
fi

announce "PostgreSQL backend deployed successfully"
