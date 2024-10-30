package repository

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
)

type Config interface {
	Get(ctx context.Context, key string) (entity.Config, error)
	GetAll(ctx context.Context) ([]entity.Config, error)
	Create(ctx context.Context, key string, value string) error
	Update(ctx context.Context, model entity.Config) error
}
