package usecase

import (
	"context"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
)

type FeeUseCase interface {
	GetFee(ctx context.Context, req vo.FeeReq) (vo.FeeReqResp, response.Status, error)
}
