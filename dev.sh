#!/usr/bin/env bash

cd $(dirname "$0")
docker build -t conjur-authn-k8s-client-go:builder .
docker run --rm -it -v $PWD:/go/src/github.com/cyberark/conjur-authn-k8s-client conjur-authn-k8s-client-go:builder bash
