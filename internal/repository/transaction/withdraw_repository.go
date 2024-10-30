package transaction

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/utils"
	"time"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

type WithdrawRepositoryCond struct {
	dig.In

	DB *gorm.DB `name:"dbM"`
}

type withdrawRepository struct {
	db *gorm.DB
}

func NewWithdrawRepository(cond WithdrawRepositoryCond) (repository.WithdrawRepository, error) {
	result := &withdrawRepository{
		db: cond.DB,
	}

	return result, nil
}

func (repo *withdrawRepository) Create(ctx context.Context, w entity.Withdraw) (int, error) {
	w.CreateTime = utils.TimeNowUnix()
	w.UpdateTime = utils.TimeNowUnix()

	result := repo.db.WithContext(ctx).Create(&w)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(w.ID), nil
}

func (repo *withdrawRepository) GetByTxHash(ctx context.Context, txHash string) (entity.Withdraw, error) {
	result := entity.Withdraw{}
	err := repo.db.WithContext(ctx).
		Where("`tx_hash` = ?", txHash).
		Take(&result).Error
	if err != nil {
		return entity.Withdraw{}, err
	}

	return result, nil
}

func (repo *withdrawRepository) GetByAddrAndNonce(ctx context.Context, addr string, nonce uint64) (entity.Withdraw, error) {
	result := entity.Withdraw{}
	err := repo.db.WithContext(ctx).Where("from_address = ?", addr).Where("nonce = ?", nonce).Take(&result).Error
	if err != nil {
		return result, err
	}

	return result, nil
}

func (repo *withdrawRepository) GetLastNonceWithdraw(ctx context.Context, addr string) (entity.Withdraw, error) {
	result := entity.Withdraw{}
	err := repo.db.WithContext(ctx).Where("from_address = ?", addr).Order("nonce desc").Take(&result).Error
	if err != nil {
		return result, err
	}

	return result, nil
}

func (repo *withdrawRepository) Update(ctx context.Context, withdraw entity.Withdraw) error {

	err := repo.db.WithContext(ctx).Model(&withdraw).Updates(&withdraw).Error
	if err != nil {
		return err
	}

	return nil
}

func (repo *withdrawRepository) DeleteWithdraw(ctx context.Context, id int64) error {
	err := repo.db.WithContext(ctx).Delete(&entity.Withdraw{ID: id}).Error
	if err != nil {
		return err
	}

	return nil
}

func (repo *withdrawRepository) IsRepeatByAddrAndNonce(ctx context.Context, fromAddr string, nonce int64) (bool, error) {
	count := int64(0)
	err := repo.db.WithContext(ctx).Where("from_address = ?", fromAddr).Where("nonce = ?", nonce).Count(&count).Error

	return count > 0, err
}

func (repo *withdrawRepository) GetByHasRetryAndHasChain(ctx context.Context, hasRetry, hasChain int64) ([]entity.Withdraw, error) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	start := time.Date(oneWeekAgo.Year(), oneWeekAgo.Month(), oneWeekAgo.Day(), 0, 0, 0, 0, oneWeekAgo.Location()).Unix()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()

	withdraws := make([]entity.Withdraw, 0, 0)
	err := repo.db.
		WithContext(ctx).
		Where("has_retried = ?", hasRetry).
		Where("has_chain = ?", hasChain).
		Where("create_time BETWEEN ? AND ?", start, end).
		Find(&withdraws).Error

	if err != nil {
		return nil, err
	}

	return withdraws, err
}

func (repo *withdrawRepository) New(db *gorm.DB) repository.WithdrawRepository {
	return &withdrawRepository{
		db: db,
	}
}
