#!/bin/bash

set -eo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

function setup_conjur_enterprise {
  check_env_var GCLOUD_PROJECT_NAME
  check_env_var GCLOUD_ZONE
  check_env_var GCLOUD_CLUSTER_NAME
  check_env_var GCLOUD_SERVICE_KEY

  docker pull "$CONJUR_APPLIANCE_IMAGE"

  pushd temp > /dev/null
    git clone --single-branch --branch master git@github.com:cyberark/kubernetes-conjur-deploy "kubernetes-conjur-deploy-$UNIQUE_TEST_ID"
  popd > /dev/null

  announce "Deploying Conjur Enterprise"

  if [[ "${CLUSTER_TYPE}" == "gke" ]]; then
    run_command_with_platform "cd temp/kubernetes-conjur-deploy-$UNIQUE_TEST_ID && ./start"
  fi
}

function setup_conjur_open_source {
  pushd temp > /dev/null
    git clone https://github.com/cyberark/conjur-oss-helm-chart.git

    pushd conjur-oss-helm-chart/examples/common > /dev/null
      source ./utils.sh

      announce "Setting demo environment variable defaults"
      source ../kubernetes-in-docker/0_export_env_vars.sh

      announce "Creating a Kubernetes-in-Docker cluster if necessary"
      ./1_create_kind_cluster.sh

      announce "Helm installing/upgrading Conjur OSS cluster"
      ./2_helm_install_or_upgrade_conjur.sh

      # Wait for Conjur pods to become ready (just in case there are old
      # Conjur pods getting terminated as part of Helm upgrade)
      announce "Waiting for Conjur to become ready"
      wait_for_conjur_ready

      announce "Enabling the Conjur Kubernetes authenticator if necessary"
      ./4_ensure_authn_k8s_enabled.sh
    popd > /dev/null

    rm -rf conjur-oss-helm-chart
  popd > /dev/null
}

mkdir -p temp
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  setup_conjur_open_source
else
  setup_conjur_enterprise
fi
