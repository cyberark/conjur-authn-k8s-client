#!/usr/bin/env bash
set -euo pipefail

. utils.sh

announce "Deploying summon-sidecar test app postgres backend for $TEST_APP_NAMESPACE_NAME."

set_namespace $TEST_APP_NAMESPACE_NAME

echo "Deploying test app backend"

# Install postgresql helm chart
if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^app-summon-sidecar-backend-pg$")" = "app-summon-sidecar-backend-pg" ]; then
    helm uninstall app-summon-sidecar-backend-pg -n "$TEST_APP_NAMESPACE_NAME"
fi

$cli delete --namespace $TEST_APP_NAMESPACE_NAME --ignore-not-found \
  pvc -l app.kubernetes.io/instance=app-summon-sidecar-backend-pg

helm repo add bitnami https://charts.bitnami.com/bitnami

helm install app-summon-sidecar-backend-pg bitnami/postgresql -n $TEST_APP_NAMESPACE_NAME --debug --wait \
    --set image.repository="postgres" \
    --set image.tag="9.6" \
    --set postgresqlDataDir="/data/pgdata" \
    --set persistence.mountPath="/data/" \
    --set fullnameOverride="test-summon-sidecar-app-backend" \
    --set tls.enabled=true \
    --set volumePermissions.enabled=true \
    --set tls.certificatesSecret="test-app-backend-certs" \
    --set tls.certFilename="server.crt" \
    --set tls.certKeyFilename="server.key" \
    --set securityContext.fsGroup="999" \
    --set postgresqlDatabase="test_app" \
    --set postgresqlUsername="test_app" \
    --set postgresqlPassword=$SAMPLE_APP_BACKEND_DB_PASSWORD
