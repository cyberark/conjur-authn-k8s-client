#!/usr/bin/env bash

# When running Conjur outside of a Kubernetes cluster where
# an application is deployed, it requires extra variables to
# connect to the cluster. These are defined in the
# conjur/authn-k8s/<authenticator-id>/kubernetes policy branch.

# This script uses the container with K8s platform tools to
# write these values to files, and uses Conjur CLI to load the
# values into Conjur.

source ./utils.sh

announce "Loading policy values for Conjur-outside-K8s connection."

run_command_with_platform "$cli config view --minify -o json | jq -r '.clusters[0].cluster.server' > kubernetes/api-url"
run_command_with_platform "$cli get secrets -n \"\$CONJUR_NAMESPACE_NAME\" | grep 'conjur.*service-account-token' | head -n1 | awk '{print \$1}' > kubernetes/token-name"
run_command_with_platform "$cli get secret -n \"\$CONJUR_NAMESPACE_NAME\" $(cat kubernetes/token-name) -o json | jq -r .data.token | base64 --decode > kubernetes/service-account-token"

host="$(cat kubernetes/api-url | sed 's/https:\/\///')"
echo -n \
  | openssl s_client -connect "$host:443" -servername "$host" -showcerts 2>/dev/null \
  | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > kubernetes/api-ca.pem
run_command_with_platform "$cli get secret -n \"\$CONJUR_NAMESPACE_NAME\" $(cat kubernetes/token-name) -o json | jq -r '.data[\"ca.crt\"]' | base64 --decode >> kubernetes/api-ca.pem"

# conjur variable values add conjur/authn-k8s/<authenticator>/kubernetes/<var> "<value>"
docker-compose -f "temp/conjur-intro-$UNIQUE_TEST_ID/docker-compose.yml" \
  run --rm \
  -v "${PWD}/kubernetes":/k8s-resources \
  -w /src/cli \
  --entrypoint /bin/bash \
  client -c "
    yes yes | conjur init -u $CONJUR_APPLIANCE_URL -a $CONJUR_ACCOUNT
    conjur authn login -u admin -p $CONJUR_ADMIN_PASSWORD
    conjur variable values add conjur/authn-k8s/$AUTHENTICATOR_ID/kubernetes/ca-cert < /k8s-resources/api-ca.pem
    conjur variable values add conjur/authn-k8s/$AUTHENTICATOR_ID/kubernetes/service-account-token < /k8s-resources/service-account-token
    conjur variable values add conjur/authn-k8s/$AUTHENTICATOR_ID/kubernetes/api-url \"\$(cat /k8s-resources/api-url | tr -d '\n')\"
  "

pushd kubernetes > /dev/null
  rm -f api-url token-name service-account-token api-ca.pem
popd > /dev/null
