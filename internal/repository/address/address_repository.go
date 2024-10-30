package address

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/utils"

	"gorm.io/gorm/clause"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

type AddressRepositoryCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(cond AddressRepositoryCond) (repository.AddressRepository, error) {
	result := &addressRepository{
		db: cond.DB,
	}

	return result, nil
}

func (repo *addressRepository) New(db *gorm.DB) repository.AddressRepository {
	return &addressRepository{
		db: db,
	}
}

func (repo *addressRepository) Create(ctx context.Context, addr entity.Address) (int, error) {
	addr.CreateTime = utils.TimeNowUnix()
	addr.UpdateTime = utils.TimeNowUnix()

	result := repo.db.WithContext(ctx).Create(&addr)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(addr.ID), nil
}

func (repo *addressRepository) GetByAddress(ctx context.Context, addr string) (entity.Address, error) {
	result := entity.Address{}
	err := repo.db.WithContext(ctx).
		Where("`address` = ?", addr).
		Take(&result).Error
	if err != nil {
		return entity.Address{}, err
	}

	return result, nil
}

func (repo *addressRepository) GetByAddrWithLock(ctx context.Context, addr string) (entity.Address, error) {
	result := entity.Address{}
	err := repo.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("address = ?", addr).First(&result).Error

	return result, err
}

func (repo *addressRepository) Update(ctx context.Context, address entity.Address) error {
	return repo.db.WithContext(ctx).Model(&address).Updates(&address).Error
}

func (repo *addressRepository) GetList(ctx context.Context) ([]entity.Address, error) {
	result := make([]entity.Address, 0, 200)
	err := repo.db.WithContext(ctx).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
