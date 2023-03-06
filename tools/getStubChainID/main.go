package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anyswap/CrossChain-Router/v3/log"
	routersdk "github.com/anyswap/RouterSDK-injective/sdk"
)

var (
	paramChainName string
	paramNetwork   string
)

func initFlags() {
	flag.StringVar(&paramChainName, "n", "injective", "chainName, eg. cosmoshub, osmosis, coreum, sei, etc.")
	flag.StringVar(&paramNetwork, "p", "", "network, eg. mainnet, testnet, etc.")

	flag.Parse()
}

func main() {
	initFlags()

	network := paramNetwork
	if network == "" && len(os.Args) > 1 {
		network = os.Args[1]
	}
	if network == "" {
		log.Fatal("miss network argument")
	}

	chainID := routersdk.GetStubChainID(paramChainName, network)
	fmt.Printf("%v %v: %v\n", paramChainName, network, chainID)
}
