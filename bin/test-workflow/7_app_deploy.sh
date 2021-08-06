#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_AUTHN_LOGIN_PREFIX

announce "Deploying summon-sidecar test app in $TEST_APP_NAMESPACE_NAME."

set_namespace "$TEST_APP_NAMESPACE_NAME"

# Uninstall sample app if it exists
if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^app-summon-sidecar$")" = "app-summon-sidecar" ]; then
    helm uninstall app-summon-sidecar -n "$TEST_APP_NAMESPACE_NAME"
fi

pushd ../../helm/conjur-app-deploy > /dev/null
  helm install app-summon-sidecar . -n "$TEST_APP_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
      --render-subchart-notes \
      --set global.conjur.conjurConnConfigMap="conjur-connect" \
      --set app-summon-sidecar.enabled=true \
      --set app-summon-sidecar.conjur.authnLogin="$CONJUR_AUTHN_LOGIN_PREFIX/test-app-summon-sidecar" \
      --set app-summon-sidecar.app.image.tag="$TEST_APP_TAG" \
      --set app-summon-sidecar.app.image.repository="$TEST_APP_REPO"
      --set app-summon-sidecar.conjur.authnConfigMap.name="conjur-authn-configmap-summon-sidecar"
popd > /dev/null

announce "Deploying secretless-sidecar test app in $TEST_APP_NAMESPACE_NAME."

# Uninstall sample app if it exists
if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^app-secretless-broker$")" = "app-secretless-sidecar" ]; then
    helm uninstall app-secretless-broker -n "$TEST_APP_NAMESPACE_NAME"
fi

pushd ../../helm/conjur-app-deploy > /dev/null
  helm install app-secretless-broker . -n "$TEST_APP_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
      --render-subchart-notes \
      --set global.conjur.conjurConnConfigMap="conjur-connect" \
      --set app-secretless-broker.enabled=true \
      --set app-secretless-broker.conjur.authnLogin="$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secretless-broker" \
      --set app-secretless-broker.conjur.authnConfigMap.name="conjur-authn-configmap-secretless"
popd > /dev/null

echo "Test app/sidecar deployed."
