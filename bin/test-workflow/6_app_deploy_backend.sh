#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var SAMPLE_APP_BACKEND_DB_PASSWORD

announce "Deploying test app postgres backend for $TEST_APP_NAMESPACE_NAME."

set_namespace "$TEST_APP_NAMESPACE_NAME"

app_name="app-backend-pg"

# Uninstall backend if it exists so any PVCs can be deleted
if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^$app_name$")" = "$app_name" ]; then
    helm uninstall "$app_name" -n "$TEST_APP_NAMESPACE_NAME"
fi

# Delete any created PVCs
$cli delete --namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found \
  pvc -l app.kubernetes.io/instance="$app_name"

echo "Create secrets for test app backend"
$cli delete --namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found \
  secret test-app-backend-certs

$cli --namespace "$TEST_APP_NAMESPACE_NAME" \
  create secret generic \
  test-app-backend-certs \
  --from-file=server.crt=./etc/ca.pem \
  --from-file=server.key=./etc/ca-key.pem

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install "$app_name" bitnami/postgresql -n "$TEST_APP_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
    --set image.repository="postgres" \
    --set image.tag="9.6" \
    --set postgresqlDataDir="/data/pgdata" \
    --set persistence.mountPath="/data/" \
    --set fullnameOverride="test-app-backend" \
    --set tls.enabled=true \
    --set volumePermissions.enabled=true \
    --set tls.certificatesSecret="test-app-backend-certs" \
    --set tls.certFilename="server.crt" \
    --set tls.certKeyFilename="server.key" \
    --set securityContext.fsGroup="999" \
    --set postgresqlDatabase="test_app" \
    --set postgresqlUsername="test_app" \
    --set postgresqlPassword="$SAMPLE_APP_BACKEND_DB_PASSWORD"

