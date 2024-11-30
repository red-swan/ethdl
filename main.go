package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {

	// Provide usage -------------------------------------------------------------
	helps := []string{"--help", "-help", "-h", "help"}
	if slices.Contains(helps, strings.ToLower(os.Args[1])) {
		lines := []string{
			"ethdl",
			"-out <PATH>: the output directory to save the downloaded contracts to.",
			"             defaults to <current-directory>/<the address we're downloading from>",
			"-etherscan-api-key <KEY>: your etherscan api key",
			"                          see https://docs.etherscan.io/getting-started/viewing-api-usage-statistics",
			"                          this will read the ETHERSCAN_API_KEY environment variable if provided by",
			"                          the system or a local .env file",
			"<ADDRESS>: the etherscan address to download contracts from. Currently only mainnet is possible.",
		}
		for _, line := range lines {
			fmt.Println(line)
		}
	} else {
		// Gathering run params ------------------------------------------------------
		var config CLIFlags = BuildConfig()

		// Call Etherscan ------------------------------------------------------------
		var targetSources SourceCode = GetSources(config.Address, config.EtherScanApiKey)

		// Write source files out ----------------------------------------------------
		WriteSourceCode(targetSources, config.OutputDir)
	}

}
