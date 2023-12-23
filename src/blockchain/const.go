package blockchain

const (
	// Block
	MaxTransactionLen = 9                 // MaxTransactionLen defines the maximum number of transactions per block
	GenesisData       = "a Genesis Block" // GenesisData represents the initial data for the genesis block

	// Wallet
	AddressVersion      = 0x00                       // AddressVersion represents the version for wallet addresses
	LocalPublicKeyFile  = "./wallet/public_key.pem"  // LocalPublicKeyFile defines the path for the local public key file
	LocalPrivateKeyFile = "./wallet/private_key.pem" // LocalPrivateKeyFile defines the path for the local private key file

	// Transaction
	Reward = 1 // Reward defines the reward amount for mining a block

	// Chain
	DataBaseFile    = "./database/MANIFEST" // DataBaseFile represents the file containing the blockchain database
	DataBasePath    = "./database"          // DataBasePath defines the path for the blockchain database
	TipHashKey      = "l"                   // TipHashKey represents the key for the tip (latest block hash) in the database
	BlockTable      = "b"                   // BlockTable represents the table storing block data in the database
	ChainStateTable = "c"                   // ChainStateTable represents the table storing chain state in the database
	MaxUTXOSize     = 1024                  // MaxUTXOSize defines the maximum size of the unspent transaction output set
	GenesisValue    = 114514                // GenesisValue represents the initial value for the genesis block
	MinerReward     = 10                    // MinerReward defines the reward for miners when mining a block

	// Transaction pool
	MaxTxPoolSize = 1024 // MaxTxPoolSize defines the maximum size of the transaction pool
)
