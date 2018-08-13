Example of registering Custom Resource Definition (CRD) through Go
--------------------------------------------------------------------

1. dep ensure

2. kubectl get customresourcedefinition

3. go run registercrd.go -kubeconfig=$HOME/.kube/config

4. kubectl get customresourcedefinition

5. kubectl delete customresourcedefinition postgreses.postgrescontroller.kubeplus