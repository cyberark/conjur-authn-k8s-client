#!/usr/bin/env bash
set -euo pipefail

. utils.sh

set_namespace default

# Prepare a given namespace with a subset of credentials from the golden configmap
announce "Installing application namespace prep chart"
pushd $(dirname "$0")/../../helm/application-namespace-prep > /dev/null
    if [ "$(helm list -q -n $TEST_APP_NAMESPACE_NAME | grep "^namespace-prep$")" = "namespace-prep" ]; then
        helm uninstall namespace-prep -n "$TEST_APP_NAMESPACE_NAME"
    fi

    # Namespace $TEST_APP_NAMESPACE_NAME will be created if it does not exist
    helm install namespace-prep . -n "$TEST_APP_NAMESPACE_NAME"  --debug --wait \
        --create-namespace \
        --set authnK8s.goldenConfigMap="authn-k8s-configmap" \
        --set authnK8s.namespace="$CONJUR_NAMESPACE"
popd > /dev/null
