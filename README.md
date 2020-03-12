# conjur-authn-k8s-client

On DockerHub: https://hub.docker.com/r/cyberark/conjur-kubernetes-authenticator/

## What's inside ?

The Conjur Kubernetes authenticator client is designed to have a light footprint both in terms of storage and memory consumption. It has very few components:

+ A static binary for the authenticator
+ The `sleep` binary from busybox for debugging
+ The `tar` binary from busybox to meet the requirement of the authentication service

## Configuration

The client is configured entirely through environment variables. These are listed below.

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
- `CONJUR_TOKEN_TIMEOUT`: Timeout for fetching a new token (defaults to 6 minutes). In most cases, this variable should not be modified.

Flow:

The client's process logs its flow to `stdout` and `stderr`.
+ Exponential backoff is exercised when an error occurs
+ Client will re-login when certificate has expired

1. Client goes through login by presenting certificate signing request (CSR) -> Server (authn-k8s running inside the appliance) injects signed client certificate out of band into requesting pod
1. Client picks up signed client certificate, deletes it from disk and uses to authenticator via mutual TLS -> Server responds with auth token (retrieved via authn-local) encrypted with the public key of the client.
1. Client decrypts the auth token and writes it to to the shared memory volume (`/run/conjur/access-token`)
1. Client proceeds to authenticate time and time again

## Contributing

We welcome contributions of all kinds to this repository. For instructions on how to get started and descriptions of our development workflows, please see our [contributing
guide][contrib].

[contrib]: https://github.com/cyberark/conjur-authn-k8s-client/blob/master/CONTRIBUTING.md
