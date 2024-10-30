package vo

import (
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/libs/response"
	"errors"

	"github.com/shopspring/decimal"
)

type WithdrawCreateReq struct {
	MerchantID string `json:"merchant_id" binding:"required"`

	CryptoType string `json:"crypto_type" binding:"required"`
	ChainType  string `json:"chain_type" binding:"required"`

	SecretKey string `json:"secret_key" binding:"required"`

	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`

	Amount decimal.Decimal `json:"amount" binding:"required"`
	Memo   string          `json:"memo"`
	TxId   string          `json:"tx_id"`
}

func (req *WithdrawCreateReq) Validate() (response.Status, error) {
	_, ok := domain.MerchantID2Type[req.MerchantID]
	if !ok {
		respStatus := response.CodeParamInvalid.WithMsg(", merchant_id is not exist")
		return respStatus, errors.New(respStatus.Messages)
	}

	if req.ChainType != domain.ChainType {
		respStatus := response.CodeParamInvalid.WithMsg(", chain_type need use KASPAD")
		return respStatus, errors.New(respStatus.Messages)
	}

	return response.Status{}, nil
}

type WithdrawCreateResp struct {
	TxHash string `json:"tx_hash"`
	Memo   string `json:"memo"`
}
