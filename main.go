package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// set up
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	etherscanKey := os.Getenv("ETHERSCAN_KEY")
	args := os.Args

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
