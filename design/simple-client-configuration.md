# Table of Contents

- [Useful Links](#useful-links)
- [Background](#background)
- [Solution](#solution)
- [Design](#design)
- [Performance](#performance)
- [Backwards Compatibility](#backwards-compatibility)
- [Affected Components](#affected-components)
- [Test Plan](#test-plan)
- [Logs](#logs)
- [Documentation](#documentation)
- [Version update](#version-update)
- [Security](#security)
- [Audit](#audit)
- [Development Tasks](#development-tasks)
- [Definition of Done](#definition-of-done)
- [Solution Review](#solution-review)
- [Appendix](#appendix)

# Useful Links

<table>
<thead>
<tr class="header">
<th>Name</th>
<th>Link</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Aha Card</td>
<td><p><a href="https://cyberark.aha.io/epics/SCR-E-76">link</a> (private)</p>
<p><em>Note: This design document covers work that is planned to lay the groundwork for “Milestone 1: Push to File” as defined in this Aha Card.</em></p></td>
</tr>
<tr class="even">
<td>Feature Doc</td>
<td><a href="https://cyberark365.sharepoint.com/:w:/s/Conjur/EQnlgrc_aYZAhaZdR3oPSr0BsBoAtMnvyJuNDHGkOzokgw?e=sgrgBx">link</a> (private)</td>
</tr>
<tr class="odd">
<td>Examples of Using Kubernetes with Conjur Enterprise</td>
<td><a href="https://github.com/cyberark/dap-wiki/tree/master/how-to-guides/kubernetes">link</a> (private)</td>
</tr>
<tr class="even">
<td>Scrum Folder with work-in-progress proposals and demos</td>
<td><a href="https://github.com/conjurinc/docs/tree/master/scrum/community-and-integrations/2105%20-%20K8s%20Configuration%20Spikes">link</a> (private)</td>
</tr>
<tr class="odd">
<td>Issue (JIRA or GitHub)</td>
<td><p>Spike: There are clear instructions for programmatically retrieving the Conjur SSL certificate</p>
<p><a href="https://github.com/cyberark/conjur/issues/2038">https://github.com/cyberark/conjur/issues/2038</a></p>
<p>Spike: There is a tool to simplify the steps required to prepare a Kubernetes cluster to use the Conjur Kubernetes Authenticator</p>
<p><a href="https://github.com/cyberark/conjur/issues/2039">https://github.com/cyberark/conjur/issues/2039</a></p>
<p>Design: there is a design document for Simplified Kubernetes Client Configuration</p>
<p><a href="https://github.com/cyberark/conjur/issues/2045">https://github.com/cyberark/conjur/issues/2045</a></p></td>
</tr>
<tr class="even">
<td>Feature Brief</td>
<td><a href="https://cyberark365.sharepoint.com/:w:/r/sites/Conjur/Shared%20Documents/SDLC/Projects/Integrations/Project%20Briefs/Feature%20Brief%20-%20Easy%20Kubernetes%20Secrets%20Delivery.docx?d=wb782e509693f408685a65d477a0f4abd&amp;csf=1&amp;web=1&amp;e=48shEd">link</a> (private)</td>
</tr>
<tr class="odd">
<td>Dap Wiki Client Configuration Info</td>
<td><a href="https://github.com/cyberark/dap-wiki/blob/master/reference/client-configuration.md">link</a> (private)</td>
</tr>
</tbody>
</table>

# Background

Users deploying applications to Kubernetes or OpenShift that use our
Conjur Kubernetes authenticator currently have to provide *for each
application* detailed configuration information for the Conjur
connection, even though most of the configuration details are shared by
all applications within the cluster. Having to copy/paste so much
boilerplate is laborious, makes it easy to make mistakes, and it’s
difficult to discover misconfigurations until the very last minute when
an application is deployed.

Additionally, the current methodology forces the persona that is
deploying each application to have direct knowledge of Conjur
configuration details.

In this effort, we’d like to make some small, concrete changes to how we
manage Conjur configuration in our Kubernetes integrations so that:

-   The amount of copy/pasting of boilerplate configuration is
    drastically reduced; in particular, much of what is currently
    required to be added to CyberArk container definitions in Kubernetes
    manifests can be replaced by a reference to a common configuration
    file.

-   The persona that is deploying each Conjur-enabled application does
    not need to know Conjur connection details.

-   Setting up the cluster for Conjur integration fails fast, so any
    potential misconfigurations are caught and highlighted early. Adding
    input validation contributes to this.

# Solution

## Current Experience

In order to use any of our sidecars or init containers in Kubernetes,
*each* application must update their Kubernetes manifest to include the
following details in their sidecar/init container definition:

| Variable                 | Secrets Provider | Summon + Authn-K8s | SDK + Authn-K8s | Secretless | Default Value | Notes                                                                                                  |
|--------------------------|------------------|--------------------|-----------------|------------|---------------|--------------------------------------------------------------------------------------------------------|
| CONJUR\_AUTHN\_URL       | Yes              | Yes                | Yes             | Yes        |               |                                                                                                        |
| CONJUR\_APPLIANCE\_URL   | Yes              | Yes                | Yes             | Yes        |               |                                                                                                        |
| CONJUR\_ACCOUNT          | Yes              | Yes                | Yes             | Yes        | “default"     |                                                                                                        |
| CONJUR\_SSL\_CERTIFICATE | Yes              | Yes                | Yes             | Yes        |               | If this is not set, CONJUR\_CERT\_FILE must be set and must point to a file containing the certificate |

## Proposed Experience

### Collecting Conjur Configuration Once

Rather than collecting and configuring the above Conjur connection
information each time a Conjur-enabled application is being deployed,
this information will be collected once per Conjur instance / Conjur
Kubernetes authenticator. The collected information will be saved in
a “reference” Conjur connection ConfigMap, and a sanity test of the
ConfigMap contents will be performed.

Once the “reference” Conjur connection ConfigMap has been created and
verified, then the information contained in this ConfigMap can be used
as a reference as each Conjur-enabled application is being deployed.

### Part of a Broader Developer UX Effort

The design being proposed here is part of a broader effort to improve
the developer user experience in deploying Conjur-authenticated applications.
More specifically, this work is complementary to other developer UX
enhancements that are in progress:

- CyberArk Sidecar Injector.
  - NOTE: The design being proposed in this document aligns quite well with
    the CyberArk Sidecar Injector, as the CyberArk Sidecar Injector requires
    Conjur connection information to be configured with a ConfigMap (see the
    [CyberArk Sidecar Injector Conjur Connection ConfigMap documentation](https://github.com/cyberark/sidecar-injector/blob/master/README.md#sidecar-injectorcyberarkcomconjurconnconfig)).
- Secrets Provider Write-to-File
- Zero-Change Summon

### Proposed Workflows

The proposed design makes use of Helm charts to create the Kubernetes
objects that are required to enable Conjur Kubernetes authentication.
Sequence diagrams for proposed workflows are depicted in the “Workflow
Sequence Diagrams” section below.

In the workflow descriptions and sequence diagrams, the following
personas are assumed:

<table>
<thead>
<tr class="header">
<th>Persona</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Conjur administrator</td>
<td>A person who can provide details required for an application to authenticate with a Conjur instance (e.g. Conjur URL, Conjur account, and Conjur authenticator ID).</td>
</tr>
<tr class="even">
<td>Kubernetes administrator</td>
<td><p>A person who possesses Kubernetes RBAC privileges for:</p>
<ul>
<li><p>Accessing a Conjur Namespace (if one exists)</p></li>
<li><p>Creating and accessing authenticator Namespaces (as described in the workflows)</p></li>
<li><p>Creating and accessing Conjur-enabled application Namespaces</p></li>
</ul></td>
</tr>
<tr class="odd">
<td>DevOps engineer</td>
<td><p>A person who possesses Kubernetes RBAC privileges for:</p>
<ul>
<li><p>Accessing Conjur-enabled application Namespaces</p></li>
</ul></td>
</tr>
</tbody>
</table>

The proposed workflows can be described as follows:

-   **Preparing the Kubernetes Cluster**  
    This workflow is performed ***once per Conjur instance / per
    authn-k8s authenticator,*** typically by a ***Kubernetes
    administrator***. A new ***Kubernetes Cluster Prep Helm chart***
    will be developed to facilitate this workflow. This steps are as
    follows:

    -   Gather Conjur configuration information.

    -   Retrieve the Conjur SSL certificate.

    -   Create a Namespace for the authn-k8s authenticator. (If a Conjur
        Namespace already exists, and only one authn-k8s authenticator
        is being used in this cluster, then that Conjur Namespace can be
        used as the authenticator Namespace).

    -   Run ‘helm install’ using Kubernetes Cluster Prep Helm chart to
        create the following Kubernetes objects for this authn-k8s
        authenticator:

        -   In the authenticator Namespace:

            -   “Golden” (or Reference) ConfigMap

            -   Authenticator ServiceAccount

        -   Cluster scoped, if not already created:

            -   Authenticator ClusterRole

> If there are multiple Conjur instances and/or multiple authn-k8s
> authenticators that will be used by Conjur-enabled applications in
> each Kubernetes cluster, then this workflow would be repeated per
> Conjur instance / per authn-k8s authenticator, with the following
> guidelines:

-   Each Conjur instance / authn-k8s authenticator must have its own
    authenticator Namespace, “Golden” ConfigMap, and ServiceAccount

-   The authenticator ClusterRole can be shared. (Optionally, a
    ClusterRole can be created per Conjur instance / authn-k8s
    authenticator if each ClusterRole has a unique name).

> NOTE: Automated testing for multiple Conjur instances and/or multiple
> authn-k8s authenticators will not be implemented for the initial
> release, for reasons discussed in the “Out of Scope” section of the
> Test Plan below. Therefore, this feature is not supported for the
> initial release.

-   **Preparing the Application Namespace**  
    This workflow is typically performed ***once per application
    Namespace*** by a ***Kubernetes administrator***. A new ***Conjur
    Application Namespace Prep Helm chart*** will be developed to
    facilitate this workflow. This step includes:

    -   Running ‘helm install’ to create the following Kubernetes
        objects in the application Namespace as required for Conjur
        authentication sidecar/init containers to authenticate with
        Conjur:

        -   Conjur Connection ConfigMap

        -   Authenticator RoleBinding

-   **Deploying the Application**  
    This workflow is typically performed by a ***DevOps engineer***.  
      
    This workflow is greatly simplified due to the introduction of the
    ***Conjur Connection ConfigMap*** that is created in the application
    Namespace preparation workflow above, since application manifests no
    longer need to contain hardcoded values of Conjur connection
    environment variables.  
      
    The workflow is as follows:

    -   Add Conjur authenticator sidecar/init container definition to
        the application manifest.

    -   Add annotations to map Conjur secrets variables to application
        secrets (or add a Conjur secrets mapping ConfigMap to define
        this mapping).

    -   Deploy the modified application manifest.

### Configuration Validation: Failing Quickly

-   ***Validation of configured Helm chart values***:

    -   Each Helm chart will include a
        ‘[values.schema.json](https://helm.sh/docs/topics/charts/#schema-files)’
        file that uses [JSON schema](https://json-schema.org/) to
        represent the expected structure of configured chart values.

    -   The schema is applied to validate chart values automatically
        when Helm ‘install’, ‘upgrade’, ‘lint’, or ‘template’
        subcommands are invoked.

    -   Schema patterns can use regex patterns to impose structure on
        chart values (e.g. for URLs, SSL certificates, and Kubernetes
        object names).

    -   Tests cases are defined in the “Helm unit tests” section below.

-   ***Helm test validation of deployed Helm releases***:

    -   Helm tests are run ***on demand*** after Helm install/upgrade to
        functionally validate a deployed Helm release.

    -   Refer to sequence diagram in the “Workflow Sequence Diagrams”
        section below.

    -   Helm tests are implemented for both the Kubernetes Cluster Prep
        Helm chart and Application Namespace Prep Helm chart.

    -   The Helm test for each Helm chart is implemented as a Kubernetes
        Job that includes ‘curl’, ‘openssl’, and some test scripts.

    -   Testing includes the following:

        -   Verify Conjur URL by using insecure ‘curl’ to verify Conjur
            server is reachable.

        -   Verify the Conjur SSL certificate by retrieving the
            certificate using openssl, and verifying that the configured
            Conjur SSL certificate matches what was retrieved.

        -   Authenticate with Conjur using a ***test host*** ***that is
            pre-configured in Conjur***. This is effectively a pass/fail
            test or smoketest. For example, if the configured Conjur
            authenticator ID does not match the authenticator ID that is
            configured in Conjur policy, this test will fail. However,
            because the test pod does not have access e.g., to Conjur
            debug logs, it will not be able to identify the root cause
            of the authentication failure.

### Upgrading, Modifying Conjur Configuration, or Rotating Certificates

This design makes use of the Helm upgrade feature to support upgrading
of container images and modifications to Conjur connection information.
Changes to Conjur connection information require some level of manual
coordination between the Conjur administrator, Kubernetes administrator,
and DevOps engineers. For example, when the SSL certificate that Conjur
uses is rotated, the following operations need to be performed
sequentially:

1.  Conjur administrator rotates the Conjur SSL certificate

2.  Kubernetes administrator retrieves the new Conjur SSL certificate
    and updates the golden/reference ConfigMap with the new certificate
    using Helm upgrade.

3.  Kubernetes administrator updates the Conjur Connection ConfigMap in
    each application Namespace using Helm upgrade.

4.  DevOps engineer forces a Pod recreate for all application Pods that
    contain Conjur authentication sidecars/init containers.

The workflow for Conjur SSL certificate rotation is depicted in the
“Workflow Sequence Diagrams” section below.

Technically, it would be possible to develop automation to detect
changes in Conjur connection details and update the configuration in
each Conjur connection ConfigMap. For example, a Kubernetes Operator
could \*potentially\* be developed that makes use of a Kubernetes
Control Loop to continuously monitor the golden/reference ConfigMap for
Conjur connection changes, and when changes are detected, the Kubernetes
Operator would modify/update the Conjur Connect ConfigMaps in each of
the application Namespaces. Additionally, there may be ways to trigger
the restart of Conjur authentication containers for those containers to
automatically start to use the new Conjur connection information (e.g.,
a new Conjur SSL certificate). However, this would be a rather involved
development effort.

### How Will the Workflow Differ With the Introduction of the Sidecar Injector?

Soon, we will be providing guidance on how the [Cyberark Sidecar
Injector project](https://github.com/cyberark/sidecar-injector) can be
combined with the workflow that is proposed in this document. Using the
[Sidecar Injector](https://github.com/cyberark/sidecar-injector) will
further streamline the process of deploying Conjur-enabled applications.

The current design of the Sidecar Injector requires the deployment of a
[Conjur connection
ConfigMap](https://github.com/cyberark/sidecar-injector#sidecar-injectorcyberarkcomconjurconnconfig)
in the application Namespace (like what’s prescribed in the workflow
above). However, the Sidecar Injector design can be potentially modified
such that the Conjur connection information is “baked” into the
authenticator sidecar definitions that are being injected as literal
environment variable settings for the container. This would be done as
part of the Helm installation of the Sidecar injector.

By “baking in” Conjur connection information as literal environment
variable settings in the injected sidecar containers, the workflow could
be streamlined:

-   The Sidecar Injector Helm chart and the Kubernetes cluster prep Helm
    chart can be merged into a single, Sidecar Injector Helm chart.

-   The Golden ConfigMap would no longer be required (logically, the
    sidecar patches that are created by the Sidecar Injector would cache
    this information).

-   The Conjur Connect ConfigMaps would no longer be required in
    application Namespaces. Helm test validation of the contents of
    these ConfigMaps would no longer be necessary.

However, the development effort required for implementing the Golden
ConfigMap and the Conjur Connect Configmaps (let’s call it the
“Sidecar-injector-less" workflow), is not total “throw-away” effort in
that:

-   We’ll still want to support “Sidecar-injector-less" workflow for
    deploying the Secrets Provider as a standalone “application” Pod.

-   The Helm test based “fail quickly” validation mechanisms that we
    develop can be ported to the Sidecar Injector Helm chart and
    possibly the Secrets Provider Helm chart.

-   The application Namespace prep Helm chart would still be required,
    although it may be reduced to simply deploying a RoleBinding.

## Project Scope and Limitations

The initial implementation and testing will be limited to:

-   Authentication containers to be tested:

    -   Secrets Provider init container

    -   Secrets Provider standalone Pod

    -   Secretless Broker sidecar

    -   authn-k8s sidecar with an app that incorporates Summon

-   Platforms:

    -   Kubernetes (this will be either Kubernetes-in-Docker, or GKE).

    -   OpenShift 4.6

-   Automated testing of multiple Conjur instances and/or multiple
    authenticators will not be implemented for the initial release, for
    reasons discussed in “Out of Scope” subsection in the “Test Plan”
    section below. Manual testing of multiple Conjur instances /
    authenticators will be performed a stretch goal.

#  Design 

Workflow Sequence Diagrams

<img width="388" alt="Prepare Cluster" src="https://user-images.githubusercontent.com/26872683/111843055-e0278f80-88d6-11eb-9e1d-cb42c49e8919.png">

<img width="360" alt="Run Cluster Helm Test" src="https://user-images.githubusercontent.com/26872683/111843065-e61d7080-88d6-11eb-86a2-a37edd7c25a4.png">

<img width="405" alt="Prepare Namespace" src="https://user-images.githubusercontent.com/26872683/111843074-eae22480-88d6-11eb-9cc3-60b1ece9139b.png">

<img width="379" alt="Deploy Application" src="https://user-images.githubusercontent.com/26872683/111843083-eddd1500-88d6-11eb-93b7-7de7537ac7c1.png">

<img width="410" alt="Conjur Certificate Rotation" src="https://user-images.githubusercontent.com/26872683/111843089-f03f6f00-88d6-11eb-95ef-83d1357edc96.png">

## Data Model

### User Input: Information Gathered for Kubernetes Cluster Preparation

As described in the workflow above, the Kubernetes administrator will gather
Conjur connection information from a Conjur administrator, will use the Conjur
URL provided to retrieve a Conjur SSL certificate (verifying the certificate’s
fingerprint if that was provided by the Conjur administrator), and will then
create a custom \`values.yaml\` file for the Kubernetes Preparation Helm chart.
The following table describes the customizable values for this chart: 

<table>
<thead>
<tr class="header">
<th><strong>Description</strong></th>
<th><strong>Kubernetes Cluster Prep Helm Chart Value</strong></th>
<th><strong>Source</strong></th>
<th><strong>Mandatory</strong></th>
<th><strong>Notes</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Conjur Account</td>
<td>conjur.account</td>
<td>Conjur Admin</td>
<td>No</td>
<td>Defaults to “default”</td>
</tr>
<tr class="even">
<td>Conjur Appliance URL</td>
<td>conjur.applianceUrl</td>
<td>Conjur Admin</td>
<td>Yes</td>
<td></td>
</tr>
<tr class="odd">
<td>Conjur SSL Certificate</td>
<td>conjur.ssl.certificate</td>
<td>Conjur SSL certificate retrieval utility (new CLI)</td>
<td>Yes</td>
<td></td>
</tr>
<tr class="even">
<td>Authenticator ID</td>
<td>authK8s.authenticatorID</td>
<td>Conjur Admin</td>
<td>Yes</td>
<td></td>
</tr>
<tr class="odd">
<td>Authenticator ClusterRole</td>
<td>authnK8s.rbac.clusterRole</td>
<td>Kubernetes Admin</td>
<td>No</td>
<td>Created by chart if necessary.</td>
</tr>
<tr class="even">
<td>Conjur / Authenticator Namespace</td>
<td>authnK8s.namespace</td>
<td>Kubernetes Admin</td>
<td>Yes</td>
<td></td>
</tr>
<tr class="odd">
<td>Authenticator<br />
Service<br />
Account</td>
<td>authnK8s.serviceAccount.name</td>
<td>Kubernetes Admin</td>
<td>No</td>
<td>Created by chart if necessary.</td>
</tr>
</tbody>
</table>

### Mapping of Helm Chart Values to ConfigMap Data and Pod Environment Variables

The following table displays the mapping between the Kubernetes Cluster
Prep Helm chart configurable values, data that is contained in
ConfigMaps, and Sidecar/Init container environment variables that
reference this ConfigMap data.  
  
Some notes about the naming conventions used for keys in this table:

-   Each value for the Kubernetes cluster prep Helm chart is represented
    by a dotted path that corresponds with its hierarchical path in the
    ‘values.yaml’ file.

-   The data that is contained in the Golden ConfigMap is in the form of
    a flat list of key/value pairs. Keys in this list use the camelCase
    format typically used in ConfigMap definitions.

-   The data that is contained in the Conjur Connection ConfigMap is
    also in the form of a flat list of key/value pairs. Keys in this
    list must match the environment variable names that are required by
    the authenticator sidecar/init containers (i.e., they must be in
    all-caps SNAKE\_CASE format) so that they can be imported as
    environment variables by the authentication containers using the
    ‘envFrom’ field in the application Pod manifests. (The ‘envFrom’
    field allows the environment variable settings to be imported
    without having to explicitly enumerate the key / environment
    variable names.)

<table>
<thead>
<tr class="header">
<th><strong>Kubernetes Cluster Prep Helm Chart Value</strong></th>
<th><strong>Key in</strong><br />
<strong>Golden ConfigMap</strong></th>
<th><strong>Key in Conjur Connection ConfigMap</strong><br />
<strong>(NOTE 1)</strong></th>
<th><strong>Notes</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>authnK8s.<br />
authenticatorID</td>
<td>authnK8s<br />
AuthenticatorID</td>
<td></td>
<td>The value of the authnK8sAuthenticatorID that is stored in the Golden ConfigMap must be URL encoded to ensure that it contains only those characters that are valid within a URI.</td>
</tr>
<tr class="even">
<td>authnK8s.rbac.<br />
clusterRole</td>
<td>authnK8s<br />
ClusterRole</td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>authnK8s.<br />
namespace</td>
<td>authnK8s<br />
Namespace</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>authnK8s.<br />
serviceAccount.<br />
name</td>
<td>authnK8sService<br />
Account</td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>conjur.<br />
account</td>
<td>conjurAccount</td>
<td>CONJUR_ACCOUNT</td>
<td>Maps to Pod environment:<br />
CONJUR_ACCOUNT</td>
</tr>
<tr class="even">
<td>conjur.<br />
applianceUrl</td>
<td>conjurApplianceUrl</td>
<td>CONJUR_APPLIANCE_URL</td>
<td>Maps to Pod environment:<br />
CONJUR_APPLIANCE_URL</td>
</tr>
<tr class="odd">
<td></td>
<td></td>
<td>CONJUR_AUTHN_URL</td>
<td><p>Maps to Pod environment:<br />
CONJUR_AUTHN_URL<br />
<br />
CONJUR_AUTHN_URL is derived from fields in the Golden ConfigMap as follows:</p>
<p>&lt;conjurApplianceUrl&gt;/<br />
authn-k8s/<br />
&lt;URL encoded authnK8sAuthenticatorID&gt;</p></td>
</tr>
<tr class="even">
<td>conjur.ssl.<br />
certificate</td>
<td>conjurSslCertificate</td>
<td>CONJUR_SSL_<br />
CERTIFICATE</td>
<td>Maps to Pod environment:<br />
CONJUR_SSL_CERTIFICATE</td>
</tr>
<tr class="odd">
<td></td>
<td>ConjurSslCertificate<br />
Base64</td>
<td></td>
<td>Base64 encoded version of conjurSslCertificate. This allows the Helm lookup function to access the certificate as a single-line value.</td>
</tr>
</tbody>
</table>

**NOTE 1**: A per-application-Namespace ConfigMap is required because
Pods can only reference ConfigMaps that are in their own Namespace.  
**NOTE 2**: The keys that are used in the Conjur Connection ConfigMap
must match their corresponding environment variables in the
authenticator Pods (i.e., they are in all-caps SNAKE\_CASE) so that they
can be mapped implicitly using ‘envFrom’ in the Pod manifest.  
**NOTE 3**: The value of the authnK8sAuthenticatorID that is stored in
the Golden ConfigMap must be URL encoded to ensure that it contains only
those characters that are valid within a URI.  
**NOTE 4**: CONJUR\_AUTHN\_URL is derived from fields in the Golden
ConfigMap as follows:  
&lt;conjurApplianceUrl&gt;/authn-k8s/&lt;URL encoded
authnK8sAuthenticatorID&gt;

## Example Kubernetes Manifests

### Example Golden/Reference ConfigMap Manifest

The following is an example of a manifest that might be rendered by the
Kubernetes cluster prep Helm chart for the Golden ConfigMap. This Helm
chart will deploy the Golden ConfigMap in the authenticator Namespace,
saving Conjur configuration information that was gathered by the
Kubernetes administrator as shown in the “Preparing Kubernetes Cluster”
workflow sequence diagram above.

The data contained in this ConfigMap will be used as a reference by the
application NameSpace prep Helm chart.

There are several labels that are included in this manifest that can be
used to find or select the ConfigMap without depending on or knowing the
actual name of the ConfigMap. For example, this could be used by a
Kubernetes administrator to find all Kubernetes objects that are related
to Conjur authn-k8s. The keys used in these labels are consistent with
the [standard labels that are
recommended](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels)
by the Kubernetes community.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: authn-k8s-config-map
  labels:
    app.kubernetes.io/name: authn-k8s
    app.kubernetes.io/component: golden-configmap
    app.kubernetes.io/instance: conjur-authn-k8s
    app.kubernetes.io/part-of: cluster-config
    app.kubernetes.io/managed-by: helm
    helm.sh/chart: authn-k8s-cluster-prep-1.0.0
 data:
  # authn-k8s Configuration
  authnK8sAuthenticatorID: my-authenticator-id
  authnK8sClusterRole: conjur-authenticator-clusterrole
  authnK8sNamespace: conjur-ns
  authnK8sServiceAccount: conjur-sa

  # Conjur Configuration
  conjurAccount: myConjurAccount
  conjurApplianceUrl: https://conjur.conjur-ns.svc.cluster.local
  conjurSslCertificate: <Unencoded Conjur SSL certificate>
  conjurSslCertificateBase64: <base64 encoded Conjur SSL certificate> 

```

### Example Conjur Connection ConfigMap Manifest 
  
The following is an example of a manifest that might be rendered by the
Kubernetes cluster prep Helm chart for the Conjur Connection ConfigMap.
This Helm chart will deploy the Conjur Connection ConfigMap in the
application Namespace, as shown in the “Preparing Application Namespace”
workflow sequence diagram. This ConfigMap will be referenced by the
authenticator sidecar/init containers in application Pods.

There are several labels that are included in this manifest that can be
used to find or select the ConfigMap without depending on or knowing the
actual name of the ConfigMap. For example, this could be used by a
Kubernetes administrator to find all Kubernetes objects that are related
to Conjur authn-k8s in all application Namespaces. The keys used in
these labels are consistent with the [standard labels that are
recommended](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels)
by the Kubernetes community.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: conjur-connect-configmap
  labels:
    app.kubernetes.io/name: authn-k8s
    app.kubernetes.io/component: conjur-conn-configmap
    app.kubernetes.io/instance: pet-store-authn-k8s
    app.kubernetes.io/part-of: app-namespace-config
    app.kubernetes.io/managed-by: helm
    helm.sh/chart: authn-k8s-namespace-prep-1.0.0
data:
  CONJUR_ACCOUNT: myConjurAccount 
  CONJUR_APPLIANCE_URL: https://conjur.conjur-ns.svc.cluster.local
  CONJUR_AUTHN_URL: https://conjur.conjur-ns.svc.cluster.local/authn-k8s/my-authenticator-id
  CONJUR_SSL_CERTIFICATE: <unencoded Conjur SSL certificate>
```

### Example Secrets Provider Pod Manifest

The following is an example of an application Pod manifest that might be
defined by a DevOps engineer that includes a Secrets Provider init
container.

Of particular interest in this Pod manifest is the use of an ‘envFrom’
section (highlighted in yellow) in the Secrets Provider init container
definition. The ‘envFrom’ section allows the Secrets Provider init
container to import ALL key/value definitions in the Conjur Connection
ConfigMap as environment variable settings, without having to explicitly
enumerate each key / environment variable name.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: pet-store
    app.kubernetes.io/component: server
    app.kubernetes.io/instance: my-pet-store
    app.kubernetes.io/part-of: web-app
    app.kubernetes.io/managed-by: helm
    helm.sh/chart: pet-store-chart-2.0.1
    app: test-env
  name: test-env
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-env
  template:
    metadata:
      labels:
        app: test-env
    spec:
      serviceAccountName: app-f9d91ac1-8-sa
      containers:
      - image: debian
        name: init-env-app
        command: ["sleep"]
        args: ["infinity"]
        env:
          - name: TEST_SECRET
            valueFrom:
              secretKeyRef:
                name: test-k8s-secret
                key: secret
          - name: VARIABLE_WITH_SPACES_SECRET
            valueFrom:
              secretKeyRef:
                name: test-k8s-secret
                key: var_with_spaces
          - name: VARIABLE_WITH_PLUSES_SECRET
            valueFrom:
              secretKeyRef:
                name: test-k8s-secret
                key: var_with_pluses
          - name: NON_CONJUR_SECRET
            valueFrom:
              secretKeyRef:
                name: test-k8s-secret
                key: non-conjur-key
      initContainers:
      - image: 'secrets-provider-for-k8s:latest'
        imagePullPolicy: Never
        name: cyberark-secrets-provider-for-k8s
        env:
          - name: MY_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
           - name: MY_POD_NAMESPACE
             valueFrom:
               fieldRef:
                 apiVersion: v1
                 fieldPath: metadata.namespace
           - name: CONTAINER_MODE
             value: init
           - name: K8S_SECRETS
             value: test-k8s-secret
           - name: SECRETS_DESTINATION
             value: k8s_secrets
           - name: DEBUG
             value: "true"
           - name: CONJUR_AUTHN_LOGIN
             valueFrom:
              configMapKeyRef:
                name: conjur-app-configmap
                key: conjurAuthLogin
        envFrom:
        - configMapRef:
            name: conjur-connect-configmap
         imagePullSecrets:
        - name: dockerpullsecret
```

# Performance

## Kubernetes Cluster Preparation Performance

The target performance metrics for performing Kubernetes cluster
preparation and cluster upgrade are TBD. The real time latency involved
in performing this task, and the system processing time for both the
local Helm client node and for the Kubernetes platform will vary widely
based on the type of platforms being used (e.g., GKE vs. OpenShift), as
well as system load on the platforms.

Given that this process is done once per Conjur instance for initial
setup, and then only for upgrades or for Conjur configuration
modifications, it’s not expected that this will be a critical
performance metric.  
  
## Application Namespace Preparation Performance

The target performance metrics for performing application Namespace
preparation and upgrade/update are TBD. As with the Kubernetes cluster
preparation process, the real time latency involved in performing this
task, and the system processing time for both the local Helm client node
and for the Kubernetes platform will vary widely based on the type of
platforms being used (e.g., GKE vs. OpenShift), as well as system load
on the platforms.

Again, it’s not expected that this will be a critical performance
metric.  
  
## Application Deployment Performance

It is not expected that the workflow changes being proposed will have a
significant impact on the real time and system time that it takes to
deploy applications.

However, it should be noted that the deployment times will be affected
with the upcoming introduction of the CyberArk Sidecar Injector, since
the mutation of application Pods will add latency to application
deployments.

## Conjur Secrets Access Performance

It is not expected that the workflow changes being proposed will have a
significant impact on the real time and system time that it takes to
access application-specific Conjur secrets, once an authenticator
sidecar/init container has been injected and configured.

# Backwards Compatibility

There may be existing application deployments that are using Conjur
Kubernetes authentication for which Kubernetes administrators or DevOps
engineers may want to migrate / update to the workflow described in this
document. Such a migration / update to the new workflow would allow for
more streamlined handling of changes or updates of Conjur connection
details (e.g., Conjur SSL certificate rotation).

Migration of an existing Conjur

1.  **DevOps Engineer**: Modify application Kubernetes manifests such
    that the authenticator sidecar/init container uses a Conjur
    connection ConfigMap to access Conjur connection environment
    variables, rather than using hardcoded values for these environment
    variables in the manifest.

2.  **Kubernetes Admin**: Create a custom values.yaml file for the
    Kubernetes cluster prep Helm chart, and run Helm install to create a
    reference/golden ConfigMap, authenticator ServiceAccount, and
    authenticator ClusterRole.

3.  **Kubernetes Admin**: Run the application Namespace prep Helm chart
    to create a Conjur connection ConfigMap and an authenticator
    RoleBinding.

4.  **DevOps Engineer**: Re-deploy the application pods, e.g., by
    deleting the existing Pods or running ‘helm upgrade …' if they were
    Helm deployed.

# Affected Components

The Helm charts and automated testing that is being proposed in this
design will be implemented in the ‘cyberark/conjur-authn-k8s-client'
repository. Implementing the design in this repository will provide a
centralized location to which all Conjur authenticator clients (Secrets
Provider, Secretless Broker, authn-k8s) can refer.

Adding both the Kubernetes Cluster Prep Helm Chart and the Application
Namespace Prep Helm Chart to the ‘conjur-authn-k8s-client' repository
will also facilitate the implementation of E2E automated tests that
incorporate both Helm charts.

Aside from the ‘cyberark/conjur-authn-k8s-client' repository, the design
can be considered mostly additive, and the existing authenticator
clients should not require changes.

However, it may be worth considering upgrading some repositories to make
use of the methodology proposed in this design:

-   Secrets Provider: Modify existing testing of the Secrets Provider
    init container to make use of Conjur Connection ConfigMaps.

-   ‘cyberark/kubernetes-conjur-deploy' Repository: Modify deployments
    to include Kubernetes cluster preparation using the newly developed
    Helm chart.

-   ‘conjurdemos/kubernetes-conjur-demo' Repository: Modify demo scripts
    to make use of the newly developed Application Namespace Prep Helm
    chart.

# Test Plan

## Test environments

<table>
<thead>
<tr class="header">
<th><strong>Feature</strong></th>
<th><strong>Platform</strong></th>
<th><strong>Version Number</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Deploy on Kubernetes</td>
<td><p>Either:</p>
<ul>
<li><p>Kubernetes-in-Docker (KinD),<br />
….or.....</p></li>
<li><p>Google Kubernetes Engine (GKE)</p></li>
</ul></td>
<td>KinD 0.10.0<br />
… or ...<br />
GKE 1.18.16</td>
</tr>
<tr class="even">
<td>Deploy on OpenShift</td>
<td>OpenShift</td>
<td>4.6</td>
</tr>
</tbody>
</table>

## Test assumptions

-   All test cases listed below are assumed to be automated, unless
    noted otherwise.

-   Helm functional tests assume that there is a special, test host ID
    configured in Conjur policy that can be use by a test Pod to
    authenticate with Conjur.

## Out of scope

-   Initial testing will not include automated testing for multiple
    conjur instances and/or multiple authn-k8s authenticators. This
    functionality may be performed manually for the first release, time
    permitting.  
      
    The development cost of providing automated testing for multiple
    conjur instances / authenticators would be that we’d need to
    develop:

    -   Scripts that run the tests proposed in the "Helm release
        validation tests” for multiple, parallel sets of authn-k8s
        authenticator, authenticator Namespace, application Namespaces,
        applications, etc.

    -   Scripts that run the workflow defined in the “Integration / E2E
        tests” for multiple, parallel sets of authn-k8s authenticator,
        authenticator Namespace, application Namespaces, applications,
        etc.

## Test prerequisites

-   Kubernetes platform is available with \`kubectl\` and \`helm\` local
    test client access.

-   Conjur instance is available. Testing should support the following
    three configurations:

    -   Conjur Open Source in the Kubernetes cluster

    -   Conjur Enterprise master outside of Kubernetes cluster,
        followers inside the Kubernetes cluster

    -   Conjur Enterprise master and followers outside of Kubernetes
        cluster

## Test cases 

### Helm lint tests for Kubernetes cluster prep Helm chart

Running ‘helm lint’ on the Kubernetes cluster prep Helm chart will validate
the structure of templates contained in the Helm chart. This can be tested
either in a Jenkins pipeline or with the ‘[helm-check](https://github.com/marketplace/actions/github-action-for-helm-templates-validation)’
GitHub action.

### Helm lint tests for application Namespace prep Helm chart

Running ‘helm lint’ on the application Namespace prep Helm chart will
validate the structure of templates contained in the Helm chart. This can
be tested either in a Jenkins pipeline or with the ‘[helm-check](https://github.com/marketplace/actions/github-action-for-helm-templates-validation)’
GitHub action.

### Unit tests for Kubernetes cluster prep Helm chart

These unit tests make use of the Helm chart [unittest project](https://github.com/lrills/helm-unittest#helm-unittest).

“Negative” test cases are included to verify that the JSON schema validation
for this chart fails as expected when configured chart values (e.g., For
Conjur URL, Conjur SSL certificate, ServiceAccount name, etc.) are invalid. 

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>Scenario</strong></th>
<th><strong>Status</strong><br />
<strong>(Done / Not Done)</strong></th>
<th><strong>Notes</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td>Default value for Conjur account used when not explicitly configured.</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>2</td>
<td><p>Happy path: With valid chart values, templates for the following objects are created with the expected names:</p>
<ul>
<li><p>Golden ConfigMap</p></li>
<li><p>Authenticator ServiceAccount</p></li>
<li><p>Authenticator ClusterRole</p></li>
</ul></td>
<td></td>
<td>NOTE 1, NOTE 2, NOTE 3, NOTE 4, NOTE 5, NOTE 6</td>
</tr>
<tr class="odd">
<td>3</td>
<td>With invalid Conjur URL, template rendering fails</td>
<td></td>
<td>NOTE 7, NOTE 8</td>
</tr>
<tr class="even">
<td>4</td>
<td>With invalid Conjur SSL certificate, template rendering fails</td>
<td></td>
<td>NOTE 9</td>
</tr>
<tr class="odd">
<td>5</td>
<td>With invalid authenticator ServiceAccount name, template rendering fails</td>
<td></td>
<td>NOTE 10</td>
</tr>
<tr class="even">
<td>6</td>
<td>With invalid authenticator Namespace name, template rendering fails</td>
<td></td>
<td>NOTE 11</td>
</tr>
</tbody>
</table>

<table>
<thead>
<tr class="header">
<th>Note</th>
<th>Text</th>
</tr>
</thead>
<tbody>
<tr class="even">
<td><strong>NOTE 1:</strong></td>
<td><p>A valid Conjur URL:</p>
<ul>
<li><p>Begins with ‘https://’ or ‘HTTPS://’</p></li>
<li><p>Followed by a valid DNS domain, or ‘&lt;IP address&gt;:&lt;port&gt;’</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><strong>NOTE 2:</strong></td>
<td><p>A valid SSL certificate must:</p>
<ul>
<li><p>Begin with ‘-----BEGIN CERTIFICATE-----’</p></li>
<li><p>End with ‘-----END CERTIFICATE-----’</p></li>
<li><p>Characters in between are base64</p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 3:</strong></td>
<td><p>A valid Kubernetes object name (e.g. ServiceAccount name, Namespace name, etc.) must:</p>
<ul>
<li><p>Contain no more than 253 characters</p></li>
<li><p>Contain only lowercase alphanumeric characters, '-' or '.'</p></li>
<li><p>Start and end with an alphanumeric character</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><strong>NOTE 4:</strong></td>
<td><p>Example valid URLs to test:</p>
<ul>
<li><p><a href="https://conjur.example.com">https://conjur.example.com</a></p></li>
<li><p>HTTPS://conjur.example.com</p></li>
<li><p><a href="https://conjur.conjur-namespace.svc.cluster.local">https://conjur.conjur-namespace.svc.cluster.local</a></p></li>
<li><p><a href="https://192.0.2.1:443">https://192.0.2.1:443</a></p></li>
<li></li>
<li><p><a href="https://conjur.example.com/foo-bar">https://conjur.example.com/foo-bar</a></p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 5:</strong></td>
<td><p>Example valid SSL certificates to test:</p>
<ul>
<li><p>Valid Self-signed certificate</p></li>
<li><p>Valid CA-signed certificate</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><strong>NOTE 6:</strong></td>
<td><p>Example valid ServiceAccount and Namespace names to test:</p>
<ul>
<li><p>my.service.account</p></li>
<li><p>my-namespace-name</p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 7:</strong></td>
<td>With Helm ‘unittest’, a failure to render templates can be detected by asserting on the ‘hasDocuments’ count value being set to 0.</td>
</tr>
<tr class="odd">
<td><strong>NOTE 8:</strong></td>
<td><p>Example invalid URLS to test:</p>
<ul>
<li><p><a href="http://conjur.example.com">http://conjur.example.com</a></p></li>
<li><p>https://conjur_example.com</p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 9:</strong></td>
<td><p>Example invalid SSL certificates to test:</p>
<ul>
<li><p>Base64 encoded certificate</p></li>
<li><p>Certificate that doesn’t contain “-----BEGIN CERTIFICATE-----”</p></li>
<li><p>Certificate that doesn’t contain “-----END CERTIFICATE-----”</p></li>
<li><p>Certificate that contains illegal ‘%’ character</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><strong>NOTE 10:</strong></td>
<td><p>Sample invalid ServiceAccount names to test:</p>
<ul>
<li><p>ServiceAccountWithUpperCase</p></li>
<li><p>service-account-ending-with-dash-</p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 11:</strong></td>
<td><p>Sample invalid Namespace names to test:</p>
<ul>
<li><p>&lt;namespace name that is longer than 253 characters&gt;</p></li>
<li><p>namespace_name_with_underscores</p></li>
</ul></td>
</tr>
</tbody>
</table>

### Unit tests for application Namespace prep Helm chart

These unit tests will make use of the Helm chart [unittest project](https://github.com/lrills/helm-unittest#helm-unittest).

“Negative” test cases are included to verify that the JSON schema
validation for this chart fails as expected when configured chart
values (e.g. for authenticator Namespace, Conjur connection ConfigMap
name, etc.) are invalid. 

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>Scenario</strong></th>
<th><strong>Status</strong><br />
<strong>(Done / Not Done)</strong></th>
<th><strong>Notes</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td><p>Happy path: With valid chart values configured, templates for the following objects are created with the expected names:</p>
<ul>
<li><p>Conjur Connect ConfigMap</p></li>
<li><p>Authenticator RoleBinding</p></li>
</ul></td>
<td></td>
<td>NOTE 1</td>
</tr>
<tr class="even">
<td>2</td>
<td>When required values `authnK8s.namespace` and `authnK8s.goldenConfigMap` are not configured, template rendering fails</td>
<td></td>
<td>NOTE 2</td>
</tr>
<tr class="odd">
<td>3</td>
<td>With invalid authenticator Namespace name, template rendering fails</td>
<td></td>
<td>NOTE 2, NOTE 3</td>
</tr>
<tr class="even">
<td>4</td>
<td>With invalid Conjur Connection ConfigMap name configured, template rendering fails</td>
<td></td>
<td>NOTE 2, NOTE 4</td>
</tr>
<tr class="odd">
<td>5</td>
<td>With invalid authenticator RoleBinding name configured, template rendering fails</td>
<td></td>
<td>NOTE 2, NOTE 5</td>
</tr>
</tbody>
</table>

<table>
<thead>
<tr class="header">
<th><strong>NOTE 1:</strong></th>
<th><p>Valid Namespace, ConfigMap, and RoleBinding names to test:</p>
<ul>
<li><p>namespace.name.with.dots</p></li>
<li><p>configmap-name-with-dashes</p></li>
<li><p>rolebinding-name-ends-with-0</p></li>
</ul></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><strong>NOTE 2:</strong></td>
<td>With Helm ‘unittest’, a failure to render templates can be detected by asserting on the ‘hasDocuments’ count value being set to 0.</td>
</tr>
<tr class="even">
<td><strong>NOTE 3:</strong></td>
<td><p>Sample invalid Namespace names to test:</p>
<ul>
<li><p>NameWithUpperCase</p></li>
<li><p>name-ending-with-dash-</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><strong>NOTE 4:</strong></td>
<td><p>Sample invalid ConfigMap name to test:</p>
<ul>
<li><p>&lt;ConfigMap name that is longer than 253 characters&gt;</p></li>
<li><p>namespace_name_with_underscores</p></li>
</ul></td>
</tr>
<tr class="even">
<td><strong>NOTE 5:</strong></td>
<td><p>Sample invalid RoleBinding name to test:</p>
<ul>
<li><p>rolebinding_name_with_underscores</p></li>
</ul></td>
</tr>
</tbody>
</table>

### Helm release validation tests

Helm release validation tests will make use of the [Helm Test
feature](https://helm.sh/docs/helm/helm_test/) to validate deployed
instances (releases) of Helm charts.

Helm release validation tests are available for both the Kubernetes
cluster prep Helm chart and the application Namespace prep Helm chart.

The list of automated test cases below include both positive test cases
(for which Helm test success is expected) and negative test cases (for
which Helm test failure is expected, e.g. for incorrect configuration).
In essence, we are “***validating the validator***” for our Helm release
validation tests.

The overall workflow for this automated testing is:

-   Create a Kubernetes cluster for testing. (This can be done using
    Kubernetes-in-Docker, a.k.a. KinD, in a GitHub action).

-   Create a Conjur instance. (This can be done using Conjur Open Source Helm
    chart).

-   Generate and load Conjur policy for the authn-k8s authenticator.

-   Run ‘git clone …' to get a local copy of the Kubernetes cluster prep
    Helm chart.

-   Generate and load application-specific Conjur policy.

-   Run ‘git clone …' to get a local copy of the application Namespace
    prep Helm chart.

-   For each test case listed below:

    -   Create an appropriate, custom \`values.yaml\` file for the
        Kubernetes cluster prep Helm chart based on the test case
        scenario.

    -   Run \`helm install …\` for the Kubernetes cluster prep Helm
        chart using the modified \`values.yaml\`.

    -   Run \`helm install …\` for the application Namespace prep Helm
        chart.

    -   Run \`helm test\` for the target Helm chart (second column in
        the chart below).

    -   Verify \`helm test\` results are as expected for this test case.

    -   Run ‘helm delete …' for Kubernetes cluster prep Helm chart.

    -   Run ‘helm delete …' for the application Namespace prep Helm
        chart.

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>Helm Chart on Which to Run Helm Test</strong></th>
<th><strong>Scenario</strong></th>
<th><strong>Status (Done / Not Done)</strong></th>
<th><strong>Comments</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td>Kubernetes<br />
Cluster Prep</td>
<td>Happy path: With correct chart values, Helm test passes (test Pod can authenticate with Conjur)</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>2</td>
<td>Kubernetes<br />
Cluster Prep</td>
<td>Incorrect Conjur URL: Helm test fails with indication that it cannot curl Conjur</td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>3</td>
<td>Kubernetes<br />
Cluster Prep</td>
<td>Incorrect Conjur SSL certificate: Helm test fails with indication that configured Conjur SSL cert does not match Conjur’s actual SSL cert.</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>4</td>
<td>Kubernetes<br />
Cluster Prep</td>
<td>Incorrect Authenticator ID: Helm test fails with indication that the test Pod cannot authenticate with Conjur</td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>5</td>
<td>Application Namespace<br />
Prep</td>
<td>Happy path: With correct chart values for cluster prep, Helm test passes (test Pod can authenticate with Conjur)</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>6</td>
<td>Application Namespace<br />
Prep</td>
<td>Incorrect Conjur URL: Helm test fails with indication that it cannot curl Conjur</td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>7</td>
<td>Application Namespace<br />
Prep</td>
<td>Incorrect Conjur SSL certificate: Helm test fails with indication that configured Conjur SSL cert does not match Conjur’s actual SSL cert.</td>
<td></td>
<td></td>
</tr>
<tr class="even">
<td>8</td>
<td>Application Namespace<br />
Prep</td>
<td>Incorrect Authenticator ID: Helm test fails with indication that the test Pod cannot authenticate with Conjur</td>
<td></td>
<td></td>
</tr>
</tbody>
</table>

### Integration / E2E tests

The automated integration / E2E tests for this feature will test “happy
path” functionality across the various Conjur topologies, platforms, and
across several types of authenticator containers.

The overall workflow for this testing is:

-   For each Conjur topology / platform combination:

    -   Generate and load Conjur policy for the authn-k8s authenticator.

    -   Run ‘git clone …' to get a local copy of the Kubernetes cluster
        prep Helm chart.

    -   Generate and load application-specific Conjur policy.

    -   Run ‘git clone …' to get a local copy of the application
        Namespace prep Helm chart.

    -   Create an appropriate, custom \`values.yaml\` file for the
        Kubernetes cluster prep Helm chart based on the test case
        scenario.

    -   Run \`helm install …\` for the Kubernetes cluster prep Helm
        chart using the modified \`values.yaml\`.

    -   Run \`helm test\` for the Kubernetes cluster prep Helm chart

    -   Run \`helm install …\` for the application Namespace prep Helm
        chart.

    -   Run \`helm test\` for the application Namespace prep Helm chart.

    -   For each authenticator container type:

        -   Generate and load an application-specific Conjur policy for
            that authenticator

        -   Deploy a secrets mapping ConfigMap for the authenticator, if
            appropriate

        -   Deploy a sample application Pod that contains the
            authenticator sidecar/init container

        -   Verify that secrets can be accessed by the sample
            application

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>Platform</strong></th>
<th><strong>Conjur Topology</strong></th>
<th><strong>Scenario</strong></th>
<th><strong>Status</strong><br />
<strong>(Done /</strong><br />
<strong>Not Done)</strong></th>
<th><strong>Authenticator Containers Tested</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td>Kubernetes</td>
<td>Conjur Open Source</td>
<td>Application with Secrets Provider several types of authenticator containers can authenticate and retrieve secrets from Conjur.</td>
<td></td>
<td><ul>
<li><p>Secrets Provider<br />
init container</p></li>
<li><p>Secrets Provider<br />
application Pod</p></li>
<li><p>Secretless Broker</p></li>
<li><p>Authn-k8s sidecar</p></li>
</ul></td>
</tr>
<tr class="even">
<td>2</td>
<td>Kubernetes</td>
<td>Conjur<br />
Enterprise<br />
with Followers in Kubernetes<br />
cluster</td>
<td>Application with Secrets Provider several types of authenticator containers can authenticate and retrieve secrets from Conjur.</td>
<td></td>
<td><ul>
<li><p>Secrets Provider<br />
init container</p></li>
<li><p>Secrets Provider<br />
sidecar container</p></li>
<li><p>Secretless Broker</p></li>
<li><p>Authn-k8s sidecar</p></li>
</ul></td>
</tr>
<tr class="odd">
<td>3</td>
<td>Kubernetes</td>
<td>Conjur<br />
Enterprise<br />
with Followers outside of Kubernetes<br />
cluster</td>
<td>Application with Secrets Provider several types of authenticator containers can authenticate and retrieve secrets from Conjur.</td>
<td></td>
<td><ul>
<li><p>Secrets Provider<br />
init container</p></li>
<li><p>Secrets Provider<br />
sidecar container</p></li>
<li><p>Secretless Broker</p></li>
<li><p>Authn-k8s sidecar</p></li>
</ul></td>
</tr>
<tr class="even">
<td>4</td>
<td>OpenShift<br />
4.6</td>
<td>Conjur<br />
Enterprise<br />
with Followers outside of Kubernetes<br />
cluster</td>
<td>Application with Secrets Provider several types of authenticator containers can authenticate and retrieve secrets from Conjur.</td>
<td></td>
<td><ul>
<li><p>Secrets Provider<br />
init container</p></li>
<li><p>Secrets Provider<br />
sidecar container</p></li>
<li><p>Secretless Broker</p></li>
<li><p>Authn-k8s sidecar</p></li>
</ul></td>
</tr>
</tbody>
</table>

### Security testing

This design does not introduce any new container images. It is assumed
that container image scans, e.g. for existing authenticator container
images, will already be performed in other repositories.

###  Performance testing

Performance benchmarks/targets for this feature are TBD, so for now,
performance testing is not included.

# Logs

This feature does not require the addition or modification of logs. The
following existing log mechanisms may be helpful in troubleshooting Helm
install/upgrade operations:

-   **Helm logs**  
    The level of output verbosity for Helm install or upgrade operations
    can be configured with two command-line flags:

    -   --debug  
        This command-line flag will enable verbose output for Helm
        client activity.

    -   -v &lt;level&gt;  
        This command-line flag sets the verbosity level for ‘kubectl’
        commands that are being executed internally by the Helm client.
        The available levels are defined in [this
        document](https://kubernetes.io/docs/reference/kubectl/cheatsheet/#kubectl-output-verbosity-and-debugging)
        as:

| **Verbosity** | **Description**                                                                                                                                                                                   |
|---------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| -v 0          | Generally useful for this to *always* be visible to a cluster operator.                                                                                                                           |
| -v 1          | A reasonable default log level if you don't want verbosity.                                                                                                                                       |
| -v 2          | Useful steady state information about the service and important log messages that may correlate to significant changes in the system. This is the recommended default log level for most systems. |
| -v 3          | Extended information about changes.                                                                                                                                                               |
| -v 4          | Debug level verbosity.                                                                                                                                                                            |
| -v 5          | Trace level verbosity.                                                                                                                                                                            |
| -v 6          | Display requested resources.                                                                                                                                                                      |
| -v 7          | Display HTTP request headers.                                                                                                                                                                     |
| -v 8          | Display HTTP request contents.                                                                                                                                                                    |
| -v 9          | Display HTTP request contents without truncation of contents.                                                                                                                                     |

-   **Kubernetes pod logs**  
    Viewable via ‘kubectl logs –n &lt;namespace&gt; &lt;pod&gt; \[-c
    &lt;container&gt;\]’.

-   **Kubernetes event logs**  
    Kubernetes events for a given Namespace are viewable via ‘kubectl
    get events -n &lt;namespace&gt;’. By default, events “age out” and
    are no longer available after one hour.

# Documentation

The deployment workflow that is described in this document needs to be
documented. This may include:

-   A Conjur Kubernetes Authenticator Quick-Start guide in the Conjur
    Kubernetes authenticator client repository.

-   A Katacoda scenario providing a tutorial showing how applications
    can be deployed using Conjur Kubernetes authentication and the
    workflows described in this document.

-   Each new Helm chart (Kubernetes cluster prep Helm chart and the
    application Namespace prep Helm chart) will require a descriptive
    README.md file to explain how to install and upgrade releases using
    the Helm charts, and what values are customizable for the Helm
    charts.

# Version update

The following repositories may require version updates as part of a
release:

-   The cyberark/conjur-authn-k8s-client repository will require
    extensive additions, e.g.:

    -   Kubernetes Cluster Prep Helm Chart

    -   Application Namespace Prep Helm Chart

    -   Unit tests, Helm test, and E2E tests for the proposed workflow.

> Given that this is such an extensive change, a new major version
> release should be created for this change.

# Security

It is not expected that the deployment workflow that is described in
this document will introduce any security concerns. This is based on the
following assumptions:

-   The data contained in the ConfigMaps that are being introduced
    (Reference/Golden ConfigMap and the Conjur Connection ConfigMaps)
    are assumed to be non-sensitive.

-   ***Namespace-scoped*** RoleBindings are being used to grant the
    Conjur authenticator “least privilege” access to Kubernetes objects;
    that is, access is granted only for those application Namespaces
    that require Conjur authentication.

-   It is assumed that access to the Authenticator Namespace is limited
    to privileged Kubernetes administrators. Of concern here is access
    to the Authenticator ServiceAccount, which could potentially be used
    to access Kubernetes objects in all Conjur-enabled application
    Namespaces if the ServiceAccount (and the ServiceAccount token)
    became compromised and was available to non-privileged personas.

# Audit

It is not expected that the workflow proposed would require any
additional auditing support. The following existing auditing should be
sufficient for most purposes:

-   [Kubernetes
    auditing](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/)

-   [Conjur
    auditing](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/11.7/en/Content/Operations/Services/Audit/dap-overview-audit-service.htm?TocPath=Fundamentals%7C_____6)

#  Development Tasks

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>Description</strong></th>
<th><strong>Add to</strong><br />
<strong>Project /</strong><br />
<strong>Repository</strong></th>
<th><strong>Subtasks</strong></th>
<th><strong>Status</strong><br />
<strong>(Done / Not Done)</strong></th>
<th><strong>Comment</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td>Create Cluster Prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Create templates and values.yaml</p></li>
<li><p>Create schema validation</p></li>
<li><p>Documentation: README.txt and NOTES.txt</p></li>
<li><p>Create Helm unittest</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/227">cyberark/conjur-authn-k8s-client#227</a></td>
</tr>
<tr class="even">
<td>2</td>
<td>Create Cluster Prep Helm test</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Define Conjur policy for special test host ID</p></li>
<li><p>Create test scripts for test Job</p></li>
<li><p>Create container image for test Job</p></li>
<li><p>Create test manifest</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/228">cyberark/conjur-authn-k8s-client#228</a></td>
</tr>
<tr class="odd">
<td>3</td>
<td>Create Namespace Prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Create templates and values.yaml</p></li>
<li><p>Create schema validation</p></li>
<li><p>Documentation: README.txt and NOTES.txt</p></li>
<li><p>Create Helm unittest</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/236">cyberark/conjur-authn-k8s-client#236</a></td>
</tr>
<tr class="even">
<td>4</td>
<td>Create Namespace Prep Helm test</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Port from Cluster Prep Helm chart test</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/237">cyberark/conjur-authn-k8s-client#237</a></td>
</tr>
<tr class="odd">
<td>5</td>
<td>Create Helm chart (or scripts) to deploy a sample application with a selectable authenticator sidecar/init container</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Create templates and values.yaml</p></li>
<li><p>Include Pod manifests for application with each of the following:<br />
<strong>- Secrets Provider</strong><br />
<strong>init container</strong><br />
<strong>- Secrets Provider</strong><br />
<strong>application Pod</strong><br />
<strong>- Secretless Broker</strong><br />
<strong>- authn-k8s sidecar</strong></p></li>
<li><p>Include secrets mapping ConfigMaps</p></li>
</ul></td>
<td></td>
<td><ul>
<li><p>This can use the conjurdemos/kubernetes-conjur-demo scripts and the Secrets Provider automated tests as a starting point.</p></li>
</ul>
<p><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/238">cyberark/conjur-authn-k8s-client#238</a></p></td>
</tr>
<tr class="even">
<td>6</td>
<td>Create scripts for development environment and CI testing</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Create script to generate and load authn-k8s-specific Conjur policy</p></li>
<li><p>Create script to generate and load application-specific Conjur policy</p></li>
<li><p>Create scripts to do Helm install for cluster prep and app Namespace prep</p></li>
<li><p>Create script to deploy the Helm chart from Step 5 to deploy example application with a selectable authenticator type</p></li>
<li><p>Create script to verify secrets access for example application</p></li>
</ul></td>
<td></td>
<td><ul>
<li><p>This can use conjurdemos/<br />
kubernetes-conjur-demo scripts as a starting point (use Helm install instead of bash, install a single application + authenticator container at a time).</p></li>
<li><p>The Helm chart to deploy sample apps can be useful for a Quick-Start guide.</p></li>
</ul>
<p><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/239">cyberark/conjur-authn-k8s-client#239</a></p></td>
</tr>
<tr class="odd">
<td>7</td>
<td>CI to run ‘helm lint’ on cluster prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td></td>
<td></td>
<td><p>This can be tested either in a Jenkins pipeline or with the ‘<a href="https://github.com/marketplace/actions/github-action-for-helm-templates-validation">helm-check</a>’ GitHub action.</p>
<p><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/240">cyberark/conjur-authn-k8s-client#240</a></p></td>
</tr>
<tr class="even">
<td>8</td>
<td>CI to run ‘helm lint’ on namespace prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td></td>
<td></td>
<td><p>This can be tested either in a Jenkins pipeline or with the ‘<a href="https://github.com/marketplace/actions/github-action-for-helm-templates-validation">helm-check</a>’ GitHub action.</p>
<p><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/240">cyberark/conjur-authn-k8s-client#240</a></p></td>
</tr>
<tr class="odd">
<td>8</td>
<td>CI to run unit tests for the cluster prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/234">cyberark/conjur-authn-k8s-client#234</a> (task)</td>
</tr>
<tr class="even">
<td>9</td>
<td>CI to run unit tests for the Namespace prep Helm chart</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td></td>
<td></td>
<td></td>
</tr>
<tr class="odd">
<td>10</td>
<td>CI to run the tests described in the “Helm release validation tests” section above</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/241">cyberark/conjur-authn-k8s-client#241</a></td>
</tr>
<tr class="even">
<td>11</td>
<td>Integration / E2E CI for Conjur Open Source</td>
<td>cyberark/<br />
conjur-authn-k8s-client</td>
<td><ul>
<li><p>Can be either Jenkins+GKE or Github Actions + KinD</p></li>
<li><p>Create script to Helm install Conjur Open Source</p></li>
<li><p>Run scripts from Task #5 above</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/242">cyberark/conjur-authn-k8s-client#242</a></td>
</tr>
<tr class="odd">
<td>12</td>
<td>Integration / E2E CI for Conjur Enterprise with Followers in Kubernetes cluster</td>
<td></td>
<td><ul>
<li><p>Create script that runs dap-intro scripts to deploy a Conjur instance with followers in the Kubernetes cluster</p></li>
<li><p>Run scripts from Task #5 above</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/243">cyberark/conjur-authn-k8s-client#243</a></td>
</tr>
<tr class="even">
<td>13</td>
<td>Integration / E2E CI for Conjur Enterprise with Followers outside Kubernetes cluster</td>
<td></td>
<td><ul>
<li><p>Create script that runs dap-intro scripts to deploy a Conjur instance with followers outside of the Kubernetes cluster (in an AWS VM?)</p></li>
<li><p>Run scripts from Task #5 above</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/244">cyberark/conjur-authn-k8s-client#244</a></td>
</tr>
<tr class="odd">
<td>14</td>
<td>Integration / E2E CI for testing workflow in OpenShift</td>
<td></td>
<td><ul>
<li><p>Use script from Task #8 that runs dap-intro scripts to deploy a Conjur Enterprise instance with followers outside of the Kubernetes cluster (in an AWS VM?)</p></li>
<li><p>Add OpenShift login steps to scripts from Task #5</p></li>
<li><p>Run scripts from Task #5 above using Conjur Enterprise cluster</p></li>
</ul></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/245">cyberark/conjur-authn-k8s-client#245</a></td>
</tr>
<tr class="even">
<td>15</td>
<td>Create Quick-Start guide</td>
<td>cyberark/<br />
conjur-authn-<br />
k8s-client</td>
<td></td>
<td></td>
<td><p>Use:</p>
<ul>
<li><p>Cluster prep Helm chart</p></li>
<li><p>Namespace prep Helm chart</p></li>
<li><p>Sample application deployment Helm chart from Task #5<br />
<br />
<a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/246">cyberark/conjur-authn-k8s-client#246</a></p></li>
</ul></td>
</tr>
<tr class="odd">
<td>16</td>
<td>Transfer of Information (TOI): Work with tech writers to prepare documentation</td>
<td></td>
<td></td>
<td></td>
<td><a href="https://github.com/cyberark/conjur-authn-k8s-client/issues/235">cyberark/conjur-authn-k8s-client#235</a></td>
</tr>
</tbody>
</table>

# Definition of Done

<table>
<thead>
<tr class="header">
<th></th>
<th><strong>DoD Criterion</strong></th>
<th><strong>Status</strong><br />
<strong>(Done / Not Done)</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>1</td>
<td>Solution design is approved</td>
<td></td>
</tr>
<tr class="even">
<td>2</td>
<td>Test plan is reviewed</td>
<td></td>
</tr>
<tr class="odd">
<td>3</td>
<td>Security review of design is completed</td>
<td></td>
</tr>
<tr class="even">
<td>4</td>
<td>Tests are implemented according to test plan</td>
<td></td>
</tr>
<tr class="odd">
<td>5</td>
<td>The proposed workflow is documented in a Quick-Start guide and/or Katacoda scenario</td>
<td></td>
</tr>
<tr class="even">
<td>6</td>
<td>Versions are bumped in all relevant projects</td>
<td></td>
</tr>
</tbody>
</table>

# Solution Review

<table>
<thead>
<tr class="header">
<th><strong>Persona</strong></th>
<th><strong>Name</strong></th>
<th><strong>Design Approval</strong></th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Team Leader</td>
<td>Dane LeBlanc</td>
<td><ul>
<li><blockquote>
<p> ✅</p>
</blockquote></li>
</ul></td>
</tr>
<tr class="even">
<td>Product Owner</td>
<td>Alex Kalish</td>
<td><ul>
<li><blockquote>
<p> ✅</p>
</blockquote></li>
</ul></td>
</tr>
<tr class="odd">
<td>System Architect</td>
<td>Rafi Schwarz</td>
<td><ul>
<li><blockquote>
<p> ✅</p>
</blockquote></li>
</ul></td>
</tr>
<tr class="even">
<td>Security Architect</td>
<td>Andy Tinkham</td>
<td><ul>
<li><blockquote>
<p>✅ </p>
</blockquote></li>
</ul></td>
</tr>
<tr class="odd">
<td>QA Architect</td>
<td>Andy Tinkham</td>
<td><ul>
<li><blockquote>
<p> ✅</p>
</blockquote></li>
</ul></td>
</tr>
</tbody>
</table>

# Appendix

## Notes from C&I group collaboration

####  Problem

*At present there is no clear guidance for consuming the Kubernetes
authenticator except for samples that don’t use best practices and are
more suited to quick demos. The challenge of consuming the Kubernetes
authenticator in production is that it needs to take into account
maintainability*

*Bash scripts relied on environment variables, and merged personas into
one. We were not allowing for sharing of information between personas.*

1.  *Kubernetes cluster preparation*

*...*

2.  *App namespace preparation*

*...*

#### General approach

*We’ve split out the stages of the deployment to a level of granularity
that takes into account the different persons involved.*

*Highlight the steps and make suggestions for improvements.*

*The goal is to keep information flowing between the separate phases of
deployment. This makes the deployment more maintainable.*

1.  *Call out the distinct stages*

2.  *Call out the standard forms of configuration that will flow through
    the stages*

#### Helm based approach

*We provide automation for the distinct stages using helm charts.*

*\[insert diagrams here\]*

*Collect information once and save it somewhere. Make the information
consumable from application, as opposed to hardcoding.*  
  
*==============================================================*  
*==============================================================*
