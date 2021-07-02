#!/bin/bash

set -o pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

function setup_kind {
  pushd conjur-oss-helm-chart/examples/common > /dev/null
    source ./utils.sh

    announce "Setting demo environment variable defaults"
    source ../kubernetes-in-docker/0_export_env_vars.sh

    announce "Creating a Kubernetes-in-Docker cluster if necessary"
    ./1_create_kind_cluster.sh
  popd > /dev/null
}

function setup_conjur_enterprise_in_cluster {
  rm -rf kubernetes-conjur-deploy
  git clone https://github.com/cyberark/kubernetes-conjur-deploy.git

  pushd kubernetes-conjur-deploy > /dev/null
    export CONJUR_APPLIANCE_IMAGE="registry2.itci.conjur.net/conjur-appliance:5.0-stable"
    export CONJUR_FOLLOWER_COUNT=1
    export CONJUR_AUTHN_LOGIN="host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/$CONJUR_NAMESPACE_NAME/service_account/conjur-cluster"
    export STOP_RUNNING_ENV="true"
    export DEPLOY_MASTER_CLUSTER="true"
    export CONJUR_NAMESPACE_NAME="$CONJUR_NAMESPACE"

    announce "Deploying Conjur Enterprise to the KinD cluster"
    ./start
  popd > /dev/null
}

function setup_conjur_open_source {
  pushd conjur-oss-helm-chart/examples/common > /dev/null
    announce "Helm installing/upgrading Conjur OSS cluster"
    ./2_helm_install_or_upgrade_conjur.sh

    # Wait for Conjur pods to become ready (just in case there are old
    # Conjur pods getting terminated as part of Helm upgrade)
    announce "Waiting for Conjur to become ready"
    wait_for_conjur_ready

    announce "Enabling the Conjur Kubernetes authenticator if necessary"
    ./4_ensure_authn_k8s_enabled.sh
  popd > /dev/null
}

mkdir -p temp
pushd temp > /dev/null

  rm -rf conjur-oss-helm-chart
  git clone https://github.com/cyberark/conjur-oss-helm-chart.git
  setup_kind

  if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
    setup_conjur_open_source
  else
    setup_conjur_enterprise_in_cluster
  fi

popd > /dev/null
