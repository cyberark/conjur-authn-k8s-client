FROM golang:1.12 as authenticator-client-builder
MAINTAINER Conjur Inc

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /opt/conjur-authn-k8s-client
COPY . /opt/conjur-authn-k8s-client

EXPOSE 8080

RUN apt-get update && apt-get install -y jq

RUN go get -u github.com/jstemmer/go-junit-report && \
    go get github.com/smartystreets/goconvey

RUN go build -a -installsuffix cgo -o authenticator ./cmd/authenticator

# =================== BUSYBOX LAYER ===================
# this layer is used to get binaries into the main container
FROM busybox

# =================== MAIN CONTAINER ===================
FROM scratch as authenticator-client
MAINTAINER CyberArk Software, Inc.

# copy a few commands from busybox
COPY --from=busybox /bin/tar /bin/tar
COPY --from=busybox /bin/sleep /bin/sleep
COPY --from=busybox /bin/sh /bin/sh
COPY --from=busybox /bin/ls /bin/ls
COPY --from=busybox /bin/id /bin/id
COPY --from=busybox /bin/whoami /bin/whoami
COPY --from=busybox /bin/mkdir /bin/mkdir
COPY --from=busybox /bin/chmod /bin/chmod
COPY --from=busybox /bin/cat /bin/cat

# allow anyone to write to this dir, container may not run as root
RUN mkdir -p /etc/conjur/ssl && chmod 777 /etc/conjur/ssl

VOLUME /run/conjur

COPY --from=authenticator-client-builder /opt/conjur-authn-k8s-client/authenticator /bin

CMD ["authenticator"]

# =================== MAIN CONTAINER (REDHAT) ===================
FROM registry.access.redhat.com/rhel as authenticator-client-redhat

MAINTAINER CyberArk Software, Inc.

# allow anyone to write to this dir, container may not run as root
RUN mkdir -p /etc/conjur/ssl && chmod 777 /etc/conjur/ssl && mkdir -p /licenses

RUN useradd -ms /bin/bash conjur

VOLUME /run/conjur

COPY --from=authenticator-client-builder /opt/conjur-authn-k8s-client/authenticator /bin

ADD LICENSE /licenses

USER conjur

CMD ["authenticator"]

ARG VERSION

LABEL name="conjur-authn-k8s-client"
LABEL vendor="CyberArk"
LABEL version="$VERSION"
LABEL release="$VERSION"
LABEL summary="Conjur OpenShift Authentication Client for use with Conjur"
LABEL description="The authentication client required to expose secrets from a Conjur server to applications running within OpenShift"
