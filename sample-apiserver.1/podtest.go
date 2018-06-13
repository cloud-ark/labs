package main 

import (
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"crypto/x509"
	"crypto/tls"
)

// reference: https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container

func main() {
	fmt.Println("1")
	serviceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	apiVersion := "v1" // For example
	namespace := "default" // For example
	resource := "pods" // For example
	httpMethod := http.MethodGet // For Example

	url1 := fmt.Sprintf("https://%s:%s/api/%s/namespaces/%s/%s", serviceHost, servicePort, apiVersion, namespace, resource)
	fmt.Println("2")
	fmt.Printf("Url:%s\n",url1)
	
	u, err := url.Parse(url1)
	if err != nil {
	  panic(err)
	}
	
	req, err := http.NewRequest(httpMethod, u.String(), nil) //bytes.NewBuffer(payload))
	//req, err := http.Get(url1)
	if err != nil {
	    fmt.Println(err)
	}
	fmt.Println("3")

	caToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
	    panic(err) // cannot find token file
	}
	fmt.Println("4")
	fmt.Printf("Token:%s", caToken)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(caToken)))

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
	    panic(err) // Can't find cert file
	}

	fmt.Println("5")
	fmt.Printf("CaCert:%s",caCert)

	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
	  Transport: &http.Transport{
	    TLSClientConfig: &tls.Config{
	        RootCAs: caCertPool,
	    },
	  },
	}

	fmt.Println("6")

	resp, err := client.Do(req)
	if err != nil {
	    log.Printf("sending request failed: %s", err.Error())
	    fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("6")

	resp_body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))

	fmt.Println("7")
}