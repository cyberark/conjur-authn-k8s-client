FROM goboring/golang:1.12.17b4 as authenticator-client-builder
MAINTAINER CyberArk Software Ltd.

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

# this value is set in ./bin/build
ARG TAG

WORKDIR /opt/conjur-authn-k8s-client
COPY . /opt/conjur-authn-k8s-client

EXPOSE 8080

RUN apt-get update && apt-get install -y jq

RUN go mod download

RUN go get -u github.com/jstemmer/go-junit-report && \
    go get github.com/smartystreets/goconvey

RUN go build -a -installsuffix cgo \
    -ldflags="-X github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator.Tag=$TAG" \
    -o authenticator ./cmd/authenticator

# Verify the binary is using BoringCrypto.
# Outputting to /dev/null so the output doesn't include all the files
RUN sh -c "go tool nm authenticator | grep '_Cfunc__goboringcrypto_' 1> /dev/null"

# =================== BUSYBOX LAYER ===================
# this layer is used to get binaries into the main container
FROM busybox

# =================== MAIN CONTAINER ===================
FROM alpine:latest as authenticator-client
MAINTAINER CyberArk Software Ltd.

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

USER authenticator

VOLUME /run/conjur

COPY --from=authenticator-client-builder /opt/conjur-authn-k8s-client/authenticator /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/authenticator" ]

# =================== MAIN CONTAINER (REDHAT) ===================
FROM registry.access.redhat.com/rhel as authenticator-client-redhat
MAINTAINER CyberArk Software Ltd.

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
