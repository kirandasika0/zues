#!/bin/bash

build_source()
{
    GOOS=$(uname)
    GOARCH=$(uname -m)
    DIR=$(pwd)
    if [ "$GOOS" = "Darwin" ]; then
        go build -o $DIR/zues $DIR/main.go
    fi
}

build_source