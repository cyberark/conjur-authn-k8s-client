# Helm Chart for Deploying Some Conjur-Enabled Applications

## Overview

This Helm Chart is composed of a number of subcharts, each describing a demo
application configured to authenticate with Conjur. Applications include:

- A demo [pet-store application](https://github.com/conjurdemos/pet-store-demo)
  with [Summon](https://github.com/cyberark/summon) installed, and either:
  - a [Conjur Kubernetes authenticator client](https://github.com/cyberark/conjur-authn-k8s-client) sidecar
  - a [Secretless Broker](https://github.com/cyberark/secretless-broker) sidecar
  - a [Secrets Provider for K8s](https://github.com/cyberark/secrets-provider-for-k8s) init container

### Prerequisites

- A running Conjur instance inside or outside of a Kubernetes cluster
- A Namespace configured with a Conjur Connection ConfigMap and an Authenticator
  RoleBinding, installed with the
  [Conjur Namespace Preparation Helm Chart](../conjur-config-namespace-prep/README.md)

## Installing Applications with the Helm Chart

To install an application, the following are required:

- enable the app's dependency condition, `<app>.enabled`
- supply a fully-qualified identifier for a Host identity permitted to
  authenticate to the `authn-k8s` endpoint

```bash
helm install application . -n "<application_namespace>" \
    --set app-summon-sidecar.enabled=true \
    --set app-summon-sidecar.conjur.authnLogin="path/to/<host_id>"
```

To install multiple applications in the same release, simply supply all required
flags for each:

```bash
helm install applications . -n "<application_namespace>" \
    --set app-summon-sidecar.enabled=true \
    --set app-summon-sidecar.conjur.authnLogin="path/to/<host_id>" \
    --set app-secretless-broker.enabled=true \
    --set app-secretless-broker.conjur.authnLogin="path/to/<host_id>"
```

## Upgrading

Upgrading the application deployment is an easy way to make changes without
uninstalling and reinstalling the deployment in question. Upgrading has the same
requirements as installing:

- enable the app's dependency condition, `<app>.enabled`
- supply a fully-qualified identifier for a Host identity permitted to
  authenticate to the `authn-k8s` endpoint

_Note: In multi-app deployments, an upgrade will uninstall
each app not explicitly enabled._

For example, to change the version of the Conjur Kubernetes authenticator
client used in the `app-summon-sidecar` subchart:

```bash
helm upgrade applications . -n "<application_namespace>" \
    --set app-summon-sidecar.enabled=true \
    --set app-summon-sidecar.conjur.authnLogin="path/to/<host_id>"
    --set app-summon-sidecar.app.image.tag="1.2.0"
```

## Configurable Values

The following tables list the configurable parameters of the Application
Deployment Helm chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.conjur.conjurConnConfigMap` | Name of the ConfigMap created by the Conjur Namespace Preparation Helm Chart | `conjur-connect` |
| `global.appServiceType` | K8s ServiceType with which to publish the Application | `NodePort` |
| `app-summon-init.enabled` | Flag to enable installation of a demo application that uses Summon and a Conjur Authenticator client init container | `false` |
| `app-summon-sidecar.enabled` | Flag to enable installation of a demo application that uses Summon and a Conjur Authenticator client sidecar | `false` |
| `app-secretless-broker.enabled` | Flag to enable installation of a demo application that uses a Secretless Broker sidecar | `false` |
| `app-secrets-provider-init.enabled` | Flag to enable installation of a demo application that uses a Secrets Provider init container | `false` |
| `app-secrets-provider-standalone.enabled` | Flag to enable installation of a demo application that uses a Secrets Provider standalone container | `false` |

### Application Subchart Configurable Values

The following values are consistent across subcharts:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `app.image.repository` | Repository and image name from which to build the test app container | `cyberark/demo-app` |
| `app.image.tag` | Image tag for test app image | `latest` |
| `app.image.pullPolicy` | Test app image pull policy | `Always` |
| `conjur.authnConfigMap.create` | Flag to enable the installation of a ConfigMap with Conjur authn details | `true` |
| `conjur.authnConfigMap.name` | Name of the ConfigMap with Conjur authn details | `conjur-authn-configmap` |
| `conjur.authnLogin` | Name of Conjur host identity with which to authenticate with Conjur | `""` |

The following values are unique to their subchart:

#### app-summon-sidecar

| Parameter | Description | Default |
|-----------|-------------|---------|
| `authnClient.image.repository` | Authenticator client image repository and name | `cyberark/conjur-authn-k8s-client` |
| `authnClient.image.tag` | Authenticator client image tag | `latest` |
| `authnClient.image.pullPolicy` | Authenticator client image pull policy | `Always` |

#### app-secretless-broker

| Parameter | Description | Default |
|-----------|-------------|---------|
| `secretless.image.repository` | Secretless Broker image repository and name | `cyberark/secretless-broker` |
| `secretless.image.tag` | Secretless Broker image tag | `latest` |
| `secretless.image.pullPolicy` | Secretless Broker image pull policy | `Always` |

#### app-secrets-provider-init

| Parameter | Description | Default |
|-----------|-------------|---------|
| `secretsProvider.image.repository` | Secrets Provider image repository and name | `cyberark/secrets-provider-for-k8s` |
| `secretsProvider.image.tag` | Secrets Provider image tag | `latest` |
| `secretsProvider.image.pullPolicy` | Secrets Provider image pull policy | `Always` |
