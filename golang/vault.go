package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Vault struct {
	PostgresqlUsername    string `json:"postgresqlUsername"`
	PostgresqlPassword    string `json:"postgresqlPassword"`
	PostgresqlAddressName string `json:"postgresqlAddressName"`
}

type DataVault struct {
	Data Vault `json:"data"`
}

type GetVaultResponse struct {
	DataVault DataVault `json:"data"`
}

type PutToken struct {
	Role string `json:"role"`
	Jwt  string `json:"jwt"`
}

type Auth struct {
	ClientToken string `json:"client_token"`
}

type KubernetesResponse struct {
	AuthObj Auth `json:"auth"`
}

func getToken(jwt string, vaultAddr string) string {
	role := "goapp"
	auth_path := "auth/kubernetes/login"

	data, err := json.Marshal(PutToken{role, jwt})
	if err != nil {
		log.Fatal(err)
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/%s", vaultAddr, auth_path), bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	// initialize http client
	client := &http.Client{}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.StatusCode)
	defer resp.Body.Close()
	response := &KubernetesResponse{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(response)
	return response.AuthObj.ClientToken
}

func getJWT() string {
	jwtPath := os.Getenv("JWT_PATH")

	file, err := os.Open(jwtPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	log.Println(string(b))
	return string(b)
}

func getFromTokenVault(token string, vaultAddr string) Vault {
	secretWebPath := "secret/data/goapp/config"
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/%s", vaultAddr, secretWebPath), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.StatusCode)
	defer resp.Body.Close()
	response := &GetVaultResponse{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(response)
	return response.DataVault.Data
}

func GetVault() Vault {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		vaultAddr = "http://localhost:8200"
	}

	return getFromTokenVault(getToken(getJWT(), vaultAddr), vaultAddr)
}
