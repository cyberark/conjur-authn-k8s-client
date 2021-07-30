#!/usr/bin/env bash

source ./utils.sh

$cli delete namespace "$CONJUR_NAMESPACE_NAME"
$cli delete namespace "$TEST_APP_NAMESPACE_NAME"
