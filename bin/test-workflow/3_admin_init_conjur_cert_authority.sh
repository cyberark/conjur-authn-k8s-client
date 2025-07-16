#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

check_env_var CONJUR_NAMESPACE_NAME
check_env_var CONJUR_OSS_HELM_INSTALLED
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

TEST_JWT_FLOW="${TEST_JWT_FLOW:-false}"

# Upon error, dump kubernetes resources in the Conjur Namespace
trap dump_conjur_namespace_upon_error EXIT

announce "Initializing Conjur certificate authority."

if [[ "$CONJUR_PLATFORM" != "jenkins" ]]; then
  set_namespace $CONJUR_NAMESPACE_NAME
  conjur_master="$(get_master_pod_name)"
fi

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]] && [[ "$TEST_JWT_FLOW" == "false" ]]; then
  $cli exec "$conjur_master" -c conjur-oss -- bash -c "CONJUR_ACCOUNT=$CONJUR_ACCOUNT bundle exec rake authn_k8s:ca_init['conjur/authn-k8s/$AUTHENTICATOR_ID']"
elif [[ "$CONJUR_PLATFORM" == "gke" ]]; then
  $cli exec "$conjur_master" -- chpst -u api:conjur conjur-plugin-service possum rake authn_k8s:ca_init["conjur/authn-k8s/$AUTHENTICATOR_ID"]
elif [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
  docker compose -f temp/conjur-intro-$UNIQUE_TEST_ID/docker-compose.yml \
    exec -T conjur-master-1.mycompany.local chpst -u api:conjur conjur-plugin-service possum rake authn_k8s:ca_init["conjur/authn-k8s/$AUTHENTICATOR_ID"]
fi

echo "Certificate authority initialized."
