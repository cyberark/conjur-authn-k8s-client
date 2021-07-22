# Conjur Authenticator End-to-End Workflow

## Table of Contents

* [Overview](#overview)
* [Ongoing Improvements](ongoing-improvements)
* [Quick Start Guide](#quick-start-guide)
  + [Prerequisites](#prerequisites)
  + [Steps](#steps)

## Overview

The scripts within this folder encompass an end-to-end workflow for testing Conjur Kubernetes authentication by deploying a PetStore demo application to a Kubernetes cluster. The authenticator and cluster configuration is validated at a high level via basic POST and GET requests to the PetStore app; implicitly verifying communication with the app's backend database (with credentials provided by a particular Conjur Kubernetes authenticator).

The workflow: 
* Deploys Conjur to a Kubernetes cluster
* Prepares the cluster with Conjur Config Cluster Prep Helm chart
* Prepares and enables the Kubernetes authenticator in Conjur
* Prepares PetStore app namespace with Conjur NameSpace Prep Helm chart
* Deploys and verifies the PetStore demo application with authenticator sidecar

The workflow currently supports testing Kubernetes authentication against Conjur Open Source or Enterprise. Each can be run from the `start` script:

```bash
# run Open Source workflow
./start
# run Enterprise workflow on GKE
./start --enterprise --platform gke
```

## Ongoing Improvements

* Demo app can be published to DockerHub instead of local builds
  ([#318](https://github.com/cyberark/conjur-authn-k8s-client/issues/318))
* Refactor the end-to-end scripts for groups of commands by persona responsibility
  ([#322](https://github.com/cyberark/conjur-authn-k8s-client/issues/322) &
  [#323](https://github.com/cyberark/conjur-authn-k8s-client/issues/323))

## Prerequisites

#### Common
* [git](https://git-scm.com/downloads)
* [Docker](https://docs.docker.com/get-docker/)

#### Conjur Open Source Workflow
* [Kubernetes in Docker (KinD)](https://github.com/kubernetes-sigs/kind#installation-and-usage)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [Helm](https://helm.sh/docs/intro/install/)

#### Conjur Enterprise Workflow
* Google Kubernetes Engine (GKE) cluster access, which requires the following environment variables to be set:
  * `GCLOUD_PROJECT_NAME`
  * `GCLOUD_ZONE`
  * `GCLOUD_CLUSTER_NAME`
  * `GCLOUD_SERVICE_KEY`

## Quick Start Guide

### Steps

1) Prepare Environment

    - The following scripts use environment variables to persist information regarding the workflow's configuration. Each is set to a default value, and can be changed by setting the envvar before invoking the script.

    - Prepare the environment by running:
      ```bash
      source ./0_prep_env.sh
      ```

2) Deploy Conjur to Kubernetes Cluster

    - The workflow can either deploy Conjur Open Source or Conjur Enterprise, and decides based on the `CONJUR_OSS_HELM_INSTALLED` environment variable
        - Deploy Conjur Open Source to KinD
            - Start a KinD cluster with local Docker registry
            - Create a new namespace for Conjur
            - Deploy Conjur Open Source with Helm
                - The Conjur Open Source Helm Chart is published on [GitHub](https://github.com/cyberark/conjur-oss-helm-chart).

                - The Conjur Open Source Helm Chart repository contains an [example](https://github.com/cyberark/conjur-oss-helm-chart/tree/main/examples/kubernetes-in-docker) folder with scripts and instructions for deploying the Conjur Open Source Helm Chart on KinD. The scripts from the example folder are used to accomplish this step by git cloning the Conjur Open Source Helm Chart repository and using them to carry out the tasks mentioned above.
        - Deploy Conjur Enterprise to GKE
            - Create a new namespace for Conjur
            - Deploy Conjur Enterprise
                - Conjur Enterprise is deployed with scripts in the [Kubernetes Conjur Deploy GitHub repo](https://github.com/cyberark/kubernetes-conjur-deploy).
    - Enable the Kubernetes Authenticator in Conjur
    - To perform these steps, run:
      ```bash
      ./1_deploy_conjur.sh
      ```

3) Load Conjur Policy

    - This step loads Conjur policy in order to:

        - [Prepare Kubernetes authenticator in Conjur](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/bin/test-workflow/policy/templates/cluster-authn-svc-def.template.yml)
        - [Prepare PetStore app identities](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/bin/test-workflow/policy/templates/project-authn-def.template.yml)
        - [Permit identities to use the authenticator](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/bin/test-workflow/policy/templates/app-identity-def.template.yml)
        - [Prepare app credentials in Conjur](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/bin/test-workflow/policy/app-access.yml)

    - To load all Conjur policy needed for the demo app to use the Conjur Kubernetes sidecar authenticator, run:
      ```bash
      ./2_admin_load_conjur_policies.sh
      ```

    - When successful, the script will output the following:
      ```
      Success!
      Conjur policy loaded.
      ```

4) Initialize Kubernetes Authenticator Certificate Authority

    - To initialize the Kubenetes authenticator CA, run:
      ```bash
      ./3_admin_init_cojur_cert_authority.sh
      ```

    - When successful, the script will output the following:
      ```
      Certificate authority initialized.
      ```

5) Cluster Preparation

    - In this step, the Kubernetes cluster is prepared to enable applications to authenticate with Conjur Open Source using:

        - a "Golden" ConfigMap
        - an authenticator ClusterRole
        - an authenticator ServiceAccount

    - To perform this setup, run:
      ```bash
      ./4_admin_cluster_prep.sh
      ```

    - When successful, the script will output the following:
      ```
      NOTES:
      The Conjur/Authenticator Namespace preparation is complete.
      The following have been deployed:

      A Golden ConfigMap

      An authenticator ClusterRole

      An authenticator ServiceAccount
      ```

    - More context on this step can be found in the [`helm/conjur-config-cluster-prep`](https://github.com/cyberark/conjur-authn-k8s-client/tree/master/helm/conjur-config-cluster-prep) directory.

6) App Namespace Preparation

    - In this step, a new namespace is created in the Kubernetes cluster for the PetStore test app deployment, and it is prepared to authenticate with Conjur Open Source using:

        - a Conjur connection ConfigMap
        - an authenticator RoleBinding

    - To perform this setup, run:
      ```bash
      ./5_app_namespace_prep.sh
      ```

    - When successful, the script will output the following:
      ```
      NOTES:
      The Application Namespace preparation is complete.
      The following have been deployed:
      A Conjur Connection ConfigMap
      An authenticator RoleBinding
      A Secret containing the sample app backend TLS certificate and key
      ```

    - More context on this step can be found in the [`helm/conjur-config-namespace-prep`](https://github.com/cyberark/conjur-authn-k8s-client/tree/master/helm/conjur-config-namespace-prep) directory.

7) Deploy PetStore App

    - At this point in the workflow:

        - Conjur Open Source has been deployed to its own namespace

        - the `conjur-oss` namespace has been configured for Conjur Kubernetes authentication

        - an application namespace has been created and prepared for connecting to and authenticating with Conjur Open Source

    - The following steps cover deploying the PetStore test app and the Conjur
Kubernetes sidecar authenticator, and are performed by the Kubernetes Admin.

        1) Build App Image and Push to Docker Registry

            - In this step, the demo PetStore app image, which includes Summon binaries, is built and pushed to the local Docker registry.

            - The app is built and pushed by running:
              ```bash
              ./6_app_build_and_push_containers.sh
              ```

        2) Deploy App Backend

            - Here, the application backend is deployed using Bitnami's PostgreSQL Helm chart. Secrets are created in the test app namespace for secure connection to the backend.

            - To deploy the app's backend, run:
              ```bash
              ./7_app_deploy_backend.sh
              ```

        3) Deploy App

            - Finally, the demo app is ready for deployment. This step uses Helm to install the demo app with Summon and a Conjur Authenticator client sidecar.

            - To deploy the app with Summon and the authenticator sidecar, run:
              ```bash
              ./8_app_deploy.sh
              ```

            - When successful, the script will output the following:
              ```
              Test app/sidecar deployed.
              ```

8) Verify Deployment and Conjur Kubernetes Authenticator

    - Verify the PetSore app's deployment with:
      ```bash
      9_app_verify_authentication.sh
      ```

    - This script adds an entry to the PetStore app, and then queries for that entry. When successful, the script will output the following:
      ```
      [{"id": 1, "name":"Mr. Sidecar"}]
      ```
