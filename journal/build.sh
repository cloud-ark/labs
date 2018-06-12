#!/bin/bash

dockerImage="$1"
echo "Docker Image:$dockerImage"

rm journal
export GOOS=linux; go build journal.go

docker build -t $dockerImage .

