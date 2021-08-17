#!/usr/bin/env bash

source ./utils.sh

$cli delete namespace "$CONJUR_NAMESPACE_NAME" --ignore-not-found
$cli delete namespace "$TEST_APP_NAMESPACE_NAME" --ignore-not-found
