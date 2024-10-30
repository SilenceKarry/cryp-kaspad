package repository

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"gorm.io/gorm"
)

type WithdrawRepository interface {
	Create(ctx context.Context, w entity.Withdraw) (int, error)

	GetByTxHash(ctx context.Context, txHash string) (entity.Withdraw, error)

	GetByAddrAndNonce(ctx context.Context, addr string, nonce uint64) (entity.Withdraw, error)

	GetLastNonceWithdraw(ctx context.Context, addr string) (entity.Withdraw, error)

	New(db *gorm.DB) WithdrawRepository

	Update(ctx context.Context, withdraw entity.Withdraw) error

	DeleteWithdraw(ctx context.Context, id int64) error

	IsRepeatByAddrAndNonce(ctx context.Context, fromAddr string, nonce int64) (bool, error)

	GetByHasRetryAndHasChain(ctx context.Context, hasRetry, hasChain int64) ([]entity.Withdraw, error)
}
