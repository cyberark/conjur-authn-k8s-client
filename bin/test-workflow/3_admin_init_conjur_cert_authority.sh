#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

check_env_var CONJUR_NAMESPACE_NAME
check_env_var CONJUR_OSS_HELM_INSTALLED
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

announce "Initializing Conjur certificate authority."

if [[ "$CONJUR_PLATFORM" != "jenkins" ]]; then
  set_namespace $CONJUR_NAMESPACE_NAME
  conjur_master="$(get_master_pod_name)"
fi

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  $cli exec "$conjur_master" -c conjur-oss -- bash -c "CONJUR_ACCOUNT=$CONJUR_ACCOUNT rake authn_k8s:ca_init['conjur/authn-k8s/$AUTHENTICATOR_ID']"
elif [[ "$CONJUR_PLATFORM" == "gke" ]]; then
  $cli exec "$conjur_master" -- chpst -u conjur conjur-plugin-service possum rake authn_k8s:ca_init["conjur/authn-k8s/$AUTHENTICATOR_ID"]
elif [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
  docker-compose -f temp/conjur-intro-$UNIQUE_TEST_ID/docker-compose.yml \
    exec -T conjur-master-1.mycompany.local chpst -u conjur conjur-plugin-service possum rake authn_k8s:ca_init["conjur/authn-k8s/$AUTHENTICATOR_ID"]
fi

echo "Certificate authority initialized."
