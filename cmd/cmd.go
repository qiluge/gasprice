package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	cmd "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
	"strings"
)

var destinationGasPriceFlag = cli.Uint64Flag{
	Name:  "dest-gas-price",
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

var GenUpdateGasPriceTxCmd = cli.Command{
	Name:   "gen-update-gasprice-tx",
	Usage:  "generate update gas price tx",
	Action: genUpdateGasPriceTx,
	Flags: []cli.Flag{
		utils.GasPriceFlag,
		utils.GasLimitFlag,
		destinationGasPriceFlag,
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

func genUpdateGasPriceTx(ctx *cli.Context) error {
	gasPrice := ctx.Uint64(utils.GasPriceFlag.GetName())
	gasLimit := ctx.Uint64(utils.GasLimitFlag.GetName())
	destinationGasPrice := ctx.Uint64(destinationGasPriceFlag.GetName())
	sdk := ontology_go_sdk.NewOntologySdk()
	tx, err := sdk.Native.GlobalParams.NewSetGlobalParamsTransaction(gasPrice, gasLimit, map[string]string{
		"gasPrice": fmt.Sprint(destinationGasPrice),
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
