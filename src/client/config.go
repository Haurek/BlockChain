package client

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Configurations for different aspects of the blockchain client.
// These structures define the configuration options for various components.

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
	PriKeyPath string `json:"priKeyPath"`
	Bootstrap  bool   `json:"bootstrap"`
	ListenAddr string `json:"listenAddr"`
	LogPath    string `json:"logPath"`
}

type PBFTCfg struct {
	IsConsensusNode bool   `json:"is_consensus_node"`
	View            uint64 `json:"view"`
	Index           uint64 `json:"index"`
	NodeNum         uint64 `json:"nodeNum"`
	MaxFaultNode    uint64 `json:"maxFaultNode"`
	LogPath         string `json:"logPath"`
}

type TxPoolCfg struct {
	TxPoolFull int    `json:"txPoolFull"`
	LogPath    string `json:"logPath"`
}

type BlockPoolCfg struct {
	BlockPoolFull int    `json:"blockPoolFull"`
	LogPath       string `json:"logPath"`
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

// LoadConfig loads the configuration from a JSON file.
// It reads the provided file, parses its content into a Config struct, and returns it.
func LoadConfig(file string) (*Config, error) {
	// Check if the file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		def := DefaultConfig()
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

	// Open the file for reading
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the file content
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into a Config struct
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		WalletCfg: WalletCfg{
			PubKeyPath: "./wallet/public_key.pem",
			PriKeyPath: "./wallet/private_key.pem",
		},
		ChainCfg: ChainCfg{
			ChainDataBasePath: "./database",
			MaxTxPerBlock:     9,
			LogPath:           "./log/chain.log",
		},
		P2PNetCfg: P2PNetCfg{
			PriKeyPath: "./wallet/private_key.pem",
			Bootstrap:  false,
			ListenAddr: "/ip4/0.0.0.0/tcp/6666",
			LogPath:    "./log/net.log",
		},
		PBFTCfg: PBFTCfg{
			IsConsensusNode: false,
			View:            0,
			Index:           0,
			NodeNum:         0,
			MaxFaultNode:    0,
			LogPath:         "./log/pbft.log",
		},
		ClientCfg: ClientCfg{
			LogPath: "./log/client.log",
		},
		TxPoolCfg: TxPoolCfg{
			TxPoolFull: 0,
			LogPath:    "./log/txpool.log",
		},
		BlockPoolCfg: BlockPoolCfg{
			BlockPoolFull: 0,
			LogPath:       "./log/blockpool.log",
		},
	}
}
