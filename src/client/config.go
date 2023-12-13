package client

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type WalletCfg struct {
	PubKeyPath string `json:"pubKeyPath"`
	PriKeyPath string `json:"priKeyPath"`
}

type ChainCfg struct {
	ChainDataBasePath string `json:"chainDataBasePath"`
}

type P2PNetCfg struct {
	BootstrapPeers []string `json:"bootstrapPeers"`
	PriKeyPath     string   `json:"priKeyPath"`
	Bootstrap      bool     `json:"bootstrap"`
	ListenAddr     string   `json:"listenAddr"`
}

type Config struct {
	WalletCfg `json:"walletCfg"`
	ChainCfg  `json:"chainCfg"`
	P2PNetCfg `json:"p2PNetCfg"`
}

func LoadConfig(file string) (*Config, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		def, _ := DefaultConfig()
		cfgValue, err := json.Marshal(def)
		if err != nil {
			return def, err
		}
		fd, err := os.Create(file)
		if err != nil {
			return def, err
		}
		defer fd.Close()
		_, err = io.Copy(fd, strings.NewReader(string(cfgValue)))
		return def, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// DefaultConfig 生成一个默认配置
func DefaultConfig() (*Config, error) {
	return &Config{
		WalletCfg{
			PubKeyPath: "../wallet/public_key.pem",
			PriKeyPath: "../wallet/private_key.pem",
		},
		ChainCfg{
			ChainDataBasePath: "../database/chain_data",
		},
		P2PNetCfg{
			BootstrapPeers: []string{},
			PriKeyPath:     "../wallet/private_key.pem",
			Bootstrap:      false,
		},
	}, nil
}
