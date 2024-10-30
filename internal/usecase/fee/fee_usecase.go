package fee

import (
	"context"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type FeeUseCaseCond struct {
	dig.In

	TokensUseCase usecase.TokensUseCase
}

type feeUseCase struct {
	FeeUseCaseCond
}

func NewFeeUseCase(cond FeeUseCaseCond) usecase.FeeUseCase {
	return &feeUseCase{
		FeeUseCaseCond: cond,
	}
}

func (t *feeUseCase) GetFee(ctx context.Context, req vo.FeeReq) (vo.FeeReqResp, response.Status, error) {
	feeResp := vo.FeeReqResp{}
	status := response.Status{}

	token, err := t.TokensUseCase.GetByCryptoType(ctx, req.CryptoType)
	if err != nil {
		status = response.CodeCryptoNotFound
		errMsg := errors.New(status.Messages)

		if err != gorm.ErrRecordNotFound {
			status = response.CodeInternalError
			errMsg = fmt.Errorf("TokensUseCase.GetByCryptoType (%s)", err)
		}
		return feeResp, status, errMsg
	}

	feeResp = vo.FeeReqResp{
		CryptoType: token.CryptoType,
		Fee:        token.TransactionFee,
	}

	feeResp.Fee = token.GasPrice.
		Mul(decimal.NewFromInt(token.GasLimit)).
		Mul(decimal.New(1, int32(-18))).
		RoundBank(8)

	return feeResp, status, nil
}
