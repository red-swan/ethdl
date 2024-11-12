package main

type Result struct {
	SourceCode           string // it's a string and then parsed again, why??
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
	Result  []Result
}
type SourceCode struct {
	Language string
	Sources  map[string]Content
	Settings Settings //todo
}
type Content struct {
	Content string
}
type Settings struct {
	Optimizer       Optimizer
	OutputSelection interface{}
	ViaIR           bool
	Libraries       interface{}
}
type Optimizer struct {
	Enabled bool
	Runs    int
	Details interface{}
}
