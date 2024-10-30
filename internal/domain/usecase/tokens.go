package usecase

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"

	"github.com/ethereum/go-ethereum/common"
)

type TokensUseCase interface {
	GetByCryptoType(ctx context.Context, cryptoType string) (entity.Tokens, error)

	GetList(ctx context.Context) ([]entity.Tokens, error)
	GetContractAddress(ctx context.Context) ([]common.Address, error)
	GetContractAddr2Tokens(ctx context.Context) (map[string]entity.Tokens, error)
	CreateContractToken(ctx context.Context, request vo.ContractCreateRequest) error
	UpdateContractToken(ctx context.Context, r vo.ContractUpdateRequest, contractAddr string) error
	GetByContractAddr(ctx context.Context, contractAddr string) (vo.GetContractResponse, response.Status, error)
}
