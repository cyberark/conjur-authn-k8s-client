#!/bin/bash

set -eo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

source utils.sh

# Upon error, dump kubernetes resources in the Conjur Namespace
trap dump_conjur_namespace_upon_error EXIT

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
      git clone --single-branch --branch main https://github.com/cyberark/sidecar-injector.git "sidecar-injector-$UNIQUE_TEST_ID"
    popd > /dev/null

    run_command_with_platform "cd temp/kubernetes-conjur-deploy-$UNIQUE_TEST_ID && ./start"

  # deploy Conjur locally
  elif [[ "${CONJUR_PLATFORM}" == "jenkins" ]]; then
    check_env_var HOST_IP

    pushd temp > /dev/null
      git clone --single-branch --branch main git@github.com:conjurdemos/conjur-intro.git "conjur-intro-$UNIQUE_TEST_ID"
      git clone --single-branch --branch main https://github.com/cyberark/sidecar-injector.git "sidecar-injector-$UNIQUE_TEST_ID"

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
CONJUR_AUTHENTICATORS=authn-k8s/\"${AUTHENTICATOR_ID}\",authn-jwt/\"${AUTHENTICATOR_ID}\",authn
        """ > .env
        ./bin/dap --provision-master --version "${CONJUR_APPLIANCE_TAG}"
        ./bin/dap --import-custom-certificates
        ./bin/dap --wait-for-master
        ./bin/dap --provision-follower --version "${CONJUR_APPLIANCE_TAG}"
      popd > /dev/null

    popd > /dev/null
  fi
}

function setup_conjur_open_source {
  # Pin Conjur OSS to specific version to avoid issues with latest
  export IMAGE_TAG=1.19.3

  pushd temp > /dev/null
    git clone --single-branch --branch main https://github.com/cyberark/conjur-oss-helm-chart.git "conjur-oss-helm-chart-$UNIQUE_TEST_ID"
    git clone --single-branch --branch main https://github.com/cyberark/sidecar-injector.git "sidecar-injector-$UNIQUE_TEST_ID"
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

      if [[ "$TEST_JWT_FLOW" == "true" ]]; then
        announce "Enable authn-jwt in conjur instead of authn-k8s"
        export AUTHN_STRATEGY="authn-jwt"
        announce "Allow access to jwks uri for unauthenticated users"
        kubectl delete clusterrolebinding oidc-reviewer --ignore-not-found
        kubectl create clusterrolebinding oidc-reviewer --clusterrole=system:service-account-issuer-discovery --group=system:unauthenticated
      fi

      announce "Helm installing/upgrading Conjur OSS cluster"
      ./2_helm_install_or_upgrade_conjur.sh

      # Wait for Conjur pods to become ready (just in case there are old
      # Conjur pods getting terminated as part of Helm upgrade)
      announce "Waiting for Conjur to become ready"
      wait_for_conjur_ready
    popd > /dev/null
  popd > /dev/null
}

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
  setup_conjur_open_source
else
  setup_conjur_enterprise
fi
