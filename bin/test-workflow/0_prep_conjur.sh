#!/bin/bash

# Install Conjur in our cluster, and load policies
mkdir -p temp &2> /dev/null
pushd temp
    git clone https://github.com/cyberark/conjur-oss-helm-chart.git &2> /dev/null

    pushd conjur-oss-helm-chart/examples/kubernetes-in-docker
        announce "Installing Conjur-OSS"
        helm uninstall conjur-oss
        
        announce "Setting demo environment variable defaults"
        . ./0_export_env_vars.sh

        announce "Creating a Kubernetes-in-Docker cluster if necessary"
        ./1_create_kind_cluster.sh

        announce "Helm installing/upgrading Conjur OSS cluster"
        ./2_helm_install_or_upgrade_conjur.sh

        # Wait for Conjur pods to become ready (just in case there are old
        # Conjur pods getting terminated as part of Helm upgrade)
        announce "Waiting for Conjur to become ready"
        wait_for_conjur_ready

        announce "Retrieving the Conjur admin password"
        export CONJUR_ADMIN_PASSWORD="$(./3_retrieve_admin_password.sh)"

        announce "Enabling the Conjur Kubernetes authenticator if necessary"
        ./4_ensure_authn_k8s_enabled.sh

    popd
popd
