#!/usr/bin/env bash

source ./utils.sh

$cli delete namespace "$CONJUR_NAMESPACE_NAME" --ignore-not-found
$cli delete namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found

if [[ "$JAEGER_NAMESPACE_NAME" != "" ]]; then
    $cli delete namespace "$JAEGER_NAMESPACE_NAME" --ignore-not-found
fi
