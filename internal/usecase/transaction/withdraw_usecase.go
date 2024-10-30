package transaction

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/eos"
	"cryp-kaspad/internal/libs/response"
	"errors"
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"go.uber.org/dig"
)

type WithdrawUseCaseCond struct {
	dig.In

	WithdrawRepo repository.WithdrawRepository

	TokensUseCase  usecase.TokensUseCase
	AddressUseCase usecase.AddressUseCase
	ConfigUseCase  usecase.ConfigUseCase
	AddressRepo    repository.AddressRepository
	TokenRepo      repository.TokensRepository
	DB             *gorm.DB `name:"dbM"`
}

type withdrawUseCase struct {
	WithdrawUseCaseCond
}

var (
	cpuLimitExp = regexp.MustCompile(`subjective cpu of \((\d+) us\).*?account cpu limit (\d+)us`)
)

func NewWithdrawUseCase(cond WithdrawUseCaseCond) usecase.WithdrawUseCase {
	uc := &withdrawUseCase{
		WithdrawUseCaseCond: cond,
	}

	return uc
}

func (uc *withdrawUseCase) Create(ctx context.Context, req vo.WithdrawCreateReq) (vo.WithdrawCreateResp, response.Status, error) {
	tokens, err := uc.TokensUseCase.GetByCryptoType(ctx, req.CryptoType)
	if err == gorm.ErrRecordNotFound {
		return vo.WithdrawCreateResp{}, response.CodeCryptoNotFound, errors.New(response.CodeCryptoNotFound.Messages)
	}

	if err != nil {
		return vo.WithdrawCreateResp{}, response.CodeInternalError, fmt.Errorf("TokensUseCase.GetByCryptoType error: %s", err)
	}

	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	client, err := eos.NewClient(ctx, urls)
	if err != nil {
		return vo.WithdrawCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.NewClient error: %s", err)
	}

	balance, err := client.GetBalanceToken(ctx, req.FromAddress, tokens.ContractAddr, tokens.CryptoType)
	if err != nil {
		return vo.WithdrawCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.GetBalanceToken error: %s", err)
	}

	if balance.Cmp(req.Amount) < 0 {
		return vo.WithdrawCreateResp{}, response.CodeBalanceInsufficient, errors.New(response.CodeBalanceInsufficient.Messages)
	}
	var txHash string
	retry := configs.App.GetRetry()
	for i := 0; i < retry; i++ {
		h, err := client.SendTransaction(ctx, req.FromAddress, req.ToAddress, req.SecretKey, req.Amount, tokens, req.Memo)
		if err != nil {
			log.WithFields(log.Fields{
				"from_address": req.FromAddress,
				"to_address":   req.ToAddress,
				"amount":       req.Amount,
				"error":        err,
			}).Error("client.SendTransaction error")

			if i+1 < retry {
				canRetry, err := eos.IsCPULimitErrorFalsePositive(err)
				if err != nil {
					log.WithFields(log.Fields{
						"from_address": req.FromAddress,
						"to_address":   req.ToAddress,
						"amount":       req.Amount,
						"error":        err,
					}).Error("eos.IsCPULimitErrorFalsePositive error")

					continue
				}

				if !canRetry {
					break
				}

				log.Info("retry")
			}
		}

		txHash = h
		break
	}

	if txHash == "" {
		return vo.WithdrawCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.SendTransaction failed")
	}

	if _, err := uc.create(ctx, req, txHash); err != nil {
		return vo.WithdrawCreateResp{}, response.CodeInternalError, fmt.Errorf("uc.create tx:%s err:%s", txHash, err)
	}

	result := vo.WithdrawCreateResp{
		TxHash: txHash,
		Memo:   req.Memo,
	}

	log.WithFields(log.Fields{
		"from_address": req.FromAddress,
		"to_address":   req.ToAddress,
		"tx_hash":      txHash,
	}).Warn("withdraw info")

	return result, response.Status{}, nil
}

func (uc *withdrawUseCase) create(ctx context.Context, req vo.WithdrawCreateReq, hash string) (int64, error) {
	withdraw := entity.Withdraw{
		MerchantType: domain.MerchantID2Type[req.MerchantID],
		CryptoType:   req.CryptoType,
		ChainType:    req.ChainType,
		FromAddress:  req.FromAddress,
		ToAddress:    req.ToAddress,
		Amount:       req.Amount,
		Memo:         req.Memo,
		TxHash:       hash,
	}
	id, err := uc.WithdrawRepo.Create(ctx, withdraw)
	if err != nil {
		return 0, fmt.Errorf("WithdrawRepo.Create error: %s", err)
	}

	return int64(id), nil
}

func (uc *withdrawUseCase) GetByTxHash(ctx context.Context, txHash string) (entity.Withdraw, error) {
	return uc.WithdrawRepo.GetByTxHash(ctx, txHash)
}
