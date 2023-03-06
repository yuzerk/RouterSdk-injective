package config

import (
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
)

// LoadConfig load config
func LoadConfig(configFile string, check bool) *ServerConfig {
	serverConfigFile = configFile
	if serverConfigFile == "" {
		log.Fatal("must specify config file")
	}
	log.Info("LoadConfig start", "path", serverConfigFile)
	if !common.FileExist(serverConfigFile) {
		log.Fatalf("LoadConfig error: config file '%v' not exist", serverConfigFile)
	}
	config := &ServerConfig{}
	if _, err := toml.DecodeFile(serverConfigFile, &config); err != nil {
		log.Fatalf("LoadConfig error (toml DecodeFile): %v", err)
	}

	serverConfig = config

	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Println("LoadConfig finished.", string(bs))

	if check {
		if err := config.CheckConfig(); err != nil {
			log.Fatalf("Check config failed. %v", err)
		}
	}

	return serverConfig
}
