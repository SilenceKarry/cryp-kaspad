package transaction

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/utils"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

type BlockHeightRepositoryCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type blockHeightRepository struct {
	db *gorm.DB
}

func NewBlockHeightRepository(cond BlockHeightRepositoryCond) (repository.BlockHeightRepository, error) {
	result := &blockHeightRepository{
		db: cond.DB,
	}

	return result, nil
}

func (repo *blockHeightRepository) New(db *gorm.DB) repository.BlockHeightRepository {
	result := &blockHeightRepository{
		db: db,
	}

	return result
}

func (repo *blockHeightRepository) Create(ctx context.Context, bh entity.BlockHeight) (int, error) {
	bh.CreateTime = utils.TimeNowUnix()
	bh.UpdateTime = utils.TimeNowUnix()

	result := repo.db.WithContext(ctx).Create(&bh)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(bh.ID), nil
}

func (repo *blockHeightRepository) Update(ctx context.Context, bh entity.BlockHeight) error {
	bh.UpdateTime = utils.TimeNowUnix()

	return repo.db.WithContext(ctx).Model(&bh).
		Where("id = ?", bh.ID).
		Updates(&bh).Error
}

func (repo *blockHeightRepository) Get(ctx context.Context) (entity.BlockHeight, error) {
	result := entity.BlockHeight{}
	err := repo.db.WithContext(ctx).
		Take(&result).Error
	if err != nil {
		return entity.BlockHeight{}, err
	}

	return result, nil
}
