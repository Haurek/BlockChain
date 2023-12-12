package blockchain

const (
	// Block
	Difficulty        = 8
	MaxTransactionLen = 9
	GenesisData       = "a Genesis Block"

	// Wallet
	AddressVersion      = 0x00
	LocalPublicKeyFile  = "./wallet/public_key.pem"
	LocalPrivateKeyFile = "./wallet/private_key.pem"

	// Transaction
	Reward = 1

	// Chain
	DataBaseFile    = "./database/MANIFEST"
	DataBasePath    = "./database"
	TipHashKey      = "l"
	BlockTable      = "b"
	ChainStateTable = "c"
	MaxUTXOSize     = 1024
	GenesisValue    = 114514
	MinerReward     = 10

	// Transaction pool
	MaxTxPoolSize = 1024
)
