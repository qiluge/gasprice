package config

type Config struct {
	Wallets       []*WalletAccount
	M             int
	RPCAddr       string
	GasPrice      uint64
	GasLimit      uint64
	NewGasPrice   uint64
	NewDeployGas  uint64
	NewMigrateGas uint64
}

type WalletAccount struct {
	Path    string
	Account string
}
