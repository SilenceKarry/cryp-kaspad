package repository

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
)

type TokensRepository interface {
	GetByCryptoType(ctx context.Context, cryptoType string) (entity.Tokens, error)

	GetList(ctx context.Context) ([]entity.Tokens, error)

	UpdateToken(ctx context.Context, model entity.Tokens) error
	Create(ctx context.Context, token entity.Tokens) error
	Update(ctx context.Context, token entity.Tokens, contractAddr string) error
	GetByContractAddr(ctx context.Context, contractAddr string) (entity.Tokens, error)
}
