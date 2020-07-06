package method

import (
	"encoding/json"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"testing"
	"time"
)

func TestGetGlobalParamsContractEvent(t *testing.T) {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress("http://172.168.3.224:20336")
	for i := uint32(0); i < 10000; i++ {
		evts, err := sdk.ClientMgr.GetSmartContractEventByBlock(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, evt := range evts {
			for _, notify := range evt.Notify {
				if notify.ContractAddress == "0400000000000000000000000000000000000000" {
					jsonEvt, err := json.Marshal(evt)
					if err != nil {
						t.Fatal(err)
					}
					t.Log(string(jsonEvt))
				}
			}
		}
	}
}

func TestGetStorage(t *testing.T) {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress("http://polaris2.ont.io:20336")
	contract := "0400000000000000000000000000000000000000"
	transferAdmin, err := sdk.GetStorage(contract, []byte(TRANSFER))
	if err != nil {
		t.Fatal(err)
	}
	if len(transferAdmin) > 0 {
		transferAdminAddr, err := parseAddr(transferAdmin)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("transferAdminAddr", transferAdminAddr.ToBase58())
	}
	admin, err := sdk.GetStorage(contract, []byte(ADMIN))
	if err != nil {
		t.Fatal(err)
	}
	if len(admin) > 0 {
		adminAddr, err := parseAddr(admin)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("adminAddr", adminAddr.ToBase58())
	}
	operator, err := sdk.GetStorage(contract, []byte(OPERATOR))
	if err != nil {
		t.Fatal(err)
	}
	if len(operator) > 0 {
		operatorAddr, err := parseAddr(operator)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("operatorAddr", operatorAddr.ToBase58())
	}
	paramsKey := append([]byte(PARAM), 0x00)
	params, err := sdk.GetStorage(contract, paramsKey)
	if err != nil {
		t.Fatal(err)
	}
	globalParams := new(global_params.Params)
	source := common.NewZeroCopySource(params)
	if err := globalParams.Deserialization(source); err != nil {
		t.Fatal(err)
	}
	for _, p := range *globalParams {
		t.Logf("%s: %s", p.Key, p.Value)
	}
}

// cmd gasPrice is 500, global gasPrice is 500.
// change global gasPrice to 2500, wait next update period so new gasPrice effects.
// send a tx that's gasPrice < 500( cmd gasPrice), this tx should be not accepted;
// send a tx that's 500( cmd gasPrice) < gasPrice < 2500( global gasPrice), should be not accepted;
// a tx that's gasPrice >= 2500( global gasPrice), should be accepted;
func TestGasPrice(t *testing.T) {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress("http://172.168.3.224:20336")

	payer, admins, err := fetchAccounts(sdk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("payer", payer.Address.ToBase58())
	pubKeys, multiSigAddr, err := genMultiSigAddr(admins, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(multiSigAddr.ToBase58())

	currentGasPrice, err := fetchGlobalGasPrice(sdk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("currentGasPrice", currentGasPrice)
	oldGasPrice := uint64(500)
	gasLimit := uint64(20000)
	paySomeONGHash, err := sdk.Native.Ong.Transfer(oldGasPrice, gasLimit, payer, payer, multiSigAddr, 100*1000000000)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("paySomeONGHash", paySomeONGHash.ToHexString())
	_, _ = sdk.WaitForGenerateBlock(30, 1)
	destinationGasPrice := uint64(2500)
	newDeployGas := uint64(4000000)
	newMigrateGas := uint64(4000000)
	updateGasPriceTxHash, err := UpdateGasPrice(sdk, oldGasPrice, gasLimit, pubKeys, admins, destinationGasPrice,
		newDeployGas, newMigrateGas)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("updateGasPriceTxHash", updateGasPriceTxHash)
	_, _ = sdk.WaitForGenerateBlock(30*time.Second, 1)
	createSnapshotTxHash, err := CreateSnapshot(sdk, oldGasPrice, gasLimit, pubKeys, admins)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("createSnapshotTxHash", createSnapshotTxHash)
	currentBlockHeight, err := sdk.GetCurrentBlockHeight()
	if err != nil {
		t.Fatal(err)
	}
	for {
		_, _ = sdk.WaitForGenerateBlock(30*time.Second, 1)
		height, err := sdk.GetCurrentBlockHeight()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("current height", height)
		// send tx to accelerate block generation
		sdk.Native.Ont.Transfer(oldGasPrice, gasLimit, payer, payer, payer.Address, 1)
		if height/100 > currentBlockHeight/100 {
			break
		}
	}
	// query new gasPrice
	newGasPrice, err := fetchGlobalGasPrice(sdk)
	if err != nil {
		t.Fatal(err)
	}
	if newGasPrice != destinationGasPrice {
		t.Fatalf("%d vs %d", newGasPrice, destinationGasPrice)
	}
	amount := uint64(1)
	// a tx with gasPrice < cmdGasPrice
	minGasPriceTx, err := sdk.Native.Ont.Transfer(oldGasPrice-100, gasLimit, payer, payer, multiSigAddr, amount)
	if err == nil {
		t.Fatalf("min gasPrice tx should failed, not %s", minGasPriceTx.ToHexString())
	}
	t.Log(err)
	// tx with cmdGasPrice< gasPrice < globalGasPrice
	middleGasPriceTx, err := sdk.Native.Ont.Transfer(oldGasPrice+500, gasLimit, payer, payer, multiSigAddr, amount)
	if err == nil {
		t.Fatalf("middle gasPrice tx should failed, not %s", middleGasPriceTx.ToHexString())
	}
	t.Log(err)
	// tx with globalGasPrice <= gasPrice
	highGasPriceTx, err := sdk.Native.Ont.Transfer(destinationGasPrice, gasLimit, payer, payer, multiSigAddr, amount)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("highGasPriceTx", highGasPriceTx.ToHexString())
	// restore gasPrice
	oldDeployGas := uint64(20000000)
	oldMigrateGas := uint64(20000000)
	restoreGasPriceTxHash, err := UpdateGasPrice(sdk, destinationGasPrice, gasLimit, pubKeys, admins, oldGasPrice,
		oldDeployGas, oldMigrateGas)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("restoreGasPriceTxHash", restoreGasPriceTxHash)
	_, _ = sdk.WaitForGenerateBlock(30*time.Second, 1)
	createSnapshotTxHash, err = CreateSnapshot(sdk, destinationGasPrice, gasLimit, pubKeys, admins)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("createSnapshotTxHash", createSnapshotTxHash)
}
