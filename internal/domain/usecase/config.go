package usecase

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain/entity"
)

type ConfigUseCase interface {
	SetOrCreate(ctx context.Context, key string, value string) error
	GetAll(ctx context.Context) ([]entity.Config, error)
	GetWithKeys(ctx context.Context, keys ...configs.ViperKey) (map[configs.ViperKey]string, error)
	GetWithVKey(ctx context.Context, vkey configs.ViperKey) (interface{}, error)
	GetClientConfig(ctx context.Context) ([]string, string, string)
	GetNodeUrl(ctx context.Context) []string
	// Sync(ctx context.Context) error
}
