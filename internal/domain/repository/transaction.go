package repository

import (
	"context"
	"cryp-kaspad/internal/domain/entity"

	"gorm.io/gorm"
)

type BlockHeightRepository interface {
	New(db *gorm.DB) BlockHeightRepository

	Create(ctx context.Context, bh entity.BlockHeight) (int, error)
	Update(ctx context.Context, bh entity.BlockHeight) error

	Get(ctx context.Context) (entity.BlockHeight, error)
}

type TransactionRepository interface {
	New(db *gorm.DB) TransactionRepository

	Create(ctx context.Context, trans entity.Transaction) (int, error)
	Update(ctx context.Context, trans entity.Transaction) error

	GetByTxHash(ctx context.Context, txHash string) (entity.Transaction, error)
	GetListByStatusAndBlockHeight(ctx context.Context, status, limit int, blockHeight int64) ([]entity.Transaction, error)
	GetListByNotifyStatus(ctx context.Context, notifyStatus int) ([]entity.Transaction, error)
	GetListByRiskControlStatus(ctx context.Context, status int) ([]entity.Transaction, error)
	GetEthListSuccessByBlockHeight(ctx context.Context, height int64) ([]entity.Transaction, error)
}
