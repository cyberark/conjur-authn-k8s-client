FROM alpine:3.23

RUN mkdir -p /src
WORKDIR /src

RUN apk update && apk upgrade --no-cache zlib apk-tools libapk && apk add --no-cache jq
