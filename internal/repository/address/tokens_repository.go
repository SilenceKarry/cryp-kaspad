package address

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/utils"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type TokensRepositoryCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type tokensRepository struct {
	db *gorm.DB
}

func NewTokensRepository(cond TokensRepositoryCond) (repository.TokensRepository, error) {
	result := &tokensRepository{
		db: cond.DB,
	}

	return result, nil
}

func (repo *tokensRepository) GetByCryptoType(ctx context.Context, cryptoType string) (entity.Tokens, error) {
	result := entity.Tokens{}
	err := repo.db.WithContext(ctx).
		Where("`crypto_type` = ?", cryptoType).
		Take(&result).Error
	if err != nil {
		return entity.Tokens{}, err
	}

	return result, nil
}

func (repo *tokensRepository) GetList(ctx context.Context) ([]entity.Tokens, error) {
	result := make([]entity.Tokens, 0, 200)
	err := repo.db.WithContext(ctx).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *tokensRepository) UpdateToken(ctx context.Context, model entity.Tokens) error {
	model.UpdateTime = utils.TimeNowUnix()
	err := repo.db.WithContext(ctx).
		Model(&model).
		Where("crypto_type = ?", model.CryptoType).
		Updates(&model).Error

	if err != nil {
		return err
	}

	return nil
}

func (repo *tokensRepository) Create(ctx context.Context, token entity.Tokens) error {
	err := repo.db.WithContext(ctx).Create(&token).Error
	if err != nil {
		return err
	}

	return nil
}

func (repo *tokensRepository) Update(ctx context.Context, token entity.Tokens, contractAddr string) error {
	err := repo.db.WithContext(ctx).Model(&token).Where("contract_addr = ?", contractAddr).Updates(&token).Error
	if err != nil {
		return err
	}

	return nil
}

func (repo *tokensRepository) GetByContractAddr(ctx context.Context, contractAddr string) (entity.Tokens, error) {
	token := entity.Tokens{}
	err := repo.db.WithContext(ctx).Where("contract_addr = ?", contractAddr).Take(&token).Error
	if err != nil {
		return token, err
	}

	return token, nil
}
