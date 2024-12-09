package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/joho/godotenv"
)

// Utility Functions -----------------------------------------------------------
// Generic panic on error
func panicIfNotNil(s string, err error) {
	if err != nil {
		log.Fatalf(s, err)
	}
}

func TrimFirstAndLastChar(s string) string {
	r := []rune(s)
	return string(r[1 : len(r)-1])
}

func isAddressString(s string) bool {
	return strings.HasPrefix(s, "0x") && utf8.RuneCountInString(s) == 42
}

// CLI Arg Handler -------------------------------------------------------------

func BuildConfig() CLIFlags {
	// Gathering flags
	pathPtr := flag.String("out", ".", "The ouptut path, will default to the current directory.")
	keyPtr := flag.String("etherscan-api-key", "", "Your etherscan api key")
	flag.Parse()
	config := CLIFlags{*pathPtr, *keyPtr, ""}

	// Handle default etherscan key ----------------
	if config.EtherScanApiKey == "" {
		// Load an env file, ignore any errors
		godotenv.Load()

		config.EtherScanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	}
	if config.EtherScanApiKey == "" {
		panic("No key provided and could not find ETHERSCAN_API_KEY envvar")
	}

	// Handle the address --------------------------
	var remainingArgs []string = flag.Args()
	if len(remainingArgs) != 1 {
		panic("Must provide an address and only an address at the end")
	}
	if !isAddressString(remainingArgs[0]) {
		panic("Address string is not correct format: " + remainingArgs[0])
	}
	config.Address = remainingArgs[0]

	// Handle default output path ------------------
	if config.OutputDir == "." {
		wd, err := os.Getwd()
		panicIfNotNil("Error when getting working directory: %v", err)
		// defaults to the working directory + the address we're downloading from
		config.OutputDir = filepath.Join(wd, config.Address)
	}

	return config
}

// Etherscan Functions ---------------------------------------------------------
func CreateSourceCodeEndpoint(address, key string) string {
	return fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getsourcecode&address=%s&apikey=%s", address, key)
}

func GetJSON(url string, result interface{}) error {
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

func GetSources(address, apiKey string) SourceCode {

	var apiResonse EndpointResponse
	err := GetJSON(CreateSourceCodeEndpoint(address, apiKey), &apiResonse)
	panicIfNotNil("Error when gathering JSON: %V", err)
	// status can come back zero
	if apiResonse.Status != "1" {
		panic("API Call returned bad status: " + apiResonse.Message)
	}

	// We have to trim the source code string to make valid JSON
	// it is wrapped in curly brackets {} for some reason
	var sourceCodeStr string = TrimFirstAndLastChar(apiResonse.Result[0].SourceCode)

	// Build the SourceCode Object from the JSON
	var sourceCode SourceCode
	err = json.NewDecoder(bytes.NewReader([]byte(sourceCodeStr))).Decode(&sourceCode)
	panicIfNotNil("Error when decoding JSON: %v", err)

	return sourceCode

}

// Writer functions ------------------------------------------------------------
func WriteSourceCode(sourceObj SourceCode, directory string) {

	for relativepath, content := range sourceObj.Sources {
		// fmt.Println(path) // maybe add this with a verbose flag
		var fullpath string = filepath.Join(directory, relativepath)

		err := os.MkdirAll(filepath.Dir(fullpath), 0770)
		panicIfNotNil("Error when making directory: %v", err)

		filePtr, err := os.Create(fullpath)
		panicIfNotNil("Error when creating file: %v", err)

		filePtr.WriteString(content.Content)
		filePtr.Close()
	}

}
