#!/usr/bin/env bash

/go/bin/reflex \
    -r "\.go$" \
    -R "_test\.go$" \
    -s -- \
        bash dev/start.sh
