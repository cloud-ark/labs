Steps:
-------
export GOOS=linux; go build journal.go
docker build -t testprov:24 -f Dockerfile.journal .
kubectl create configmap kind-compositions-config-map --from-file=kind_compositions.yaml
kubectl apply -f journal.yaml
kubectl get pods
kubectl logs "journal-deployment-pod"