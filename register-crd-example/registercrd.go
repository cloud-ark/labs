package main

import (
	"fmt"
	"flag"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	crdtypedef "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		fmt.Println("Error building kubeconfig: %s", err.Error())
	}

	crdclient, err := apiextensionsclientset.NewForConfig(cfg)
	registerCRD(crdclient)
}


func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func registerCRD(crdclient *apiextensionsclientset.Clientset) {
	fmt.Println("Inside registerCRD")

	crdName := "postgreses.postgrescontroller.kubeplus"
	crdGroup := "postgrescontroller.kubeplus"
	crdVersion := "v1"
	crdKind := "Postgres"
	crdPlural := "postgreses"
	postgrescrd := &crdtypedef.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: crdtypedef.CustomResourceDefinitionSpec{
			Group: crdGroup,
			Version: crdVersion,
			Names: crdtypedef.CustomResourceDefinitionNames{
				Plural: crdPlural,
				Kind: crdKind,
			},
		},
	}
	postgrescrdObj, err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(postgrescrd)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Postgres object created:%+v", postgrescrdObj)
	}
}
