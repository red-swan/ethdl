package main

import (
	"log"
)

func main() {

	// Gathering run params ------------------------------------------------------
	config, err := BuildConfig()

	if err != nil {
		log.Fatalf(err.Error())
	}

	// Call Etherscan ------------------------------------------------------------
	var result JSONResult = MustGetResult(config)

	// Massage the Data ----------------------------------------------------------
	var sources []SourceCode = MustGetSources(result)

	// Write source files out ----------------------------------------------------
	MustWriteSourceCode(sources, config.OutputDir)

}
