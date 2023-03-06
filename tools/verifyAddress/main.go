package main

import (
	"flag"
	"fmt"

	routersdk "github.com/anyswap/RouterSDK-injective/sdk"
)

var (
	paramPrefix  string
	paramAddress string
)

func initFlags() {
	flag.StringVar(&paramPrefix, "prefix", "inj", "prefix, eg. cosmos, sei, etc.")
	flag.StringVar(&paramAddress, "address", "", "address")

	flag.Parse()
}

func main() {
	initFlags()

	res := routersdk.IsValidAddress(paramPrefix, paramAddress)
	if res {
		fmt.Printf("address %s is valid (prefix: %v)\n", paramAddress, paramPrefix)
	} else {
		fmt.Printf("address %s is invalid (prefix: %v)\n", paramAddress, paramPrefix)
	}
}
