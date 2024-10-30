package eos

import (
	"context"
	"cryp-kaspad/internal/libs/eos/client"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

// func (e *Eos) GetBalanceEOS(addrStr string) (decimal.Decimal, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), utils.Time30S)
// 	defer cancel()

// 	result, err := e.getBalance(ctx, addrStr)
// 	if err != nil {
// 		return decimal.Decimal{}, fmt.Errorf("get balance wei error: %s", err)
// 	}

// 	return balance, nil
// }

// func (e *Eos) GetBalanceWei(ctx context.Context, addrStr string) (decimal.Decimal, error) {
// 	result, err := e.getBalance(ctx, addrStr)
// 	if err != nil {
// 		return decimal.Decimal{}, fmt.Errorf("get balance wei error: %s", err)
// 	}

// 	balance := decimal.NewFromBigInt(result, 0)

// 	return balance, nil
// }

// func (e *Eos) getBalance(ctx context.Context, addrStr string) (*big.Int, error) {
// 	addr := common.HexToAddress(addrStr)

// 	// 注意: nil = 取最新餘額
// 	result, err := e.Client.BalanceAt(ctx, addr, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("BalanceAt error: %s", err)
// 	}

// 	return result, nil
// }

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyz"
	numberBytes    = "12345"
	firstCharBytes = letterBytes
	otherCharBytes = letterBytes + numberBytes
)

func (e *Eos) GetBalanceToken(ctx context.Context, accountId, contractId, symbol string) (decimal.Decimal, error) {
	balance, err := e.Client.GetTokenBalance(ctx, accountId, contractId, symbol)
	if err != nil {
		return decimal.Decimal{}, err
	}

	//回傳是個陣列，不曉得多筆情況，先取第一筆

	if len(balance) == 0 {
		return decimal.Decimal{}, fmt.Errorf("balance array is empty, accountId=%s, contractId=%s, symbol=%s", accountId, contractId, symbol)
	}

	split := strings.Split(balance[0], " ")

	if len(split) != 2 {
		return decimal.Decimal{}, fmt.Errorf("balance split is invalid, accountId=%s, contractId=%s, symbol=%s", accountId, contractId, symbol)
	}

	return decimal.NewFromString(split[0])
}

func (e *Eos) CreateAccount(ctx context.Context, newAccountStr, newAccountPubKey string, creatorConf client.CreatorConfig) error {
	err := e.Client.ImportPrivateKey(ctx, creatorConf.PrivateKey)
	if err != nil {
		return fmt.Errorf("ImportPrivateKey error: %s", err)
	}

	creator := eos.AccountName(creatorConf.AccountID)
	newAccount := eos.AccountName(newAccountStr)

	pk, err := ecc.NewPublicKey(newAccountPubKey)
	if err != nil {
		return fmt.Errorf("NewPublicKey error: %s", err)
	}

	// 构建创建账户的Action
	newAccountAction := system.NewNewAccount(creator, newAccount, pk)

	// 构建购买RAM的Action（假设购买8KB RAM）
	buyRAMAction := system.NewBuyRAMBytes(creator, newAccount, 2048)

	// openAction := eos.Action{
	// 	Account: eos.AccountName("eosio.token"),
	// 	ActionData: eos.ActionData{
	// 		Data: Open{
	// 			Owner:    newAccountName,
	// 			Symbol:   "4,EOS",
	// 			RamPayer: "shikanoko.gm",
	// 		},
	// 	},
	// }
	// 构建抵押CPU和NET的Action（分别抵押0.1 EOS）
	//delegateBWAction := system.NewDelegateBW(creator, newAccount, eos.NewEOSAsset(1000), eos.NewEOSAsset(1000), false)

	// 构建转账一些初始EOS到新账户的Action（可选，这里转0.1 EOS）
	//transferAction := token.NewTransfer(creator, newAccount, eos.NewEOSAsset(1000), "Initial balance")

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(ctx, e.Client.API); err != nil {
		return fmt.Errorf("FillFromChain error: %s", err)
	}

	tx := eos.NewTransaction([]*eos.Action{
		newAccountAction,
		buyRAMAction,

		//delegateBWAction,
		//transferAction,
	}, txOpts)

	signedTx, packedTx, err := e.Client.API.SignTransaction(ctx, tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		return fmt.Errorf("SignTransaction error: %s", err)
	}

	content, err := json.MarshalIndent(signedTx, "", "  ")
	if err != nil {
		return fmt.Errorf("MarshalIndent error: %s", err)
	}

	logrus.WithFields(logrus.Fields{
		"signedTx": string(content),
	}).Info("CreateAccount sigendTx")

	response, err := e.Client.API.PushTransaction(ctx, packedTx)
	if err != nil {
		return fmt.Errorf("PushTransaction error: %s", err)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("CreateAccount success")

	return nil
}

func (e *Eos) ExistAccount(ctx context.Context, accountName string) (bool, error) {
	account, err := e.Client.API.GetAccount(ctx, eos.AN(accountName))
	if err != nil {
		if strings.Contains(err.Error(), "Account Query Exception") {
			return false, nil
		}

		if errors.Is(err, eos.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("GetAccount error: %s", err)
	}

	if account == nil {
		return false, nil
	}

	return true, nil
}

func RandomEOSName() string {
	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, 12)

	// 生成第一個字符（必須是字母）
	b[0] = firstCharBytes[rand.Intn(len(firstCharBytes))]

	// 生成剩餘的 11 個字符
	for i := 1; i < 12; i++ {
		b[i] = otherCharBytes[rand.Intn(len(otherCharBytes))]
	}

	return string(b)
}

func GenerateKeyPair() (string, string, error) {
	priv, err := ecc.NewRandomPrivateKey()
	if err != nil {
		return "", "", err
	}

	return priv.PublicKey().String(), priv.String(), nil
}

func AssetConvertToDecimal(assetStr string) (decimal.Decimal, error) {
	split := strings.Split(assetStr, " ")

	if len(split) != 2 {
		return decimal.Decimal{}, fmt.Errorf("asset split is invalid, assetStr=%s", assetStr)
	}

	return decimal.NewFromString(split[0])
}
