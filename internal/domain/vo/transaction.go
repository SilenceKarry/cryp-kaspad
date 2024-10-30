package vo

import (
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type TransGetTxHashReq struct {
	TxHash string

	CryptoType string
	ChainType  string
}

type CreateTransactionByBlockNumberReq struct {
	BlockNumber int64
}

func (req *CreateTransactionByBlockNumberReq) Parse(c *gin.Context) (response.Status, error) {
	blockNumberString := c.Param("blockNumber")

	if utils.IsEmpty(blockNumberString) {
		respStatus := response.CodeParamInvalid.WithMsg(", blockNumber is empty")
		return respStatus, errors.New(respStatus.Messages)
	}

	blockNumber, err := strconv.Atoi(blockNumberString)
	if err != nil {
		respStatus := response.CodeInternalError.WithMsg(", blockNumber transfer fail")
		return respStatus, errors.New(respStatus.Messages)
	}

	if blockNumber < 0 {
		respStatus := response.CodeParamInvalid.WithMsg(", invalid blockNumber")
		return respStatus, errors.New(respStatus.Messages)
	}

	req.BlockNumber = int64(blockNumber)

	return response.Status{}, nil
}

func (req *TransGetTxHashReq) Parse(c *gin.Context) (response.Status, error) {
	req.TxHash = c.Param("txHash")
	if utils.IsEmpty(req.TxHash) {
		respStatus := response.CodeParamInvalid.WithMsg(", tx_hash is empty")
		return respStatus, errors.New(respStatus.Messages)
	}

	req.CryptoType = c.Query("crypto_type")
	if utils.IsEmpty(req.CryptoType) {
		respStatus := response.CodeParamInvalid.WithMsg(", crypto_type is empty")
		return respStatus, errors.New(respStatus.Messages)
	}

	req.ChainType = c.Query("chain_type")
	if req.ChainType != domain.ChainType {
		respStatus := response.CodeParamInvalid.WithMsg(", chain_type need use KASPAD")
		return respStatus, errors.New(respStatus.Messages)
	}

	return response.Status{}, nil
}

type TransGetTxHashResp struct {
	BlockHeight int64  `json:"block_height"`
	TxHash      string `json:"tx_hash"`

	CryptoType string `json:"crypto_type"`
	ChainType  string `json:"chain_type"`

	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`

	Amount    decimal.Decimal `json:"amount"`
	Fee       decimal.Decimal `json:"fee"`
	FeeCrypto string          `json:"fee_crypto"`

	Status int    `json:"status"`
	Memo   string `json:"memo"`
}
