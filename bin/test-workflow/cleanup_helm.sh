#!/usr/bin/env bash

source ./utils.sh

uninstall_helm_release "cluster-prep-$UNIQUE_TEST_ID" "$CONJUR_NAMESPACE_NAME"
uninstall_helm_release "namespace-prep-$UNIQUE_TEST_ID" "$TEST_APP_NAMESPACE_NAME"
uninstall_helm_release app-backend-pg "$TEST_APP_NAMESPACE_NAME"
uninstall_helm_release test-apps "$TEST_APP_NAMESPACE_NAME"
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  uninstall_helm_release "$HELM_RELEASE" "$CONJUR_NAMESPACE_NAME"
fi
