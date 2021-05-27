# Conjur NameSpace Preparation Helm Chart

- [Conjur NameSpace Preparation Helm Chart](#conjur-NameSpace-preparation-helm-chart)
  * [Overview](#overview)
    + [Prerequisites](#prerequisites)
    + [Objects Created](#objects-created)
  * [Configuration](#configuration)
  * [Examples](#examples)
    + [Fresh Installation](#fresh-installation)
    + [Upgrading](#upgrading)
  * [Issues with Helm "lookup"](#issues-with-helm-"lookup")

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

## Configuration

The following table lists the configurable parameters of the Conjur Namespace-prep-chart and their default values.

|Parameter|Description|Default|
|---------|-----------|-------|
|`authnK8s.goldenConfigMap`|Name for the "Golden Configmap" containing authn-k8s and Conjur credentials (*Required*)|`""`|
|`authnK8s.NameSpace:`|The NameSpace name where the "Golden Configmap" resides. (*Required*)|`""`|
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
  `authnK8s.goldenConfigMap` and `authnK8s.NameSpace` to match the name and NameSpace location for our "Golden Configmap", respectively.

```shell-session
helm install NameSpace-prep . -n "my-NameSpace" \
  --set authnK8s.goldenConfigMap="conjur-configmap" \
  --set authnK8s.NameSpace="default"
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

## Issues with Helm "lookup"

When using `helm install -dry-run` or `helm lint`, you may notice that the
outputted ConfigMap and RoleBinding do not contain the actual credentials
that would be retrieved from your "Golden Configmap". This is due to the
Helm `lookup` function, which retrieves Kubernetes objects and data, not being
supported by `--dry-run` or `lint`.
