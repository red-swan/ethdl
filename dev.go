package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func createSourceCodeEndpoint(address, key string) string {
	return fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getsourcecode&address=%s&apikey=%s", address, key)
}

func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot fetch URL %q: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}
	// We could check the resulting content type
	// here if desired.
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("cannot decode JSON: %v", err)
	}
	return nil
}

func main() {

	// set up
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	etherscanKey := os.Getenv("ETHERSCAN_KEY")
	args := os.Args

	// get json
	type SourceCode struct {
		SourceCode           string
		ABI                  string
		ContractName         string
		CompilerVersion      string
		Runs                 string
		ConstructorArguments string
		EVMVersion           string
		Library              string
		LicenseType          string
		Proxy                string
		Implmentation        string
		SwarmSource          string
	}
	type EndpointResponse struct {
		Status  string
		Message string
		Result  []SourceCode
	}
	var apiResonse EndpointResponse
	err = getJSON(createSourceCodeEndpoint(args[1], etherscanKey), &apiResonse)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range apiResonse.Result {
		fmt.Println(v.ContractName)
	}

}
