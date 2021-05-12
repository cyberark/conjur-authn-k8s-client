#!/usr/bin/env bash
set -euo pipefail

. utils.sh

set_namespace default

# Prepare our cluster with conjur and authnK8s credentials in a golden configmap
announce "Installing cluster prep chart"
pushd helm/kubernetes-cluster-prep
    helm uninstall cluster-prep -n "$CONJUR_NAMESPACE"

    ./bin/get-conjur-cert.sh -v -i -u "$CONJUR_APPLIANCE_URL"

    helm install cluster-prep . -n "$CONJUR_NAMESPACE"  --wait \
        --set conjur.account="$CONJUR_ACCOUNT" \
        --set conjur.applianceUrl="$CONJUR_APPLIANCE_URL" \
        --set conjur.certificateFilePath="files/conjur-cert.pem" \
        --set authnK8s.authenticatorID="$AUTHENTICATOR_ID"
popd
