package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Gathering run params ------------------------------------------------------
	// Gathering flags
	pathPtr := flag.String("out", ".", "The ouptut path, will default to the current directory.")
	ethPtr := flag.String("etherscan-api-key", "", "Your etherscan api key")
	config := CLIFlags{*pathPtr, *ethPtr, ""}

	// Handle default output path
	if config.OutputDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		config.OutputDir = wd
	}

	// Handle default etherscan key
	if config.OutputDir == "." {
		// Load an env file, ignore any errors
		godotenv.Load()
		config.EtherScanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	}

	// Handle the address
	var remainingArgs []string = flag.Args()
	if len(remainingArgs) != 1 {
		panic("Must provide an address without flag at the end")
	}
	config.Address = remainingArgs[0]

	// Call Etherscan ------------------------------------------------------------
	var apiResonse EndpointResponse
	err = GetJSON(CreateSourceCodeEndpoint(args[1], etherscanKey), &apiResonse)
	if err != nil {
		log.Fatal(err)
	}

	// for _, v := range apiResonse.Result {
	// 	fmt.Println(v.ContractName)
	// }

	// Work the JSON
	var outputPath string = "scratch.json"
	var output *os.File

	if _, err := os.Stat(outputPath); errors.Is(err, os.ErrNotExist) {
		output, err = os.Create(outputPath)
		if err != nil {
			fmt.Println(err)
		}
	}

	defer output.Close()

	var sourceCodeStr string = apiResonse.Result[0].SourceCode
	sourceCodeStr = TrimFirstAndLastChar(sourceCodeStr) // has wrapping {} for some reason
	// fmt.Println(sourceCodeStr)

	var sourceCode SourceCode
	err = json.NewDecoder(bytes.NewReader([]byte(sourceCodeStr))).Decode(&sourceCode)
	if err != nil {
		fmt.Println(err)
	}

	// writing files out

	for path, content := range sourceCode.Sources {
		fmt.Println(path)
		var file, err = CreateFile(path)
		if err != nil {
			fmt.Println(err)
		}
		file.WriteString(content.Content)
		file.Close()
	}

	// _, err = output.WriteString(scratch)
	// if err != nil {
	// 	fmt.Println(err)
	// }

}
