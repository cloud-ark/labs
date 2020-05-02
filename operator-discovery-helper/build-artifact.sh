#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GOOS=linux; go build .
    cp operator-discovery-helper ./artifacts/deployment/operator-discovery-helper
    docker build -t lmecld/operator-discovery-helper:latest ./artifacts/deployment
    docker push lmecld/operator-discovery-helper:latest
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GOOS=linux; go build .
    cp operator-discovery-helper ./artifacts/deployment/operator-discovery-helper
    docker build -t lmecld/operator-discovery-helper:$version ./artifacts/deployment
    docker push lmecld/operator-discovery-helper:$version
fi



