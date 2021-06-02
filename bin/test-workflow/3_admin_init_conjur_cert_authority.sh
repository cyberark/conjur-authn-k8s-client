#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

PLATFORM="${PLATFORM:-kubernetes}"

source utils.sh

check_env_var CONJUR_NAMESPACE
check_env_var CONJUR_OSS_HELM_INSTALLED
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

announce "Initializing Conjur certificate authority."

set_namespace $CONJUR_NAMESPACE

conjur_master=$(get_master_pod_name)

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
    $cli exec $conjur_master -c conjur-oss -- bash -c "CONJUR_ACCOUNT=$CONJUR_ACCOUNT rake authn_k8s:ca_init['conjur/authn-k8s/$AUTHENTICATOR_ID']"
else
    $cli exec $conjur_master -- chpst -u conjur conjur-plugin-service possum rake authn_k8s:ca_init["conjur/authn-k8s/$AUTHENTICATOR_ID"]
fi

echo "Certificate authority initialized."
