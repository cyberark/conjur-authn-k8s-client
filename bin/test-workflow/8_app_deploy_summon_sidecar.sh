#!/usr/bin/env bash
set -euo pipefail

export PLATFORM="${PLATFORM:-kubernetes}"

. utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_AUTHN_LOGIN_PREFIX

announce "Deploying summon-sidecar test app for $TEST_APP_NAMESPACE_NAME."

set_namespace $TEST_APP_NAMESPACE_NAME

if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^app-summon-sidecar$")" = "app-summon-sidecar" ]; then
    helm uninstall app-summon-sidecar -n "$TEST_APP_NAMESPACE_NAME"
fi

pushd $(dirname "$0")/../../helm/app-deploy > /dev/null
  helm install app-summon-sidecar . -n "$TEST_APP_NAMESPACE_NAME" --debug --wait \
      --set app-summon-sidecar.enabled=true \
      --set global.conjur.conjurConnConfigMap="conjur-connect-configmap" \
      --set app-summon-sidecar.conjur.authnLogin="$CONJUR_AUTHN_LOGIN_PREFIX/test-app-summon-sidecar"
popd > /dev/null

echo "Test app/sidecar deployed."
