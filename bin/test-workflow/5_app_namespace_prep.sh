#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var TEST_APP_NAMESPACE_NAME
check_env_var CONJUR_NAMESPACE_NAME

set_namespace default

# Prepare a given namespace with a subset of credentials from the golden configmap
announce "Installing namespace prep chart"
pushd ../../helm/conjur-config-namespace-prep > /dev/null
    # Namespace $TEST_APP_NAMESPACE_NAME will be created if it does not exist
    helm upgrade --install "namespace-prep-$UNIQUE_TEST_ID" . -n "$TEST_APP_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
        --create-namespace \
        --set authnK8s.goldenConfigMap="conjur-configmap" \
        --set authnK8s.namespace="$CONJUR_NAMESPACE_NAME"

popd > /dev/null
