#!/usr/bin/env bash
set -euo pipefail

. utils.sh

announce "Creating Test App namespace."

set_namespace default

conjur_appliance_url="${CONJUR_APPLIANCE_URL:-https://conjur-oss.$CONJUR_NAMESPACE.svc.cluster.local}"

pushd helm
    # Prepare our cluster with conjur and authnK8s credentials in a golden configmap
    pushd kubernetes-cluster-prep
        announce "Installing cluster prep chart"
        helm uninstall cluster-prep -n "$CONJUR_NAMESPACE"

        ./bin/get-conjur-cert.sh -v -i -u "$conjur_appliance_url"

        helm install cluster-prep . -n "$CONJUR_NAMESPACE"  --wait \
            --set conjur.account="$CONJUR_ACCOUNT" \
            --set conjur.applianceUrl="$conjur_appliance_url" \
            --set conjur.certificateFilePath="files/conjur-cert.pem" \
            --set authnK8s.authenticatorID="$AUTHENTICATOR_ID"
    popd

    # Prepare a given namespace with a subset of credentials from the golden configmap
    pushd application-namespace-prep
        announce "Installing application namespace prep chart"
        helm uninstall namespace-prep -n "$TEST_APP_NAMESPACE_NAME"

        helm install namespace-prep . -n "$TEST_APP_NAMESPACE_NAME"  --wait \
            --set authnK8s.goldenConfigMap="$TEST_APP_NAMESPACE_NAME" \
            --set authnK8s.namespace="$CONJUR_NAMESPACE"
    popd
popd
