#!/bin/bash

cd api
for GOOS in darwin linux ; do
    GOARCH=amd64
    architecture="${GOOS}-${GOARCH}"
    echo "Building ${architecture} ${path}"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/tat-${architecture}
    echo "file bin/tat-${architecture}"
    file bin/tat-${architecture}
    echo "ldd bin/tat-${architecture}"
done
