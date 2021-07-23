#!/usr/bin/env bash

if [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
  cluster_prep_namespace="$TEST_APP_NAMESPACE_NAME"
elif [[ "$CONJUR_PLATFORM" == "gke" ]]; then
  cluster_prep_namespace="$CONJUR_NAMESPACE_NAME"
fi

helm uninstall "cluster-prep-$UNIQUE_TEST_ID" -n "$cluster_prep_namespace"
helm uninstall "namespace-prep-$UNIQUE_TEST_ID" -n "$TEST_APP_NAMESPACE_NAME"
helm uninstall app-backend-pg -n "$TEST_APP_NAMESPACE_NAME"
