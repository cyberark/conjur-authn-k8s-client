#!/usr/bin/env bash
set -euo pipefail

. utils.sh

set_namespace default

# Prepare a given namespace with a subset of credentials from the golden configmap
announce "Installing application namespace prep chart"
pushd helm/application-namespace-prep
    helm uninstall namespace-prep -n "$TEST_APP_NAMESPACE_NAME"

    # Namespace $TEST_APP_NAMESPACE_NAME will be created if it does not exist
    helm install namespace-prep . -n "$TEST_APP_NAMESPACE_NAME"  --wait \
        --set authnK8s.goldenConfigMap="$TEST_APP_NAMESPACE_NAME" \
        --set authnK8s.namespace="$CONJUR_NAMESPACE"
popd
