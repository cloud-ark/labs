#!/bin/bash

rm sample-apiserver
rm artifacts/simple-image/kube-sample-apiserver
go build .
cp ./sample-apiserver ./artifacts/simple-image/kube-sample-apiserver