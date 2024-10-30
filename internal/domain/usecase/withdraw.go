package usecase

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
)

type WithdrawUseCase interface {
	Create(ctx context.Context, req vo.WithdrawCreateReq) (vo.WithdrawCreateResp, response.Status, error)

	GetByTxHash(ctx context.Context, txHash string) (entity.Withdraw, error)
}
