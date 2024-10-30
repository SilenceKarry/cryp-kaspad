package repository

import (
	"context"
	"cryp-kaspad/internal/domain/entity"

	"gorm.io/gorm"
)

type AddressRepository interface {
	New(db *gorm.DB) AddressRepository
	Create(ctx context.Context, addr entity.Address) (int, error)

	GetByAddress(ctx context.Context, addr string) (entity.Address, error)

	GetByAddrWithLock(ctx context.Context, addr string) (entity.Address, error)

	Update(ctx context.Context, address entity.Address) error

	GetList(ctx context.Context) ([]entity.Address, error)
}
