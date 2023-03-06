package config

import (
	"fmt"

	"github.com/anyswap/CrossChain-Router/v3/tokens"
)

var (
	serverConfig     = &ServerConfig{}
	serverConfigFile string
)

// GetServerConfig get server config
func GetServerConfig() *ServerConfig {
	return serverConfig
}

// ServerConfig config items (decode from toml file)
type ServerConfig struct {
	ChainID string

	RouterConfigFile string
	InitRouterServer bool `toml:",omitempty" json:",omitempty"`

	Port             int
	AllowedOrigins   []string        `toml:",omitempty" json:",omitempty"`
	MaxRequestsLimit int             `toml:",omitempty" json:",omitempty"`
	SessionTokens    []*SessionToken `toml:",omitempty" json:",omitempty"`

	GatewayConfig *tokens.GatewayConfig
}

// SessionToken session token
type SessionToken struct {
	Token string
	User  string
	Salt  string `json:"-"`
}

func (t *SessionToken) String() string {
	return fmt.Sprintf("Token %v, User %v", t.Token, t.User)
}
