package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/account"
	cmd "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/types"
	"github.com/qiluge/globalparam/config"
	"github.com/qiluge/globalparam/method"
	"github.com/urfave/cli"
	"io/ioutil"
	"strings"
)

var newGasPriceFlag = cli.Uint64Flag{
	Name:  "new-gas-price",
	Usage: "define destination gas price",
}

var newDeployGasFlag = cli.Uint64Flag{
	Name:  "new-deploy-gas",
	Usage: "define destination gas price",
}

var newMigrateGasFlag = cli.Uint64Flag{
	Name:  "new-migrate-gas",
	Usage: "define destination gas price",
}

var mFlag = cli.Uint64Flag{
	Name:  "m",
	Usage: "define threshold in multi sign",
}

var rawTxFlag = cli.StringFlag{
	Name:  "raw-tx",
	Usage: "serialized transaction",
}

var pubKeysFlag = cli.StringFlag{
	Name:  "pub-keys",
	Usage: "all pub keys in multi sig",
	Value: "aaa,bbb",
}

var rpcAddrFlag = cli.StringFlag{
	Name:  "rpc-addr",
	Usage: "rpc addr which sdk used",
}

var GenUpdateGlobalParamTxCmd = cli.Command{
	Name:   "gen-update-param-tx",
	Usage:  "generate update global param tx",
	Action: genUpdateGasPriceTx,
	Flags: []cli.Flag{
		utils.GasPriceFlag,
		utils.GasLimitFlag,
		newGasPriceFlag,
		newDeployGasFlag,
		newMigrateGasFlag,
	},
}

var GenCreateSnapshotTxCmd = cli.Command{
	Name:   "gen-create-snapshot-tx",
	Usage:  "generate create snapshot tx",
	Action: genCreateSnapshotTx,
	Flags: []cli.Flag{
		utils.GasPriceFlag,
		utils.GasLimitFlag,
	},
}

var MultiSignTxCmd = cli.Command{
	Name:   "multi-sign-tx",
	Usage:  "generate one signature for multi sig tx",
	Action: multiSignTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.AccountAddressFlag,
		mFlag,
		rawTxFlag,
		pubKeysFlag,
	},
}

var SendTxCmd = cli.Command{
	Name:   "send-tx",
	Usage:  "send tx",
	Action: sendTx,
	Flags: []cli.Flag{
		rawTxFlag,
		rpcAddrFlag,
	},
}

var UpdateGlobalParamByCfgCmd = cli.Command{
	Name:   "update-param",
	Usage:  "update global param price by config",
	Action: updateGlobalParamByCfg,
	Flags: []cli.Flag{
		utils.ConfigFlag,
	},
}

var CreateSnapshotByCfgCmd = cli.Command{
	Name:   "create-snapshot",
	Usage:  "create snapshot by config",
	Action: createSnapshotByCfg,
	Flags: []cli.Flag{
		utils.ConfigFlag,
	},
}

func genUpdateGasPriceTx(ctx *cli.Context) error {
	gasPrice := ctx.Uint64(utils.GasPriceFlag.GetName())
	gasLimit := ctx.Uint64(utils.GasLimitFlag.GetName())
	newGasPrice := ctx.Uint64(newGasPriceFlag.GetName())
	newDeployGas := ctx.Uint64(newDeployGasFlag.GetName())
	newMigrageGas := ctx.Uint64(newMigrateGasFlag.GetName())
	sdk := ontology_go_sdk.NewOntologySdk()
	tx, err := sdk.Native.GlobalParams.NewSetGlobalParamsTransaction(gasPrice, gasLimit, map[string]string{
		"gasPrice":                  fmt.Sprint(newGasPrice),
		"Ontology.Contract.Create":  fmt.Sprint(newDeployGas),
		"Ontology.Contract.Migrate": fmt.Sprint(newMigrageGas),
	})
	if err != nil {
		return err
	}
	sink := common.NewZeroCopySink(nil)
	t, err := tx.IntoImmutable()
	if err != nil {
		return err
	}
	t.Serialization(sink)
	fmt.Println(hex.EncodeToString(sink.Bytes()))
	return nil
}

func genCreateSnapshotTx(ctx *cli.Context) error {
	gasPrice := ctx.Uint64(utils.GasPriceFlag.GetName())
	gasLimit := ctx.Uint64(utils.GasLimitFlag.GetName())
	sdk := ontology_go_sdk.NewOntologySdk()
	tx, err := sdk.Native.GlobalParams.NewCreateSnapshotTransaction(gasPrice, gasLimit)
	if err != nil {
		return err
	}
	sink := common.NewZeroCopySink(nil)
	t, err := tx.IntoImmutable()
	if err != nil {
		return err
	}
	t.Serialization(sink)
	fmt.Println(hex.EncodeToString(sink.Bytes()))
	return nil
}

func multiSignTx(ctx *cli.Context) error {
	address := ctx.String(utils.GetFlagName(utils.AccountAddressFlag))
	pwd, err := password.GetAccountPassword()
	if err != nil {
		return err
	}
	wallet, err := cmd.OpenWallet(ctx)
	if err != nil {
		return err
	}
	acc, err := cmd.GetAccountMulti(wallet, pwd, address)
	if err != nil {
		return err
	}
	tx := ctx.String(rawTxFlag.GetName())
	txData, err := hex.DecodeString(tx)
	if err != nil {
		return err
	}
	tran, err := types.TransactionFromRawBytes(txData)
	if err != nil {
		return err
	}
	mutTx, err := tran.IntoMutable()
	if err != nil {
		return err
	}
	pubkeys := strings.Split(ctx.String(pubKeysFlag.GetName()), ",")
	keys := make([]keypair.PublicKey, 0, len(pubkeys))
	for _, pkStr := range pubkeys {
		pkData, err := hex.DecodeString(pkStr)
		if err != nil {
			return err
		}
		pk, err := keypair.DeserializePublicKey(pkData)
		if err != nil {
			return err
		}
		keys = append(keys, pk)
	}
	m := ctx.Uint64(mFlag.GetName())
	if err = utils.MultiSigTransaction(mutTx, uint16(m), keys, acc); err != nil {
		return err
	}
	immutx, err := mutTx.IntoImmutable()
	if err != nil {
		return nil
	}
	sink := common.NewZeroCopySink(nil)
	immutx.Serialization(sink)
	fmt.Println(hex.EncodeToString(sink.Bytes()))
	return nil
}

func sendTx(ctx *cli.Context) error {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(ctx.String(rpcAddrFlag.GetName()))
	rawTx := ctx.String(rawTxFlag.GetName())
	raw, err := hex.DecodeString(rawTx)
	if err != nil {
		return err
	}
	tx, err := types.TransactionFromRawBytes(raw)
	if err != nil {
		return err
	}
	muttx, err := tx.IntoMutable()
	if err != nil {
		return err
	}
	hash, err := sdk.SendTransaction(muttx)
	if err != nil {
		return err
	}
	fmt.Println(hash.ToHexString())
	return nil
}

func updateGlobalParamByCfg(ctx *cli.Context) error {
	cfg, accounts, pubkeys, err := parseCfg(ctx)
	if err != nil {
		return err
	}
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(cfg.RPCAddr)
	txHash, err := method.UpdateGasPrice(sdk, cfg.GasPrice, cfg.GasLimit, pubkeys, accounts, cfg.NewGasPrice,
		cfg.NewDeployGas, cfg.NewMigrateGas)
	if err != nil {
		return err
	}
	fmt.Println(txHash)
	return nil
}

func createSnapshotByCfg(ctx *cli.Context) error {
	cfg, accounts, pubkeys, err := parseCfg(ctx)
	if err != nil {
		return err
	}
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(cfg.RPCAddr)
	txHash, err := method.CreateSnapshot(sdk, cfg.GasPrice, cfg.GasLimit, pubkeys, accounts)
	if err != nil {
		return err
	}
	fmt.Println(txHash)
	return nil
}

func parseCfg(ctx *cli.Context) (*config.Config, []*account.Account, []keypair.PublicKey, error) {
	cfgFilePath := ctx.String(utils.GetFlagName(utils.ConfigFlag))
	cfgContent, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	cfg := &config.Config{}
	if err := json.Unmarshal(cfgContent, cfg); err != nil {
		return nil, nil, nil, err
	}
	accounts := make([]*account.Account, 0)
	pubkeys := make([]keypair.PublicKey, 0)
	for _, walletCfg := range cfg.Wallets {
		if !common.FileExisted(walletCfg.Path) {
			return nil, nil, nil, fmt.Errorf("cannot find wallet file: %s", walletCfg.Path)
		}
		wallet, err := account.Open(walletCfg.Path)
		if err != nil {
			return nil, nil, nil, err
		}
		fmt.Printf("please input account %s in wallet %s password\n", walletCfg.Account, walletCfg.Path)
		pwd, err := password.GetAccountPassword()
		if err != nil {
			return nil, nil, nil, err
		}
		acc, err := cmd.GetAccountMulti(wallet, pwd, walletCfg.Account)
		if err != nil {
			return nil, nil, nil, err
		}
		accounts = append(accounts, acc)
		pubkeys = append(pubkeys, acc.PublicKey)
	}
	multiSigAddr, err := types.AddressFromMultiPubKeys(pubkeys, cfg.M)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("multi sig addr: %s\n", multiSigAddr.ToBase58())
	return cfg, accounts, pubkeys, nil
}
