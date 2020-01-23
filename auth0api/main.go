package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/go-resty/resty/v2"
	"gopkg.in/auth0.v2/management"
)

// DataKeys - structure of the response data for a Hashicorp Vault LIST object
// {"request_id":"8dcc5db1-2090-835c-a8b1-a0d3b449f2fd","lease_id":"","renewable":false,"lease_duration":0,"data":{"keys":["config/"]},"wrap_info":null,"warnings":null,"auth":null}
// type DataKeys struct {
// 	RequestID     string `json:"request_id"`
// 	LeaseID       string `json:"lease_id"`
// 	Renewable     bool   `json:"renewable"`
// 	LeaseDuration int64  `json:"lease_duration"`
// 	Data          struct {
// 		Keys []string `json:"keys"`
// 	} `json:"data"`
// 	WrapInfo string `json:"wrap_info"`
// 	Warning  string `json:"warnings"`
// 	Auth     string `json:"auth"`
// }

// ValueData - structure of the response for a Hashicorp Vault GET object
type ValueData struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int64  `json:"lease_duration"`
	Data          struct {
		Data map[string]interface{} `json:"data"`
	} `json:"data"`
	MetaData struct {
		CreatedTime  string `json:"created_time"`
		DeletionTime string `json:"deletion_time"`
		Destroyed    bool   `json:"destroyed"`
		Version      int64  `json:"version"`
	} `json:"metadata"`
	WrapInfo string `json:"wrap_info"`
	Warning  string `json:"warnings"`
	Auth     string `json:"auth"`
}

func main() {

	var clientID string
	home := os.Getenv("HOME")
	vaultToken, err := readFile(fmt.Sprintf("%s/.vault-token", home))
	if err != nil {
		fmt.Println("Cannot get vault access token")
		os.Exit(1)
	}

	client := resty.New()

	vaultAddr := os.Getenv("VAULT_ADDR")
	keypath := "auth0"
	resp, err := client.R().
		SetHeader("X-Vault-Token", string(vaultToken[:])).
		Get(fmt.Sprintf("%s/v1/secret/data/%s", vaultAddr, keypath))

	var response ValueData
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		fmt.Println("Error decoding json")
		os.Exit(1)
	}

	// fmt.Println(response.Data.Data["ManagementDomain"])
	// fmt.Println(response.Data.Data["ManagementClientID"])
	// fmt.Println(response.Data.Data["ManagementClientSecret"])

	domain := fmt.Sprintf("%v", response.Data.Data["ManagementDomain"])
	id := fmt.Sprintf("%v", response.Data.Data["ManagementClientID"])
	secret := fmt.Sprintf("%v", response.Data.Data["ManagementClientSecret"])

	m, err := management.New(domain, id, secret)
	if err != nil {
		fmt.Println("Authentication to Auth0 Failed.")
		os.Exit(1)

	}
	clientList, listerr := m.Client.List()
	if listerr != nil {
		fmt.Println("Cannot get a list of clients")
		os.Exit(1)
	}

	for _, client := range clientList {
		if *client.Name == "Nginx_Ingress_test" {
			//fmt.Println("-----------------------------------------------------------")
			//fmt.Println(fmt.Sprintf("ClientName: %s, ClientID: %s", *client.Name, *client.ClientID))
			//fmt.Println("-----------------------------------------------------------")
			clientID = *client.ClientID
			outputCallbacks(m, clientID)
		}
	}

	//updateCallbacks(m, clientID, "./callbacks.yaml")

}

func updateCallbacks(m *management.Management, clientID string, filename string) {

	yamlData, readerr := readFile(filename)
	if readerr != nil {
		fmt.Println(fmt.Sprintf("Failed to read YAML file %s", filename))
	}

	c := &management.Client{}

	err := yaml.Unmarshal(yamlData, c)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(c)

	m.Client.Update(clientID, c)

}

func outputCallbacks(m *management.Management, clientID string) {
	clientInfo, geterr := m.Client.Read(clientID)
	if geterr != nil {
		fmt.Println(fmt.Sprintf("Failed to get Client Details: %s", clientID))
	}

	// for _, callback := range clientInfo.Callbacks {
	// 	fmt.Println(fmt.Sprintf("%v", callback))
	// }

	//clientInfo.Callbacks = append(clientInfo.Callbacks, "https://mynewfun-aks-dev4admin.dev.splicemachine.dev.io/oauth2/callback")
	c := &management.Client{
		Callbacks: clientInfo.Callbacks,
	}

	//m.Client.Update(clientID, c)

	// keyData, err := json.Marshal(c)
	// if err != nil {
	// 	fmt.Println(fmt.Sprintf("keyData: %s", string(keyData[:])))
	// }
	// //fmt.Println(fmt.Sprintf("keyData: %s", string(keyData[:])))
	// fmt.Println(string(keyData))

	y, yerr := yaml.Marshal(c)
	if yerr != nil {
		fmt.Printf("err: %v\n", yerr)
		os.Exit(1)
	}
	fmt.Println(string(y))

}

func readFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}
