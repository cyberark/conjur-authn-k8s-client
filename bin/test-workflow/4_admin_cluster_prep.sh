#!/usr/bin/env bash

set -euox pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"
TEST_CLIENT_IMAGE_TAG="${TEST_CLIENT_IMAGE_TAG:-edge}"

source utils.sh

check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_NAMESPACE_NAME
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "false" ]]; then
  check_env_var CONJUR_FOLLOWER_URL
fi

set_namespace default

if [[ "$TEST_CLIENT_IMAGE_TAG" != "edge" ]]; then
  announce "Pushing test client image conjur-k8s-cluster-test:$TEST_CLIENT_IMAGE_TAG"

  source_image="conjur-k8s-cluster-test"
  test_client_image="$(platform_image_for_push $source_image $CONJUR_NAMESPACE_NAME)"
  docker tag "$source_image:$TEST_CLIENT_IMAGE_TAG" "$test_client_image"

  docker push "$test_client_image"

  docker image ls
fi

# Prepare our cluster with conjur and authnK8s credentials in a golden configmap
announce "Installing cluster prep chart"
pushd ../../helm/conjur-config-cluster-prep > /dev/null

  if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
    conjur_url="$CONJUR_APPLIANCE_URL"
    get_cert_options="-v -i -s -t $CONJUR_NAMESPACE_NAME -u"
    service_account_options=""
  else
    conjur_url="$CONJUR_FOLLOWER_URL"
    if [[ "$CONJUR_PLATFORM" == "gke" ]]; then
      get_cert_options="-v -i -s -t $CONJUR_NAMESPACE_NAME -u"
      service_account_options="--set authnK8s.serviceAccount.create=false --set authnK8s.serviceAccount.name=conjur-cluster"
    elif [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
      get_cert_options="-v -s -t $CONJUR_NAMESPACE_NAME -u"
      service_account_options=""
    fi
  fi

  ./bin/get-conjur-cert.sh $get_cert_options "$conjur_url"
  helm upgrade --install "cluster-prep-$UNIQUE_TEST_ID" . -n "$CONJUR_NAMESPACE_NAME" --debug --wait --timeout "$TIMEOUT" \
      --create-namespace \
      --set conjur.account="$CONJUR_ACCOUNT" \
      --set conjur.applianceUrl="$conjur_url" \
      --set conjur.certificateFilePath="files/conjur-cert.pem" \
      --set authnK8s.authenticatorID="$AUTHENTICATOR_ID" \
      --set authnK8s.clusterRole.name="conjur-clusterrole-$UNIQUE_TEST_ID" \
      $service_account_options

popd > /dev/null
