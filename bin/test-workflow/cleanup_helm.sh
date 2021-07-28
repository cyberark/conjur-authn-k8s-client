#!/usr/bin/env bash

helm uninstall "cluster-prep-$UNIQUE_TEST_ID" -n "$CONJUR_NAMESPACE_NAME"
helm uninstall "namespace-prep-$UNIQUE_TEST_ID" -n "$TEST_APP_NAMESPACE_NAME"
helm uninstall app-backend-pg -n "$TEST_APP_NAMESPACE_NAME"
