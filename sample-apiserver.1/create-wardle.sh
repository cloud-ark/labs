#!/bin/bash

kubectl create ns wardle

kubectl create configmap crdlist --from-file=crds.list -n wardle

kubectl create -f artifacts/example/sa.yaml -n wardle

kubectl create -f artifacts/example/rc.yaml -n wardle

kubectl create -f artifacts/example/service.yaml -n wardle

kubectl create -f artifacts/example/apiservice.yaml