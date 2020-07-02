package test_case

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/account"
	cmd "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"strconv"
)

const (
	PARAM    = "param"
	TRANSFER = "transfer"
	ADMIN    = "admin"
	OPERATOR = "operator"
)

func parseAddr(value []byte) (common.Address, error) {
	bf := common.NewZeroCopySource(value)
	address, err := utils.DecodeAddress(bf)
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("decode addr, %s", err)
	}
	return address, nil
}

func fetchGlobalGasPrice(sdk *ontology_go_sdk.OntologySdk) (uint64, error) {
	param, err := sdk.Native.GlobalParams.GetGlobalParams([]string{"gasPrice"})
	if err != nil {
		return 0, err
	}
	gasPriceStr, ok := param["gasPrice"]
	if !ok {
		return 0, fmt.Errorf("param not exist")
	}
	gasPrice, err := strconv.ParseUint(gasPriceStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return gasPrice, nil
}

func fetchAccounts(sdk *ontology_go_sdk.OntologySdk) (payer *ontology_go_sdk.Account,
	admins []*ontology_go_sdk.Account, err error) {
	wallet, err := sdk.OpenWallet("../wallets/wallet_multi.json")
	if err != nil {
		return
	}
	pwd := []byte("123456")
	payer, err = wallet.GetDefaultAccount(pwd)
	if err != nil {
		return
	}
	adminPubKeys := []string{
		"0261e3344b70c38cdaff538e4a97993930764aad17b44f69de8329e24357fb9eab",
		"03b23fa36103d1e2b2f5bed153aa583d4388ed2d2a58ea6fa25f9e77820e5dc01f",
		"02455f4dfcb4eafb1bde298243d65daacc15bbb7feb3e256f91e8a9435cea8dfa7",
		"020c26270a9a7dadc101071c603bc3a7d370fccf9a66c8042d9829e4377bd5edec",
		"02f9e7c9c1734a2223d05b6c99cf69824fcaec777640b761d1cfd195de62417916",
		"039965917b8906fde226d137e0a6851dd62b2a2e2837958e238bcd6e45c6213531",
		"0227ac9fc4b3c9ab892308cd1d413349fefcef2fb3b6a7eaf3f5945df51f6982a7"}
	admins = make([]*ontology_go_sdk.Account, 0)
	for _, pubKey := range adminPubKeys {
		pubKeyData, err := hex.DecodeString(pubKey)
		if err != nil {
			return nil, nil, err
		}
		pub, err := keypair.DeserializePublicKey(pubKeyData)
		if err != nil {
			return nil, nil, err
		}
		addr := types.AddressFromPubKey(pub)
		acc, err := wallet.GetAccountByAddress(addr.ToBase58(), pwd)
		if err != nil {
			return nil, nil, err
		}
		admins = append(admins, acc)
	}
	return
}

func genMultiSigAddr(accounts []*ontology_go_sdk.Account, m int) (pubkeys []keypair.PublicKey,
	addr common.Address, err error) {
	for _, acc := range accounts {
		pubkeys = append(pubkeys, acc.PublicKey)
	}
	addr, err = types.AddressFromMultiPubKeys(pubkeys, m)
	return
}

func updateGasPrice(sdk *ontology_go_sdk.OntologySdk, txGasPrice, gasLimit, destinationGasPrice uint64,
	pubKeys []keypair.PublicKey, admins []*ontology_go_sdk.Account) (string, error) {
	updateGasPriceTx, err := sdk.Native.GlobalParams.NewSetGlobalParamsTransaction(txGasPrice, gasLimit,
		map[string]string{"gasPrice": fmt.Sprint(destinationGasPrice)})
	if err != nil {
		return "", fmt.Errorf("create tx, %s", err)
	}
	m := uint16((5*len(pubKeys) + 6) / 7)
	for _, signer := range admins {
		acc := account.Account(*signer)
		err = cmd.MultiSigTransaction(updateGasPriceTx, m, pubKeys, &acc)
		if err != nil {
			return "", fmt.Errorf("multi sig tx, %s", err)
		}
	}
	updateGasPriceTxHash, err := sdk.SendTransaction(updateGasPriceTx)
	if err != nil {
		return "", fmt.Errorf("send tx, %s", err)
	}
	return updateGasPriceTxHash.ToHexString(), nil
}

func createSnapshot(sdk *ontology_go_sdk.OntologySdk, txGasPrice, gasLimit uint64, pubKeys []keypair.PublicKey,
	admins []*ontology_go_sdk.Account) (string, error) {
	createSnapshotTx, err := sdk.Native.GlobalParams.NewCreateSnapshotTransaction(txGasPrice, gasLimit)
	if err != nil {
		return "", fmt.Errorf("create tx, %s", err)
	}
	m := uint16((5*len(pubKeys) + 6) / 7)
	for _, signer := range admins {
		acc := account.Account(*signer)
		err = cmd.MultiSigTransaction(createSnapshotTx, m, pubKeys, &acc)
		if err != nil {
			return "", fmt.Errorf("multi sig tx, %s", err)
		}
	}
	createSnapshotTxHash, err := sdk.SendTransaction(createSnapshotTx)
	if err != nil {
		return "", fmt.Errorf("send tx, %s", err)
	}
	return createSnapshotTxHash.ToHexString(), nil
}
