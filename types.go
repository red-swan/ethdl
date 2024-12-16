package main

// CLI Flags
// dev: will be changed when config file feature is added
type CLIFlags struct {
	OutputDir       string
	EtherScanApiKey string
	Address         string
}

// JSON Demarshalling Types ----------------------------------------------------
type JSONEndpointResponse struct {
	Status  string
	Message string
	Result  []JSONResult
}

type JSONResult struct {
	SourceCode           string // it's a string but could be the source code string or a json string of multiple files
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

type JSONSourceCode struct {
	Language string
	Sources  map[string]JSONContent
	Settings JSONSettings //todo
}
type JSONContent struct {
	Content string
}
type JSONSettings struct {
	Optimizer       JSONOptimizer
	OutputSelection interface{}
	ViaIR           bool
	Libraries       interface{}
}
type JSONOptimizer struct {
	Enabled bool
	Runs    int
	Details interface{}
}

// Convenience Types -----------------------------------------------------------
type SourceCode struct {
	Content      string
	RelativePath string
}
