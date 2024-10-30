package transaction

import (
	"context"
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/utils"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

type TransactionRepositoryCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(cond TransactionRepositoryCond) (repository.TransactionRepository, error) {
	result := &transactionRepository{
		db: cond.DB,
	}

	return result, nil
}

func (repo *transactionRepository) New(db *gorm.DB) repository.TransactionRepository {
	result := &transactionRepository{
		db: db,
	}

	return result
}

func (repo *transactionRepository) Create(ctx context.Context, trans entity.Transaction) (int, error) {
	trans.CreateTime = utils.TimeNowUnix()
	trans.UpdateTime = utils.TimeNowUnix()

	result := repo.db.WithContext(ctx).Create(&trans)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(trans.ID), nil
}

func (repo *transactionRepository) Update(ctx context.Context, trans entity.Transaction) error {
	trans.UpdateTime = utils.TimeNowUnix()

	return repo.db.WithContext(ctx).Model(&trans).
		Where("id = ?", trans.ID).
		Updates(&trans).Error
}

func (repo *transactionRepository) GetByTxHash(ctx context.Context, txHash string) (entity.Transaction, error) {
	result := entity.Transaction{}
	err := repo.db.WithContext(ctx).
		Where("`tx_hash` = ?", txHash).
		Take(&result).Error
	if err != nil {
		return entity.Transaction{}, err
	}

	return result, nil
}

func (repo *transactionRepository) GetListByStatusAndBlockHeight(ctx context.Context, status, limit int, blockHeight int64) ([]entity.Transaction, error) {
	result := make([]entity.Transaction, 0, 0)
	err := repo.db.WithContext(ctx).
		Where("`status` = ?", status).
		Where("`block_height` <= ?", blockHeight).
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *transactionRepository) GetListByNotifyStatus(ctx context.Context, notifyStatus int) ([]entity.Transaction, error) {
	result := make([]entity.Transaction, 0, 0)
	err := repo.db.WithContext(ctx).
		Where("`notify_status` = ?", notifyStatus).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *transactionRepository) GetListByRiskControlStatus(ctx context.Context, status int) ([]entity.Transaction, error) {
	result := make([]entity.Transaction, 0, 0)
	err := repo.db.WithContext(ctx).
		Where("`risk_control_status` = ?", status).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *transactionRepository) GetEthListSuccessByBlockHeight(ctx context.Context, height int64) ([]entity.Transaction, error) {
	trans := make([]entity.Transaction, 0)

	err := repo.db.WithContext(ctx).
		Where("block_height = ?", height).
		Where("crypto_type = ?", domain.CryptoType).
		Where("status = ?", domain.TxStatusSuccess).
		Find(&trans).Error

	if err != nil {
		return nil, err
	}

	return trans, nil
}
