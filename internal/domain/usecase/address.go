package usecase

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
)

type AddressUseCase interface {
	Create(ctx context.Context, req vo.AddressCreateReq) (vo.AddressCreateResp, response.Status, error)

	GetByAddress(ctx context.Context, addr string) (entity.Address, error)

	GetBalance(ctx context.Context, req vo.AddressGetBalanceReq) (vo.AddressGetBalanceResp, response.Status, error)

	GetList(ctx context.Context) ([]entity.Address, error)
}
