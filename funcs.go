package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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

func ProvideUsage() error {
	lines := []string{
		"getChainCode <chainId> <address> <optional flags>",
		"chainId can be: mainnet, arbitrum, arbnova, polygon, base, optimism",
		"address need not be checksummed but must be a valid, 40 character string",
		"",
		"Optional Flags:",
		"-out <PATH>: the output directory to save the downloaded contracts to.",
		"             defaults to <current-directory>/<the address we're downloading code for>",
		"-api-key <KEY>: your etherscan api key",
		"                see https://docs.etherscan.io/getting-started/viewing-api-usage-statistics",
		"                if not provided, the cli will read the appropriate environment variable if",
		"                provided by the system or a local .env file",
		"                Default variable: ETHERSCAN_API_KEY",
	}
	return errors.New(strings.Join(lines, "\n"))
}

func BuildConfig() (ProgramConfig, error) {
	var config ProgramConfig

	// parse help arg ----------------------------------------
	helps := []string{"--help", "-help", "-h", "help"}
	if len(os.Args) < 3 || slices.Contains(helps, strings.ToLower(os.Args[1])) {
		return config, ProvideUsage()
	}

	// parse chain arg ---------------------------------------
	chains := [...]string{"mainnet", "arbitrum", "arbnova", "polygon", "base", "optimism"}
	specifiedChain := strings.ToLower(os.Args[1])
	if !slices.Contains(chains[:], specifiedChain) {
		chainErr := errors.New("must provide a valid chain id")
		return config, errors.Join(chainErr, ProvideUsage())
	}
	switch specifiedChain {
	case "mainnet":
		config.ChainId = "1"
	case "arbitrum":
		config.ChainId = "42161"
	case "arbnova":
		config.ChainId = "42170"
	case "polygon":
		config.ChainId = "137"
	case "base":
		config.ChainId = "8453"
	case "optimism":
		config.ChainId = "10"
	default:
		log.Fatalf("The logic for chains is not complete: %s went unhandled", specifiedChain)
	}

	// Handle the address --------------------------
	specifiedAddress := os.Args[2]
	if !isAddressString(specifiedAddress) {
		errAddress := errors.New("specified address is not correct. It must start with 0x and be 42 characters long")
		return config, errors.Join(errAddress, ProvideUsage())
	}
	config.Address = specifiedAddress

	// Gathering flags -----------------------------
	pathPtr := flag.String("out", ".", "The ouptut path, will default to the current directory.")
	keyPtr := flag.String("etherscan-api-key", "", "Your etherscan api key")
	flag.Parse()
	config.ApiKey = *keyPtr
	config.OutputDir = *pathPtr

	// Handle default etherscan key ----------------
	var envVar string = "ETHERSCAN_API_KEY"
	if config.ApiKey == "" {
		// Load an env file, ignore any errors
		godotenv.Load()
		config.ApiKey = os.Getenv(envVar)
	}
	if config.ApiKey == "" {
		errApiKey := errors.New("No key provided and could not find " + envVar + " envvar")
		return config, errors.Join(errApiKey, ProvideUsage())
	}

	// Handle default output path ------------------
	if config.OutputDir == "." || config.OutputDir == "" {
		wd, err := os.Getwd()
		panicIfNotNil("Error when getting working directory: %v", err)
		// defaults to the working directory + the address we're downloading from
		config.OutputDir = filepath.Join(wd, config.Address)
	}

	return config, nil
}

// Etherscan Functions ---------------------------------------------------------
func CreateSourceCodeEndpoint(chain, address, key string) string {
	v := url.Values{}
	v.Set("chainid", chain)
	v.Set("module", "contract")
	v.Set("action", "getsourcecode")
	v.Set("address", address)
	v.Set("apikey", key)
	output := "https://api.etherscan.io/v2/api?" + v.Encode()
	// fmt.Println(output)
	return output
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

func MustGetResult(config ProgramConfig) JSONResult {

	var apiResonse JSONEndpointResponse
	err := GetJSON(CreateSourceCodeEndpoint(config.ChainId, config.Address, config.ApiKey), &apiResonse)
	panicIfNotNil("Error when gathering JSON: %V", err)
	// status can come back zero
	if apiResonse.Status != "1" {
		panic("API Call returned bad status: " + apiResonse.Message)
	}

	// is there always only 1??
	return apiResonse.Result[0]

}

// Writer functions ------------------------------------------------------------

func MustGetSources(result JSONResult) []SourceCode {
	var output []SourceCode

	r, _ := utf8.DecodeRuneInString(result.SourceCode)
	if r == '{' {
		// sometimes we get inner json

		// We have to trim the source code string to make valid JSON
		// it is wrapped in curly brackets {} for some reason
		var sourceCodeStr string = TrimFirstAndLastChar(result.SourceCode)

		// Build the SourceCode Object from the JSON
		var sourceCode JSONSourceCode
		err := json.NewDecoder(bytes.NewReader([]byte(sourceCodeStr))).Decode(&sourceCode)
		panicIfNotNil("Error when decoding multiple sources from JSON: %v", err)
		for relPath, content := range sourceCode.Sources {
			output = append(output, SourceCode{content.Content, relPath})
		}

	} else {
		// sometimes we don't
		output = append(output, SourceCode{result.SourceCode, result.ContractName + ".sol"})
	}

	return output
}

func MustWriteSourceCode(sourceObj []SourceCode, directory string) {

	for _, source := range sourceObj {

		// fmt.Println(path) // maybe add this with a verbose flag
		var fullpath string = filepath.Join(directory, source.RelativePath)

		err := os.MkdirAll(filepath.Dir(fullpath), 0770)
		panicIfNotNil("Error when making directory: %v", err)

		filePtr, err := os.Create(fullpath)
		panicIfNotNil("Error when creating file: %v", err)

		filePtr.WriteString(source.Content)
		filePtr.Close()
	}

}
