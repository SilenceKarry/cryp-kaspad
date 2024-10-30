package config

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"time"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

type ConfigCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(cond ConfigCond) repository.Config {
	return &configRepository{
		cond.DB,
	}
}

func (uc configRepository) Get(ctx context.Context, key string) (entity.Config, error) {
	model := entity.Config{}
	if err := uc.db.WithContext(ctx).Where("`key` = ?", key).First(&model).Error; err != nil {
		return model, err
	}

	return model, nil
}

func (uc configRepository) GetAll(ctx context.Context) ([]entity.Config, error) {
	model := make([]entity.Config, 0)
	if err := uc.db.WithContext(ctx).Find(&model).Error; err != nil {
		return model, err
	}

	return model, nil
}

func (uc configRepository) Create(ctx context.Context, key string, value string) error {
	now := time.Now().Unix()
	return uc.db.WithContext(ctx).Create(&entity.Config{
		Key:        key,
		Value:      value,
		CreateTime: now,
		UpdateTime: now,
	}).Error
}
func (uc configRepository) Update(ctx context.Context, model entity.Config) error {
	return uc.db.WithContext(ctx).Model(&model).Updates(&model).Error
}
