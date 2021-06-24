# Conjur Kubernetes Authentication Helm Charts

## Overview

The Helm charts contained within this repository together aim to simplify the deployment of applications that use the Conjur Kubernetes authenticator. They do so primarily by leveraging Helm to create several Kubernetes resources allowing the following benefits to be realized:

* The amount of copy/paste boilerplate for configuration is drastically reduced; in particular, much of what is currently required to be added to CyberArk container definitions in Kubernetes manifests can be replaced by a reference to a common configuration `ConfigMap`.

* The Application or DevOps engineer who deploys each Conjur-enabled application does not need to know Conjur connection details.

* Setting up the cluster for Conjur integration fails fast, so any potential misconfigurations are caught and highlighted early. Helm's input validation contributes to this.

## Getting Started

The Helm charts in this repository can be used with either the Enterprise or OSS versions of Conjur, and should be installed in the following order:

1) `conjur-config-cluster-prep` ([README](conjur-config-cluster-prep/README.md))
2) `conjur-config-namespace-prep` ([README](conjur-config-namespace-prep/README.md))

Optionally, use the following to deploy a sample application:

3) `conjur-app-deploy` ([README](conjur-app-deploy/README.md))
