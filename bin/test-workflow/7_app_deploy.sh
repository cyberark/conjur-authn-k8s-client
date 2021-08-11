#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_AUTHN_LOGIN_PREFIX

set_namespace "$TEST_APP_NAMESPACE_NAME"

pushd ../../helm/conjur-app-deploy > /dev/null

  # Uninstall any existing sample apps
  if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^test-apps$")" = "test-apps" ]; then
    helm uninstall test-apps -n "$TEST_APP_NAMESPACE_NAME"
  fi

  summon_sidecar_options="--set app-summon-sidecar.enabled=true \
    --set app-summon-sidecar.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-summon-sidecar \
    --set app-summon-sidecar.app.image.tag=$TEST_APP_TAG \
    --set app-summon-sidecar.app.image.repository=$TEST_APP_REPO \
    --set app-summon-sidecar.conjur.authnConfigMap.name=conjur-authn-configmap-summon-sidecar \
    --set app-summon-sidecar.app.platform=$PLATFORM"
  secretless_broker_options="--set app-secretless-broker.enabled=true \
    --set app-secretless-broker.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secretless-broker \
    --set app-secretless-broker.conjur.authnConfigMap.name=conjur-authn-configmap-secretless \
    --set app-secretless-broker.app.platform=$PLATFORM"

  declare -A app_options
  app_options[summon-sidecar]="$summon_sidecar_options"
  app_options[secretless-broker]="$secretless_broker_options"

  # restore array of apps to install
  declare -a install_apps=($(split_on_comma_delimiter $INSTALL_APPS))
  options_string=""
  for app in "${install_apps[@]}"; do
    options_string+="${app_options[$app]} "
  done

  announce "Deploying test apps in $TEST_APP_NAMESPACE_NAME"
  helm install test-apps . -n "$TEST_APP_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
    --render-subchart-notes \
    --set global.conjur.conjurConnConfigMap="conjur-connect" \
    $options_string

popd > /dev/null

echo "Test app/sidecar deployed."
