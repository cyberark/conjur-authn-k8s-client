#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_AUTHN_LOGIN_PREFIX
check_env_var SECRETS_PROVIDER_TAG
check_env_var SECRETLESS_BROKER_TAG

# Upon error, dump kubernetes resources in the application Namespace
# trap dump_application_namespace_upon_error EXIT

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
  summon_sidecar_jwt_options="--set app-summon-sidecar-jwt.enabled=true \
    --set app-summon-sidecar-jwt.app.image.tag=$TEST_APP_TAG \
    --set app-summon-sidecar-jwt.app.image.repository=$TEST_APP_REPO \
    --set app-summon-sidecar-jwt.app.platform=$PLATFORM"
  secretless_broker_options="--set app-secretless-broker.enabled=true \
    --set app-secretless-broker.secretless.image.tag=$SECRETLESS_BROKER_TAG \
    --set app-secretless-broker.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secretless-broker \
    --set app-secretless-broker.conjur.authnConfigMap.name=conjur-authn-configmap-secretless \
    --set app-secretless-broker.app.platform=$PLATFORM"
  secretless_broker_jwt_options="--set app-secretless-broker-jwt.enabled=true \
      --set app-secretless-broker.secretless.image.tag=$SECRETLESS_BROKER_TAG \
      --set app-secretless-broker.app.platform=$PLATFORM"
  secrets_provider_standalone_options="--set app-secrets-provider-standalone.enabled=true \
    --set app-secrets-provider-standalone.secrets-provider.environment.conjur.authnLogin="$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secrets-provider-standalone" \
    --set app-secrets-provider-standalone.app.image.tag="$TEST_APP_TAG" \
    --set app-secrets-provider-standalone.app.image.repository="$TEST_APP_REPO" \
    --set app-secrets-provider-standalone.app.platform=$PLATFORM"
  secrets_provider_k8s_options="--set app-secrets-provider-k8s.enabled=true \
    --set app-secrets-provider-k8s.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-k8s.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secrets-provider-k8s \
    --set app-secrets-provider-k8s.conjur.authnConfigMap.name=conjur-authn-configmap-secrets-provider-k8s \
    --set app-secrets-provider-k8s.app.platform=$PLATFORM"
  secrets_provider_k8s_jwt_options="--set app-secrets-provider-k8s-jwt.enabled=true \
    --set app-secrets-provider-k8s-jwt.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-k8s-jwt.app.platform=$PLATFORM"
  secrets_provider_p2f_options="--set app-secrets-provider-p2f.enabled=true \
    --set app-secrets-provider-p2f.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-p2f.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secrets-provider-p2f \
    --set app-secrets-provider-p2f.app.platform=$PLATFORM"
  secrets_provider_p2f_injected_options="--set app-secrets-provider-p2f-injected.enabled=true \
    --set app-secrets-provider-p2f-injected.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-p2f-injected.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secrets-provider-p2f-injected \
    --set app-secrets-provider-p2f_injected.app.platform=$PLATFORM"
  secrets_provider_p2f_jwt_options="--set app-secrets-provider-p2f-jwt.enabled=true \
    --set app-secrets-provider-p2f-jwt.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-p2f-jwt.app.platform=$PLATFORM"
  secrets_provider_rotation_options="--set app-secrets-provider-rotation.enabled=true \
    --set app-secrets-provider-rotation.secretsProvider.image.tag=$SECRETS_PROVIDER_TAG \
    --set app-secrets-provider-rotation.conjur.authnLogin=$CONJUR_AUTHN_LOGIN_PREFIX/test-app-secrets-provider-rotation \
    --set app-secrets-provider-rotation.app.platform=$PLATFORM"

  declare -A app_options
  app_options[summon-sidecar]="$summon_sidecar_options"
  app_options[summon-sidecar-jwt]="$summon_sidecar_jwt_options"
  app_options[secretless-broker]="$secretless_broker_options"
  app_options[secretless-broker-jwt]="$secretless_broker_jwt_options"
  app_options[secrets-provider-standalone]="$secrets_provider_standalone_options"
  app_options[secrets-provider-k8s]="$secrets_provider_k8s_options"
  app_options[secrets-provider-k8s-jwt]="$secrets_provider_k8s_jwt_options"
  app_options[secrets-provider-p2f]="$secrets_provider_p2f_options"
  app_options[secrets-provider-p2f-injected]="$secrets_provider_p2f_injected_options"
  app_options[secrets-provider-p2f-jwt]="$secrets_provider_p2f_jwt_options"
  app_options[secrets-provider-rotation]="$secrets_provider_rotation_options"

  # restore array of apps to install
  declare -a install_apps=($(split_on_comma_delimiter $INSTALL_APPS))
  options_string=""
  for app in "${install_apps[@]}"; do
    # If application that uses Secrets Provider in standalone mode is enabled,
    # make sure that the Secrets Provider Helm chart has been downloaded as a
    # dependency for that application's subchart.
    if [ "$app" = "secrets-provider-standalone" ]; then
      pushd charts/app-secrets-provider-standalone > /dev/null
        if ! ls charts/*.tgz 1>/dev/null 2>&1; then
          announce "Downloading Secrets Provider Helm chart"
          helm repo add cyberark https://cyberark.github.io/helm-charts
          helm repo update
          helm dependency update . --skip-refresh
        fi
      popd
    fi
    options_string+="${app_options[$app]} "
  done

  announce "Deploying test apps in $TEST_APP_NAMESPACE_NAME timeout is $TIMEOUT"
  helm install test-apps . -n "$TEST_APP_NAMESPACE_NAME" --wait --timeout "$TIMEOUT" \
    --render-subchart-notes \
    --set global.conjur.conjurConnConfigMap="conjur-connect" \
    $options_string

popd > /dev/null

echo "Test apps deployed."
