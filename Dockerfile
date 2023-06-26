FROM golang:1.20 as authenticator-client-builder
MAINTAINER CyberArk Software Ltd.

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1 \
    GOEXPERIMENT=boringcrypto

# this value changes in ./bin/build
ARG TAG_SUFFIX="-dev"
ARG VERSION="unreleased"

# On CyberArk dev laptops, golang module dependencies are downloaded with a
# corporate proxy in the middle. For these connections to succeed we need to
# configure the proxy CA certificate in build containers.
#
# To allow this script to also work on non-CyberArk laptops where the CA
# certificate is not available, we copy the (potentially empty) directory
# and update container certificates based on that, rather than rely on the
# CA file itself.
ADD build_ca_certificate /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /opt/conjur-authn-k8s-client
COPY . /opt/conjur-authn-k8s-client

EXPOSE 8080

RUN apt-get update && apt-get install -y jq

RUN go mod download

RUN go get -u github.com/jstemmer/go-junit-report

RUN go build -a -installsuffix cgo \
    -ldflags="-X 'github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator.TagSuffix=$TAG_SUFFIX' \
        -X 'github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator.Version=$VERSION'" \
    -o authenticator ./cmd/authenticator

# Verify the binary is using BoringCrypto.
# Outputting to /dev/null so the output doesn't include all the files
RUN sh -c "go tool nm authenticator | grep '_Cfunc__goboringcrypto_' 1> /dev/null"

# =================== MAIN CONTAINER ===================
FROM alpine:latest as authenticator-client
MAINTAINER CyberArk Software Ltd.

RUN apk add -u shadow libc6-compat && \
    # Add Limited user
    groupadd -r authenticator \
             -g 777 && \
    useradd -c "authenticator runner account" \
            -g authenticator \
            -u 777 \
            -m \
            -r \
            authenticator && \
    # Ensure authenticator dir is owned by authenticator user and setup a
    # directory for the Conjur client certificate/access token
    mkdir -p /usr/local/lib/authenticator /etc/conjur/ssl /run/conjur && \
    # Use GID of 0 since that is what OpenShift will want to be able to read things
    chown authenticator:0 /usr/local/lib/authenticator \
                       /etc/conjur/ssl \
                       /run/conjur && \
    # We need open group permissions in these directories since OpenShift won't
    # match our UID when we try to write files to them
    chmod 770 /etc/conjur/ssl \
              /run/conjur

# Ensure openssl development libraries are always up to date
RUN apk add --no-cache openssl-dev

USER authenticator

VOLUME /run/conjur

COPY --from=authenticator-client-builder /opt/conjur-authn-k8s-client/authenticator /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/authenticator" ]

# =================== MAIN CONTAINER (REDHAT) ===================
FROM registry.access.redhat.com/ubi9/ubi as authenticator-client-redhat
MAINTAINER CyberArk Software Ltd.

RUN yum -y distro-sync

    # Add Limited user
RUN groupadd -r authenticator \
             -g 777 && \
    useradd -c "authenticator runner account" \
            -g authenticator \
            -u 777 \
            -m \
            -r \
            authenticator && \
    # Ensure plugin dir is owned by authenticator user
    mkdir -p /usr/local/lib/authenticator && \
    # Make and setup a directory for the Conjur client certificate/access token
    mkdir -p /etc/conjur/ssl /run/conjur /licenses && \
    # Use GID of 0 since that is what OpenShift will want to be able to read things
    chown authenticator:0 /usr/local/lib/authenticator \
                       /etc/conjur/ssl \
                       /run/conjur && \
    # We need open group permissions in these directories since OpenShift won't
    # match our UID when we try to write files to them
    chmod 770 /etc/conjur/ssl \
              /run/conjur

VOLUME /run/conjur

COPY --from=authenticator-client-builder /opt/conjur-authn-k8s-client/authenticator /usr/local/bin/

ADD LICENSE /licenses

USER authenticator

CMD [ "/usr/local/bin/authenticator" ]

ARG VERSION

LABEL name="conjur-authn-k8s-client"
LABEL vendor="CyberArk"
LABEL version="$VERSION"
LABEL release="$VERSION"
LABEL summary="Conjur OpenShift Authentication Client for use with Conjur"
LABEL description="The authentication client required to expose secrets from a Conjur server to applications running within OpenShift"

# =================== CONTAINER FOR HELM TEST ===================

FROM golang:1.20-alpine as k8s-cluster-test

# Install packages for testing
RUN apk add --no-cache bash bind-tools coreutils curl git ncurses openssl openssl-dev

# Install bats-core in /usr/local
RUN curl -#L https://github.com/bats-core/bats-core/archive/master.zip | unzip - && \
    bash bats-core-master/install.sh /usr/local && \
    rm -rf ./bats-core-master

# Install bats-support, bats-assert, and bats-files libraries
# These need to be sourced at run time, e.g.:
#    source '/bats/bats-support/load.bash'
#    source '/bats/bats-assert/load.bash'
#    source '/bats/bats-file/load.bash'
RUN git clone https://github.com/ztombol/bats-support /bats/bats-support && \
    git clone https://github.com/ztombol/bats-assert /bats/bats-assert && \
    git clone https://github.com/ztombol/bats-file /bats/bats-file

# Install yq
RUN wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq

# Temporarily update go.mod due to CVE-2022-41723
# This script will fail when the version of golang.org/x/net changes and this section
# will need to be updated or removed.
RUN grep v0.4.1-0.20230214201333-88ed8ca3307d /usr/local/go/src/go.mod
RUN sed -i "s|v0.4.1-0.20230214201333-88ed8ca3307d|v0.7.0|g" /usr/local/go/src/go.mod

RUN mkdir -p /tests
WORKDIR /tests

LABEL name="conjur-k8s-cluster-test"
LABEL vendor="CyberArk"
LABEL version="$VERSION"
LABEL release="$VERSION"
LABEL summary="Conjur Kubernetes test client for use with Helm"
LABEL description="The Conjur test client that is used with Helm test to validate the configuration created by Helm"

ENTRYPOINT ["sleep", "infinity"]
