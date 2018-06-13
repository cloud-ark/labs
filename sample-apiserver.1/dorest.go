package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"net/http"
	"io/ioutil"

	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	//masterURL  string
	//kubeconfig string
)

func main() {

	fmt.Printf("1\n")

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("2\n")
	fmt.Printf("Client CFG:%v", cfg)

	cfg.NegotiatedSerializer = scheme.Codecs
	cfg.GroupVersion = &schema.GroupVersion{Group:"apps", Version:"v1"}

	restClient, err := rest.RESTClientFor(cfg)

	fmt.Printf("==========================\n")

	if err != nil {
		panic(err)
	}

	fmt.Printf("Restclient:%v\n", restClient)
	fmt.Printf("==========================\n")

	req, _ := http.NewRequest("GET", "https://192.168.99.100:8443/apis/apps/v1/namespaces/default/deployments", nil)
	req.Header.Add("Accept", "application/json")
    resp, err := restClient.Client.Do(req)
    if err != nil {
        fmt.Println("Errored when sending request to the server")
        fmt.Println(err)
    }

    //defer resp.Body.Close()
    resp_body, _ := ioutil.ReadAll(resp.Body)

    fmt.Println(resp.Status)
    fmt.Println(string(resp_body))

	//resp := restClient.Verb("GET").AbsPath("/apis/apps/v1/namespaces/default/deployments").Do().Into()

	/*
	resp, err := restClient.Verb("GET").AbsPath("/apis/apps/v1/namespaces/default").Do().Get()   //Resource("pods").Do().Get()

	if err != nil {
		fmt.Println(err)
	}
	*/

	//fmt.Printf("Response:%v\n", resp)
	//fmt.Printf("Response status:%v", resp)
	//fmt.Printf("Response body:%v", resp.body)

	//if err != nil {
	//	panic(err)
	//}

}


func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}