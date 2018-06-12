package main 

import (
	"encoding/json"
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	cert "crypto/x509"
	"crypto/tls"
	"context"
	"github.com/coreos/etcd/client"
)

var (
	serviceHost string
	servicePort string
	namespace string
	httpMethod string
	etcdServiceURL string

	kindPluralMap map[string]string
	kindVersionMap map[string]string
	compositionMap map[string]string

	REPLICA_SET string
	DEPLOYMENT string
	POD string
)

type MetaDataAndOwnerReferences struct {
	MetaDataName string
	OwnerReferenceName string
	OwnerReferenceKind string
	OwnerReferenceAPIVersion string
}

type CompositionTreeNode struct {
	Level int
	ChildKind string
	Children []MetaDataAndOwnerReferences
}

func init() {
	serviceHost = os.Getenv("KUBERNETES_SERVICE_HOST")
	servicePort = os.Getenv("KUBERNETES_SERVICE_PORT")
	namespace = "default"
	httpMethod = http.MethodGet

	etcdServiceURL = "http://example-etcd-cluster-client:2379"

	DEPLOYMENT = "Deployment"
	REPLICA_SET = "ReplicaSet"
	POD = "Pod"

	kindPluralMap = make(map[string]string)
	kindPluralMap[DEPLOYMENT] = "deployments"
	kindPluralMap[REPLICA_SET] = "replicasets"
	kindPluralMap[POD] = "pods"

	kindVersionMap = make(map[string]string)
	kindVersionMap[DEPLOYMENT] = "apis/apps/v1"
	kindVersionMap[REPLICA_SET] = "apis/extensions/v1beta1"
	kindVersionMap[POD] = "api/v1"

	compositionMap = make(map[string]string)
	compositionMap[DEPLOYMENT] = "ReplicaSet"
	compositionMap[REPLICA_SET] = "Pod"
}

// Reference: 
// 1. https://stackoverflow.com/questions/30690186/how-do-i-access-the-kubernetes-api-from-within-a-pod-container
// 2. https://www.sohamkamani.com/blog/2017/10/18/parsing-json-in-golang/#unstructured-data
// 3. https://github.com/coreos/etcd/tree/master/client

func main() {
	fmt.Println("1")
	resourceName := "greetings-deployment"  //"podtest5-deployment"
	resourceKind := "Deployment"
	level := 1
	compositionTree := []CompositionTreeNode{}
	fmt.Println("2")

	buildProvenance(resourceKind, resourceName, level, &compositionTree)

	fmt.Println("###################################")
	fmt.Println("Printing the Composition Tree")
	for _, compTreeNode := range compositionTree {
		fmt.Printf("%v\n", compTreeNode)
	}
	fmt.Println("Done printing the Composition Tree")
	fmt.Println("###################################")

	storeProvenance(resourceKind, resourceName, &compositionTree)
}

func storeProvenance(resourceKind string, resourceName string, compositionTree *[]CompositionTreeNode) {
	fmt.Println("Entering storeProvenance")
    jsonCompositionTree, err := json.Marshal(compositionTree)
    if err != nil {
        panic (err)
    }
    resourceProv := string(jsonCompositionTree)
	cfg := client.Config{
		//Endpoints: []string{"http://192.168.99.100:32379"},
		Endpoints: []string{etcdServiceURL},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		//HeaderTimeoutPerRequest: time.Second,
	}
	fmt.Printf("%v\n", cfg)
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)
	// set "/foo" key with "bar" value
	//resourceKey := "/compositions/Deployment/pod42test-deployment"
	//resourceProv := "{1 ReplicaSet; 2 Pod -1}"
	resourceKey := string("/compositions/" + resourceKind + "/" + resourceName)
	fmt.Printf("Setting %s->%s\n",resourceKey, resourceProv)
	resp, err := kapi.Set(context.Background(), resourceKey, resourceProv, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}
	fmt.Printf("Getting value for %s\n", resourceKey)
	resp, err = kapi.Get(context.Background(), resourceKey, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
	fmt.Println("Exiting storeProvenance")
}

func buildProvenance(parentResourceKind string, parentResourceName string, level int, compositionTree *[]CompositionTreeNode) {
	fmt.Printf("$$$$$ Building Provenance Level %d $$$$$ \n", level)
	childResourceKind, present := compositionMap[parentResourceKind]
	if present {
		childKindPlural := kindPluralMap[childResourceKind]
		childResourceApiVersion := kindVersionMap[childResourceKind]
		fmt.Println("3")
		content := getResourceListContent(childResourceApiVersion, childKindPlural)
		fmt.Println("4")
		childrenList := findChildren(content, parentResourceName)
		fmt.Println("5")
		compTreeNode := CompositionTreeNode{
			Level: level,
			ChildKind: childResourceKind,
			Children: childrenList,
		}
		*compositionTree = append(*compositionTree, compTreeNode)
		level = level + 1

		for _, metaDataRef := range childrenList {
			resourceName := metaDataRef.MetaDataName
			resourceKind := childResourceKind
			fmt.Println("6")
			buildProvenance(resourceKind, resourceName, level, compositionTree)
		}
	} else {
		return
	}
}

func getResourceListContent(resourceApiVersion, resourcePlural string) []byte {
	fmt.Println("Entering getResourceListContent")
	url1 := fmt.Sprintf("https://%s:%s/%s/namespaces/%s/%s", serviceHost, servicePort, resourceApiVersion, namespace, resourcePlural)
	fmt.Printf("Url:%s\n",url1)
	caToken := getToken()
	caCertPool := getCACert()
	u, err := url.Parse(url1)
	if err != nil {
	  panic(err)
	}
	req, err := http.NewRequest(httpMethod, u.String(), nil)
	if err != nil {
	    fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(caToken)))
	client := &http.Client{
	  Transport: &http.Transport{
	    TLSClientConfig: &tls.Config{
	        RootCAs: caCertPool,
	    },
	  },
	}
	resp, err := client.Do(req)
	if err != nil {
	    log.Printf("sending request failed: %s", err.Error())
	    fmt.Println(err)
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	fmt.Println("Exiting getResourceListContent")
	return resp_body
}

func findChildren(content []byte, parentResourceName string) []MetaDataAndOwnerReferences {
	fmt.Println("Entering findChildren")
	var result map[string]interface{}
	json.Unmarshal([]byte(content), &result)
	// We need to parse following from the result
	// metadata.name
	// metadata.ownerReferences.name
	// metadata.ownerReferences.kind
	// metadata.ownerReferences.apiVersion
	//parentName := "podtest5-deployment"
	metaDataSlice := []MetaDataAndOwnerReferences{}
	items, ok := result["items"].([]interface{})

	if ok {
		for _, item := range items {
			fmt.Println("=======================")
			itemConverted := item.(map[string]interface{})
			for key, value := range itemConverted {
				if key == "metadata" {
					fmt.Println("----")
					//fmt.Println(key, value.(interface{}))
					metadataMap := value.(map[string]interface{})
					metaDataRef := MetaDataAndOwnerReferences{}
					for mkey, mvalue := range metadataMap {
						fmt.Printf("%v ==> %v\n", mkey, mvalue.(interface{}))
						if mkey == "ownerReferences" {
							ownerReferencesList := mvalue.([]interface{})
							for _, ownerReference := range ownerReferencesList {
								ownerReferenceMap := ownerReference.(map[string]interface{})
								for okey, ovalue := range ownerReferenceMap {
									fmt.Printf("%v --> %v\n", okey, ovalue)
									if okey == "name" {
										metaDataRef.OwnerReferenceName = ovalue.(string)
									}
									if okey == "kind" {
										metaDataRef.OwnerReferenceKind = ovalue.(string)
									}
									if okey == "apiVersion" {
										metaDataRef.OwnerReferenceAPIVersion = ovalue.(string)
									}
								}
							}
						}
						if mkey == "name" {
							metaDataRef.MetaDataName = mvalue.(string)
						}
					}
					metaDataSlice = append(metaDataSlice, metaDataRef)
					fmt.Println("----")
				}
			}
			fmt.Println("=======================")
		}
	}
	metaDataSliceToReturn := []MetaDataAndOwnerReferences{}
	fmt.Println("Printing the MetaDataSlice")
	for _, metaDataRef := range metaDataSlice {
		if metaDataRef.OwnerReferenceName == parentResourceName {
			fmt.Println("%v\n", metaDataRef)
			fmt.Println("*************")
			metaDataSliceToReturn = append(metaDataSliceToReturn, metaDataRef)
		}
	}
	fmt.Println("Exiting findChildren")
	return metaDataSliceToReturn
}


func parse_prev(content []byte) map[string]string {
	var result map[string]interface{}
	json.Unmarshal([]byte(content), &result)

	// We need to parse following from the result
	// metadata.name
	// metadata.ownerReferences.name
	// metadata.ownerReferences.kind
	// metadata.ownerReferences.apiVersion

	var mapToReturn map[string]string

	items := result["items"].([]interface{})
	for _, item := range items {
		fmt.Println("=======================")
		itemConverted := item.(map[string]interface{})
		for key, value := range itemConverted {
			if key == "metadata" {
				fmt.Println("----")
				//fmt.Println(key, value.(interface{}))
				metadataMap := value.(map[string]interface{})
				for mkey, mvalue := range metadataMap {
					fmt.Printf("%v ==> %v\n", mkey, mvalue.(interface{}))
					if mkey == "ownerReferences" {
						ownerReferencesList := mvalue.([]interface{})
						for _, ownerReference := range ownerReferencesList {
							ownerReferenceMap := ownerReference.(map[string]interface{})
							for okey, ovalue := range ownerReferenceMap {
								fmt.Printf("%v --> %v\n", okey, ovalue)
							}
						}
					}
				}
				fmt.Println("----")
			}
		}
		fmt.Println("=======================")
	}
	fmt.Println("**************")
	fmt.Println("Map to Return:")
	for key, value := range mapToReturn {
		fmt.Printf("%v --> %v\n", key, value)
	}
	fmt.Println("**************")
	return mapToReturn
}


func getToken() []byte {
	fmt.Println("Entering getToken")
	caToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
	    panic(err) // cannot find token file
	}
	fmt.Println("4")
	fmt.Printf("Token:%s", caToken)
	fmt.Println("Exiting getToken")
	return caToken
}

func getCACert() *cert.CertPool {
	fmt.Println("Entering getCACert")
	caCertPool := cert.NewCertPool()
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
	    panic(err) // Can't find cert file
	}

	fmt.Println("5")
	fmt.Printf("CaCert:%s",caCert)

	caCertPool.AppendCertsFromPEM(caCert)
	fmt.Println("Exiting getCACert")
	return caCertPool
}

