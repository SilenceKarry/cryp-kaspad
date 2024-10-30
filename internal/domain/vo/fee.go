package vo

import (
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type FeeReq struct {
	CryptoType string
}

func (req *FeeReq) Parse(c *gin.Context) (response.Status, error) {
	cryptoType := c.Param("crypto_type")
	respStatus := response.Status{}
	if utils.IsEmpty(cryptoType) {
		respStatus = response.CodeParamInvalid.WithMsg(", crypto_type is empty")
		return respStatus, errors.New(respStatus.Messages)
	}

	req.CryptoType = cryptoType
	return respStatus, nil
}

type FeeReqResp struct {
	CryptoType string          `json:"crypto_type"`
	Fee        decimal.Decimal `json:"fee"`
}
