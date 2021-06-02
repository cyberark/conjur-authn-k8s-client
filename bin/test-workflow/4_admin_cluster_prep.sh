#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"
TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_NAMESPACE
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

set_namespace default

# Prepare our cluster with conjur and authnK8s credentials in a golden configmap
announce "Installing cluster prep chart"
pushd ../../helm/conjur-config-cluster-prep > /dev/null
    ./bin/get-conjur-cert.sh -v -i -s -u "$CONJUR_APPLIANCE_URL"

    helm upgrade --install cluster-prep . -n "$CONJUR_NAMESPACE" --debug --wait --timeout $TIMEOUT \
        --set conjur.account="$CONJUR_ACCOUNT" \
        --set conjur.applianceUrl="$CONJUR_APPLIANCE_URL" \
        --set conjur.certificateFilePath="files/conjur-cert.pem" \
        --set authnK8s.authenticatorID="$AUTHENTICATOR_ID"
        
popd > /dev/null
