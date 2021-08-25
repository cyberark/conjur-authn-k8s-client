FROM alpine:latest

RUN mkdir -p /src
WORKDIR /src

RUN apk update && \
    apk upgrade libcrypto1.1 && \
    apk add jq
