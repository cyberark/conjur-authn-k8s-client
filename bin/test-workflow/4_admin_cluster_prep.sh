#!/usr/bin/env bash
set -euo pipefail

export PLATFORM="${PLATFORM:-kubernetes}"
export TIMEOUT="${TIMEOUT:-5m0s}"

. utils.sh

check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_NAMESPACE
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

set_namespace default

# Prepare our cluster with conjur and authnK8s credentials in a golden configmap
announce "Installing cluster prep chart"
pushd $(dirname "$0")/../../helm/kubernetes-cluster-prep > /dev/null
    ./bin/get-conjur-cert.sh -v -i -u "$CONJUR_APPLIANCE_URL"

    helm upgrade --install cluster-prep . -n "$CONJUR_NAMESPACE"  --debug --wait \
        --set conjur.account="$CONJUR_ACCOUNT" \
        --set conjur.applianceUrl="$CONJUR_APPLIANCE_URL" \
        --set conjur.certificateFilePath="files/conjur-cert.pem" \
        --set authnK8s.authenticatorID="$AUTHENTICATOR_ID" \
        --timeout $TIMEOUT
popd > /dev/null
