package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ClientCfg struct {
	LogPath string `json:"logPath"`
}

type WalletCfg struct {
	PubKeyPath string `json:"pubKeyPath"`
	PriKeyPath string `json:"priKeyPath"`
}

type ChainCfg struct {
	ChainDataBasePath string `json:"chainDataBasePath"`
	MaxTxPerBlock     int    `json:"maxTxPerBlock"`
	LogPath           string `json:"logPath"`
}

type P2PNetCfg struct {
	BootstrapPeers []string `json:"bootstrapPeers"`
	PriKeyPath     string   `json:"priKeyPath"`
	Bootstrap      bool     `json:"bootstrap"`
	ListenAddr     string   `json:"listenAddr"`
	LogPath        string   `json:"logPath"`
}

type PBFTCfg struct {
	View         uint64 `json:"view"`
	Index        uint64 `json:"index"`
	NodeNum      uint64 `json:"nodeNum"`
	MaxFaultNode uint64 `json:"maxFaultNode"`
	LogPath      string `json:"logPath"`
}

type TxPoolCfg struct {
	LogPath string `json:"logPath"`
}

type BlockPoolCfg struct {
	LogPath string `json:"logPath"`
}

type Config struct {
	WalletCfg    `json:"walletCfg"`
	ChainCfg     `json:"chainCfg"`
	P2PNetCfg    `json:"p2PNetCfg"`
	PBFTCfg      `json:"PBFTCfg"`
	ClientCfg    `json:"clientCfg"`
	TxPoolCfg    `json:"txPoolCfg"`
	BlockPoolCfg `json:"blockPoolCfg"`
}

func LoadConfig(file string) (*Config, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
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
