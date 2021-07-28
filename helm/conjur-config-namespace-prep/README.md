# Conjur NameSpace Preparation Helm Chart

- [Conjur NameSpace Preparation Helm Chart](#conjur-NameSpace-preparation-helm-chart)
  * [Overview](#overview)
    + [Prerequisites](#prerequisites)
    + [Objects Created](#objects-created)
    + [Conjur Enterprise Documentation Reference](#conjur-enterprise-documentation-reference)
  * [Configuration](#configuration)
  * [Examples](#examples)
    + [Fresh Installation](#fresh-installation)
    + [Upgrading](#upgrading)
    + [Alternative: Creating K8s Resources with `kubectl` instead of Helm](#alternative-creating-k8s-resources-with-kubectl-instead-of-helm)
  * [Issues with Helm "lookup"](#issues-with-helm-"lookup")
  * [Using the Conjur Connection ConfigMap in your application deployment manifest](#using-the-conjur-connection-configmap-in-your-application-deployment-manifest)

<!--
  Table of contents generated with markdown-toc
 'http://ecotrust-canada.github.io/markdown-toc/'
-->

## Overview 

The purpose of this Helm chart is to prepare a new NameSpace with credentials
needed for applications to connect to a Conjur instance, either in the same cluster,
or elsewhere.

This is done by retrieving the necessary Kubernetes and Conjur credentials, stored
in the "Golden Configmap", and making them available within the NameSpace through a
ConfigMap and RoleBinding of its own. These objects will expose the credentials as 
environment variables and prepare for communication for any Kubernetes authenticators 
in the given Conjur NameSpace. 

<img alt="Prepare Namespace" src="https://user-images.githubusercontent.com/26872683/111843074-eae22480-88d6-11eb-9cc3-60b1ece9139b.png">

### Prerequisites

- A running Conjur instance inside or outside of a Kubernetes cluster
- A NameSpace configured to contain a ["Golden Configmap"](../conjur-config-cluster-prep/README.md)

### Objects Created

The per-Kubernetes-NameSpace resources created by this Helm chart include:

- _Conjur Connection Configmap_

    The [Conjur Connection Configmap](templates/conjur-connect-configmap.yml) 
    contains references to Conjur credentials, taken from the 
    "Golden Configmap". These can be used to enable Conjur authentication for 
    applications to retrieve secrets securely.

- _Authenticator RoleBinding_

    The [Authenticator RoleBinding](templates/authenticator-RoleBinding.yml) 
    grants permissions to the Conjur Authenticator ServiceAccount for the Authn-Kubernetes ClusterRole, which provides a list of Kubernetes API access permissions. This is required to validate application identities.

### Conjur Enterprise Documentation Reference

Installation of this Helm chart replaces the manual creation of the Kubernetes resources outlined in [Steps 4 and 5 of the Conjur Enterprise Kubernetes Authenticator Documentation](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/Latest/en/Content/Integrations/k8s-ocp/cjr-k8s-authn-client.htm?tocpath=Integrations%7COpenShift%252FKubernetes%7CSet%20Up%20Applications%7C_____1#Setuptheapplicationtoretrievesecrets).

## Configuration

The following table lists the configurable parameters of the Conjur Namespace-prep-chart and their default values.

|Parameter|Description|Default|
|---------|-----------|-------|
|`authnK8s.goldenConfigMap`|Name for the "Golden Configmap" containing authn-k8s and Conjur credentials (*Required*)|`""`|
|`authnK8s.namespace:`|The NameSpace name where the "Golden Configmap" resides. (*Required*)|`""`|
|`authnRoleBinding.create`|Flag to generate the authenticator RoleBinding.|`true`|
|`authnRoleBinding.name`|Name for the RoleBinding generated if the `create` flag is set to `true`.|`"conjur-RoleBinding"`|
|`authnRoleBinding.create`|Flag to generate the ConfigMap with credentials for accessing Conjur instance.|`true`|
|`authnRoleBinding.name`|Name for the ConfigMap generated if the `create` flag is set to `true`|`"conjur-configmap"`|

## Examples

### Fresh Installation 

In this example, we will be installing the helm chart in a new NameSpace.

While you can edit [`values.yaml`](./values.yaml) to modify settings, we will be 
using the default entries in `values.yaml`.

- Create a new NameSpace

```shell-session
kubectl create NameSpace my-NameSpace
```

- Install the chart in your new NameSpace. Note that we set the values for 
  `authnK8s.goldenConfigMap` and `authnK8s.namespace` to match the name and NameSpace location for our "Golden Configmap", respectively.

```shell-session
helm install NameSpace-prep . -n "my-NameSpace" \
  --set authnK8s.goldenConfigMap="conjur-configmap" \
  --set authnK8s.namespace="default"
```

If successful, this should output the details of your chart installation,
including the new ConfigMap and RoleBinding. Your NameSpace can now utilize
the Conjur API to access your Conjur instance using the environment variables 
located in `conjur-configmap`. To view these, use the following command:

```shell-session
kubectl describe configmap -n my-NameSpace
```
### Upgrading

No special changes need to be made when upgrading. If the location of your "Golden Configmap" changes, follow the example for a ["fresh installation"](#fresh-installation), but use the `helm upgrade` command in place of `helm install`. This will generate a new Configmap and RoleBinding to reflect the changed information. 

For example:

```shell-session
helm upgrade NameSpace-prep . -n "my-NameSpace" \
  --set authnK8s.goldenConfigMap="conjur-configmap" \
  --set authnK8s.NameSpace="new-NameSpace"
```

### Alternative: Creating K8s Resources with `kubectl` instead of Helm

If Helm can not be used to deploy Kubernetes resources, the raw Kubernetes manifests can instead be generated ahead of time with the `helm template` command. The generated manifests can then be applied with `kubectl`.

```shell-session
helm template NameSpace-prep . -n "my-NameSpace" \
  --set authnK8s.goldenConfigMap="conjur-configmap" \
  --set authnK8s.NameSpace="default" > conjur-config-namespace-prep.yaml

kubectl apply -f conjur-config-namespace-prep.yaml
```

## Issues with Helm "lookup"

When using `helm install --dry-run` or `helm lint`, you may notice that the
outputted ConfigMap and RoleBinding do not contain the actual credentials
that would be retrieved from your "Golden Configmap". This is due to the
Helm `lookup` function, which retrieves Kubernetes objects and data, not being
supported by `--dry-run` or `lint`.

## Using the Conjur Connection ConfigMap in your application deployment manifest

In order to leverage the standardized `ConfigMap` containing Conjur connection details, it needs to be exposed as environment variables to the Kubernetes Authenticator Client. Edit the relevant `container` and/or `image` subsections of your application manifest to include the `ConfigMap` in the container environment, as seen in the example below:

```
containers:
- name: test-app
  envFrom:
    - configMapRef:
        name: conjur-connect
- image: cyberark/conjur-authn-k8s-client
  envFrom:
    - configMapRef:
        name: conjur-connect
```

Applications using [Secrets Provider](https://github.com/cyberark/secrets-provider-for-k8s) or [Secretless Broker](https://github.com/cyberark/secretless-broker) can be modified similarly to make use of the `ConfigMap` values.
