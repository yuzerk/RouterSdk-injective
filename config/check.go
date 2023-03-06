package config

import (
	"fmt"

	"github.com/anyswap/CrossChain-Router/v3/common"
)

// CheckConfig check config
func (c *ServerConfig) CheckConfig() (err error) {
	if c.ChainID == "" {
		return fmt.Errorf("must specify 'ChainID'")
	}
	if c.GatewayConfig.IsEmpty() {
		return fmt.Errorf("empty 'GatewayConfig'")
	}
	for _, tok := range c.SessionTokens {
		pubKey := tok.Token
		if common.IsHex(pubKey) {
			pkBytes := common.FromHex(pubKey)
			if (len(pkBytes) == 65 && pkBytes[0] == 4) ||
				(len(pkBytes) == 33 && (pkBytes[0] == 2 || pkBytes[0] == 3)) {
				continue
			}
		}
		return fmt.Errorf("wrong session token: %v", tok.Token)
	}
	return nil
}
