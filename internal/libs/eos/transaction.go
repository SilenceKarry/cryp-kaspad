package eos

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type Transfer struct {
	From     eos.AccountName `json:"from"`
	To       eos.AccountName `json:"to"`
	Quantity eos.Asset       `json:"quantity"`
	Memo     string          `json:"memo"`
}

func (e *Eos) SendTransaction(ctx context.Context, fromAccountID, toAccountID, privateKey string, amount decimal.Decimal, token entity.Tokens, memo string) (string, error) {
	err := e.Client.ImportPrivateKey(ctx, privateKey)
	if err != nil {
		return "", fmt.Errorf("import private key error: %s", err)
	}

	from := eos.AccountName(fromAccountID)
	to := eos.AccountName(toAccountID)
	quantity, err := eos.NewFixedSymbolAssetFromString(eos.Symbol{
		Precision: uint8(token.Decimals),
		Symbol:    token.CryptoType,
	}, fmt.Sprintf("%s %s", amount.String(), token.CryptoType))

	if err != nil {
		return "", fmt.Errorf("NewFixedSymbolAssetFromString error: %s", err)
	}

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(ctx, e.Client.API); err != nil {
		return "", fmt.Errorf("fill from chain error: %s", err)
	}

	transferAction := &eos.Action{
		Account: eos.AN(token.ContractAddr),
		Name:    eos.ActN("transfer"), //如果有token不是用transfer的話，要調整
		Authorization: []eos.PermissionLevel{
			{Actor: from, Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(Transfer{
			From:     from,
			To:       to,
			Quantity: quantity,
			Memo:     memo,
		}),
	}

	logrus.WithFields(logrus.Fields{
		"action": "transfer",
	}).Info("transferAction Info")

	tx := eos.NewTransaction([]*eos.Action{transferAction}, txOpts)
	signedTx, packedTx, err := e.Client.API.SignTransaction(ctx, tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error: %s", err)
	}

	content, err := json.MarshalIndent(signedTx, "", "  ")
	if err != nil {
		return "", fmt.Errorf("MarshalIndent error: %s", err)
	}

	logrus.WithFields(logrus.Fields{
		"signedTx": string(content),
	}).Info("SendTransaction sigendTx")

	response, err := e.Client.API.PushTransaction(ctx, packedTx)
	if err != nil {
		return "", fmt.Errorf("PushTransaction error: %s", err)
	}

	fmt.Printf("Transaction [%s] submitted to the network succesfully.\n", hex.EncodeToString(response.Processed.ID))

	return hex.EncodeToString(response.Processed.ID), nil
}

func (e *Eos) GetTransaction(ctx context.Context, txHash string) (*eos.TransactionResp, error) {
	return e.Client.API.GetTransaction(ctx, txHash)
}
