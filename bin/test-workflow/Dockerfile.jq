FROM alpine:latest

RUN mkdir -p /src
WORKDIR /src

RUN apk update && apk add jq
