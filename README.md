# conjur-authn-k8s-client

Available images:
- [DockerHub](https://hub.docker.com/r/cyberark/conjur-authn-k8s-client)
- [RedHat Container Registry](https://catalog.redhat.com/software/containers/cyberark/conjur-openshift-authenticator/5c67286cecb5240adf708252)

## What's inside ?

The Conjur Kubernetes authenticator client is designed to have a light footprint both in terms of storage and memory consumption. It has very few components:

+ A static binary for the authenticator
+ The `sleep` binary from busybox for debugging
+ The `tar` binary from busybox to meet the requirement of the authentication service

## Configuration

The client is configured entirely through environment variables. These are listed below.

### Using conjur-authn-k8s-client with Conjur Open Source 

Are you using this project with [Conjur Open Source](https://github.com/cyberark/conjur)? Then we 
**strongly** recommend choosing the version of this project to use from the latest [Conjur OSS 
suite release](https://docs.conjur.org/Latest/en/Content/Overview/Conjur-OSS-Suite-Overview.html). 
Conjur maintainers perform additional testing on the suite release versions to ensure 
compatibility. When possible, upgrade your Conjur version to match the 
[latest suite release](https://docs.conjur.org/Latest/en/Content/ReleaseNotes/ConjurOSS-suite-RN.htm); 
when using integrations, choose the latest suite release that matches your Conjur version. For any 
questions, please contact us on [Discourse](https://discuss.cyberarkcommons.org/c/conjur/5).

## Orchestrator
- `MY_POD_NAME`: Pod name (see [downwards API](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information))
- `MY_POD_NAMESPACE`: Pod namespace (see [downwards API](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information))
- `CONTAINER_MODE`: Set this to `init` to run as an init container that will exit after performing authentication. All other values (including blank) will cause the container to run as a sidecar.

## Conjur
- `CONJUR_VERSION`: Conjur version ('4' or '5', defaults to '5'). Must use a string value in the manifest due to YAML parsing not handling integer values well.
- `CONJUR_ACCOUNT`: Conjur account name
- `CONJUR_AUTHN_URL`: URL pointing to authenticator service endpoint
- `CONJUR_AUTHN_LOGIN`: Host login for pod e.g. `namespace/service_account/some_service_account`
- `CONJUR_SSL_CERTIFICATE`: Public SSL cert for Conjur connection
- `CONJUR_TOKEN_TIMEOUT`: Timeout for fetching a new token (defaults to 6 minutes). 
                          In most cases, this variable should not be modified. The value should be in a
                          format that can be parsed with [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) (e.g "6m0s")

Flow:

The client's process logs its flow to `stdout` and `stderr`.
+ Exponential backoff is exercised when an error occurs
+ Client will re-login when certificate has expired

1. Client goes through login by presenting certificate signing request (CSR) -> Server (authn-k8s running inside the Conjur Enterprise) injects signed client certificate out of band into requesting pod
1. Client picks up signed client certificate, deletes it from disk and uses to authenticator via mutual TLS -> Server responds with auth token (retrieved via authn-local) encrypted with the public key of the client.
1. Client decrypts the auth token and writes it to to the shared memory volume (`/run/conjur/access-token`)
1. Client proceeds to authenticate time and time again

## Running Authenticator Client with a Non-Default User ID in Kubernetes

By default, the Conjur Kubernetes authenticator client container runs using
a default username `authenticator`, user ID `777`, and group ID `777`.

If you would like to run the authenticator client on a *non-OpenShift*
Kubernetes platform, using a non-default user and/or group ID in a Pod that
includes the authenticator client as a sidecar or init container, then you
can configure your Pod manifest as follows:

_**NOTE:** This technique is not supported on OpenShift platforms. For
   OpenShift platforms, the authenticator container should be run
   with the container's default user and group._

- Configure the
  [Pod's Security Context](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#podsecuritypolicy-v1beta1-policy)
  for the desired user ID / group ID. Setting the `fsGroup` to the desired
  user group will cause Kubernetes to set that group as the owner of
  any files that are created in volumes of type `emptyDir`, including the
  authenticator client's SSL certificate and the application's Conjur access
  token.

  For example, to run with a user ID of `65534` (the `nobody` user) and a
  group ID of `65534` (the `nobody` group):

  ```
        securityContext:
          fsGroup: 65534
          runAsGroup: 65534
          runAsNonRoot: true
          runAsUser: 65534
  ```

- Include a `volumeMount` for the authenticator client certificate directory:

  ```
          volumeMounts:
          - name: client-ssl
            mountPath: /etc/conjur/ssl
  ```

- Include an `emptyDir` volume for the authenticator client certificate
  directory. Using a volume of type `emptyDir` allows the client certificate
  file to be created with its group owner set to the value of `fsGroup` as
  configured in the above PodSecurityContext:

  ```
          volumes:
          - name: client-ssl
            emptyDir:
              medium: Memory
  ```

## Contributing

We welcome contributions of all kinds to this repository. For instructions on how to get started and descriptions of our development workflows, please see our [contributing
guide][contrib].

[contrib]: https://github.com/cyberark/conjur-authn-k8s-client/blob/master/CONTRIBUTING.md
