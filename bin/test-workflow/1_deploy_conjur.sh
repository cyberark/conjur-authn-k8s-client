#!/bin/bash

set -eo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

function setup_conjur_enterprise {
  docker pull "$CONJUR_APPLIANCE_IMAGE"

  announce "Deploying Conjur Enterprise"

  # deploy Conjur to GKE cluster
  if [[ "${CONJUR_PLATFORM}" == "gke" ]]; then
    check_env_var GCLOUD_PROJECT_NAME
    check_env_var GCLOUD_ZONE
    check_env_var GCLOUD_CLUSTER_NAME
    check_env_var GCLOUD_SERVICE_KEY

    pushd temp > /dev/null
      git clone --single-branch --branch master git@github.com:cyberark/kubernetes-conjur-deploy "kubernetes-conjur-deploy-$UNIQUE_TEST_ID"
    popd > /dev/null

    run_command_with_platform "cd temp/kubernetes-conjur-deploy-$UNIQUE_TEST_ID && ./start"

  # deploy Conjur locally
  elif [[ "${CONJUR_PLATFORM}" == "jenkins" ]]; then
    check_env_var HOST_IP

    pushd temp > /dev/null
      # TODO - once these changes are merged, this branch has to be updated to main
      git clone --single-branch --branch custom-port-follower git@github.com:conjurdemos/conjur-intro.git "conjur-intro-$UNIQUE_TEST_ID"

      pushd "conjur-intro-$UNIQUE_TEST_ID" > /dev/null

        # add public IP address to custom certificate config as SAN
        docker run --rm \
          -v "${PWD}":/src \
          -w /src/artifacts/certificate-generator/configuration \
          "custom-certs" \
          ash -c "
            jq '.hosts[.hosts| length] |= . + \"${HOST_IP}\"' dap-follower.json > tmp
            mv tmp dap-follower.json
          "

        echo """
CONJUR_MASTER_PORT=\"${CONJUR_MASTER_PORT}\"
CONJUR_FOLLOWER_PORT=\"${CONJUR_FOLLOWER_PORT}\"
CONJUR_AUTHENTICATORS=authn-k8s/\"${AUTHENTICATOR_ID}\",authn
        """ > .env
        ./bin/dap --provision-master
        ./bin/dap --import-custom-certificates
        ./bin/dap --provision-follower
      popd > /dev/null

    popd > /dev/null
  fi
}

function setup_conjur_open_source {
  pushd temp > /dev/null
    git clone --single-branch --branch main https://github.com/cyberark/conjur-oss-helm-chart.git "conjur-oss-helm-chart-$UNIQUE_TEST_ID"

    pushd "conjur-oss-helm-chart-$UNIQUE_TEST_ID/examples/common" > /dev/null
      source ./utils.sh

      announce "Setting demo environment variable defaults"

      if [[ "$PLATFORM" == "openshift" ]]; then
        announce "Using OpenShift"
        source ../openshift/0_export_env_vars.sh
      else
        source ../kubernetes-in-docker/0_export_env_vars.sh
        announce "Creating a Kubernetes-in-Docker cluster if necessary"
        ./1_create_kind_cluster.sh
       fi

      announce "Helm installing/upgrading Conjur OSS cluster"
      ./2_helm_install_or_upgrade_conjur.sh

      # Wait for Conjur pods to become ready (just in case there are old
      # Conjur pods getting terminated as part of Helm upgrade)
      announce "Waiting for Conjur to become ready"
      wait_for_conjur_ready

      announce "Enabling the Conjur Kubernetes authenticator if necessary"
      ./4_ensure_authn_k8s_enabled.sh
    popd > /dev/null
  popd > /dev/null
}

mkdir -p temp
if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  setup_conjur_open_source
else
  setup_conjur_enterprise
fi
