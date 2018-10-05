#!/bin/bash

kubectl delete deployment moodle1
kubectl delete service moodle1
kubectl delete pv moodle1
kubectl delete pvc moodle1