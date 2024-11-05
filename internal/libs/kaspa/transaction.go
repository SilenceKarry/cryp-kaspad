package kaspa

import (
	"context"

	"cryp-kaspad/internal/domain/entity"
	"encoding/hex"

	"fmt"

	"github.com/kaspanet/go-secp256k1"
	_ "github.com/kaspanet/go-secp256k1"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/eoscanada/eos-go"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript" // 簽署交易
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util"
	"github.com/shopspring/decimal"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type Transfer struct {
	From     eos.AccountName `json:"from"`
	To       eos.AccountName `json:"to"`
	Quantity eos.Asset       `json:"quantity"`
	Memo     string          `json:"memo"`
}
type UTXOInfo struct {
	Outpoint     *appmessage.RPCOutpoint
	UTXOEntry    *appmessage.RPCUTXOEntry
	ScriptPubKey string
}
type RpcOutpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
	Index         uint32 `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
}

type RpcTransactionInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PreviousOutpoint *RpcOutpoint                    `protobuf:"bytes,1,opt,name=previousOutpoint,proto3" json:"previousOutpoint,omitempty"`
	SignatureScript  string                          `protobuf:"bytes,2,opt,name=signatureScript,proto3" json:"signatureScript,omitempty"`
	Sequence         uint64                          `protobuf:"varint,3,opt,name=sequence,proto3" json:"sequence,omitempty"`
	SigOpCount       uint32                          `protobuf:"varint,5,opt,name=sigOpCount,proto3" json:"sigOpCount,omitempty"`
	VerboseData      *RpcTransactionInputVerboseData `protobuf:"bytes,4,opt,name=verboseData,proto3" json:"verboseData,omitempty"`
}

type RpcTransactionInputVerboseData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
type RPCTransactionVerboseData struct {
	TransactionID string
	Hash          string
	Mass          uint64
	BlockHash     string
	BlockTime     uint64
}

type RPCScriptPublicKey struct {
	Version uint16
	Script  string
}
type Transaction struct {
	Version      uint16
	Inputs       []Input
	Outputs      []Output
	LockTime     uint64
	SubnetworkID []byte
	Gas          uint64
	PayloadHash  []byte
	Payload      []byte
}

// Input 代表交易輸入
type Input struct {
	PreviousTxID     []byte
	PreviousOutIndex uint32
	SignatureScript  []byte
	Sequence         uint64
}

// Output 代表交易輸出
type Output struct {
	Value        uint64
	ScriptPubKey []byte
}

// UTXO 代表未花費輸出
type UTXO struct {
	OutIndex      uint32
	TxID          []byte
	Amount        uint64
	ScriptPubKey  []byte
	BlockDaaScore uint64
	IsCoinbase    bool
}

// KaspaClient 代表Kaspa客戶端
type KaspaClient struct {
	rpcAddress string
	privateKey []byte
	publicKey  []byte
}

func (e *Eos) SendTransaction(ctx context.Context, fromAccountID, toAccountID, privateKey string, amount decimal.Decimal, token entity.Tokens, memo string) (string, error) {

	intValue := amount.IntPart() // 转为 int64 整数部分
	uintAmount := uint64(intValue)

	// 1. 创建交易输入
	utxos, err := e.RpcClient.GetUTXOsByAddresses([]string{fromAccountID})
	if err != nil {
		return "", fmt.Errorf("failed to get UTXOs: %v", err)
	}

	// 2. 選擇合適的UTXO
	var selectedUtxos []*appmessage.RPCUTXOEntry
	var totalAmount uint64
	for _, utxo := range utxos.Entries {
		selectedUtxos = append(selectedUtxos, utxo.UTXOEntry)
		totalAmount += utxo.UTXOEntry.Amount
		if totalAmount >= uintAmount {
			break
		}
	}

	if totalAmount < uintAmount {
		return "", fmt.Errorf("insufficient funds: got %d, need %d", totalAmount, amount)
	}
	// 3. 創建交易
	// 解析私鑰
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %v", err)
	}

	// 創建新交易
	tx := &appmessage.RPCTransaction{
		Version:  constants.MaxTransactionVersion,
		Inputs:   make([]*appmessage.RPCTransactionInput, 0),
		Outputs:  make([]*appmessage.RPCTransactionOutput, 0),
		LockTime: 0,
		//SubnetworkID: constants.DefaultSubnetworkID,
		Gas:     0,
		Payload: "",
	}
	inputs := make([]*externalapi.DomainTransactionInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		transactionID, err := transactionid.FromString(input.PreviousOutpoint.TransactionID)
		if err != nil {
			return "", fmt.Errorf("transactionid.FromString: %v", err)
		}
		domainOut := &externalapi.DomainOutpoint{
			TransactionID: *transactionID,
			Index:         input.PreviousOutpoint.Index,
		}
		signatureScript, err := hex.DecodeString(input.SignatureScript)
		if err != nil {
			return "", fmt.Errorf("inhex.DecodeString: %v", err)
		}
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *domainOut,
			SignatureScript:  signatureScript,
			Sequence:         input.Sequence,
			SigOpCount:       input.SigOpCount,
		}
	}
	outputs := make([]*externalapi.DomainTransactionOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		scriptPublicKey, err := hex.DecodeString(output.ScriptPublicKey.Script)
		if err != nil {
			return "", fmt.Errorf("outhex.DecodeString: %v", err)
		}
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           output.Amount,
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: scriptPublicKey, Version: output.ScriptPublicKey.Version},
		}
	}
	subnetworkID, err := subnetworks.FromString(tx.SubnetworkID)
	if err != nil {
		return "", fmt.Errorf("subnetworks.FromString: %v", err)
	}
	payload, err := hex.DecodeString(tx.Payload)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString: %v", err)
	}
	domainTransaction := &externalapi.DomainTransaction{
		Version:      tx.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     tx.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          tx.Gas,
		Payload:      payload,
	}
	submitTransactionResponse, err := e.RpcClient.SubmitTransaction(appmessage.DomainTransactionToRPCTransaction(domainTransaction), consensushashing.TransactionID(domainTransaction).String(), false)
	if err != nil {
		return "", fmt.Errorf("SubmitTransaction: %v", err)
	}
	fmt.Println("submitTransactionResponse:", submitTransactionResponse)

	return "", nil
}

type UTXOEntry struct {
	Address   string
	Outpoint  *appmessage.Outpoint
	UTXOEntry *appmessage.UTXOEntry
	Amount    uint64
}

func BuildTransaction(entry *appmessage.UTXOsByAddressesEntry, fromAccountID, toAccountID, privateKey, romprivateKey string) (*appmessage.RPCTransaction, string) {
	transactionIDBytes, err := hex.DecodeString(entry.Outpoint.TransactionID)
	if err != nil {
		fmt.Printf("Error decoding transaction ID: %s", err)
	}
	transactionID, err := transactionid.FromBytes(transactionIDBytes)
	if err != nil {
		fmt.Printf("Error decoding transaction ID: %s", err)
	}

	txIns := make([]*appmessage.TxIn, 1)
	txIns[0] = appmessage.NewTxIn(appmessage.NewOutpoint(transactionID, entry.Outpoint.Index), []byte{}, 0, 1)

	payeeAddress, err := util.DecodeAddress(fromAccountID, util.Bech32PrefixKaspaSim)
	if err != nil {
		fmt.Printf("Error decoding payeeAddress: %+v", err)
	}
	toScript, err := txscript.PayToAddrScript(payeeAddress)
	if err != nil {
		fmt.Printf("Error generating script: %+v", err)
	}

	txOuts := []*appmessage.TxOut{appmessage.NewTxOut(entry.UTXOEntry.Amount-1000, toScript)}

	fromScriptCode, err := hex.DecodeString(entry.UTXOEntry.ScriptPublicKey.Script)
	if err != nil {
		fmt.Printf("Error decoding script public key: %s", err)
	}
	fromScript := &externalapi.ScriptPublicKey{Script: fromScriptCode, Version: 0}
	fromAmount := entry.UTXOEntry.Amount

	msgTx := appmessage.NewNativeMsgTx(constants.MaxTransactionVersion, txIns, txOuts)

	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		fmt.Printf("Error decoding private key: %+v", err)
	}
	getPrivateKey, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		fmt.Printf("Error deserializing private key: %+v", err)
	}

	tx := appmessage.MsgTxToDomainTransaction(msgTx)
	tx.Inputs[0].UTXOEntry = utxo.NewUTXOEntry(fromAmount, fromScript, false, 500)

	signatureScript, err := txscript.SignatureScript(tx, 0, consensushashing.SigHashAll, getPrivateKey,
		&consensushashing.SighashReusedValues{})
	if err != nil {
		fmt.Printf("Error signing transaction: %+v", err)
	}
	msgTx.TxIn[0].SignatureScript = signatureScript

	domainTransaction := appmessage.MsgTxToDomainTransaction(msgTx)
	return appmessage.DomainTransactionToRPCTransaction(domainTransaction), consensushashing.TransactionID(domainTransaction).String()
}

// estimateFee 估算交易手續費
func estimateFee(numInputs int) uint64 {
	// 基於輸入數量計算手續費
	// 這裡使用簡單的計算方式，實際使用時可能需要更複雜的計算
	return uint64(numInputs * 1000)
}
