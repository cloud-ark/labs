#!/bin/bash
kubectl delete ns wardle
kubectl delete apiservice v1alpha1.wardle.k8s.io