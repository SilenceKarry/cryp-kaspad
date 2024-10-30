package address

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/eos"
	"cryp-kaspad/internal/libs/eos/client"
	"cryp-kaspad/internal/libs/response"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"go.uber.org/dig"
)

type AddressUseCaseCond struct {
	dig.In

	AddressRepo repository.AddressRepository

	TokensUseCase usecase.TokensUseCase

	ConfigUseCase usecase.ConfigUseCase
}

type addressUseCase struct {
	AddressUseCaseCond
}

func NewAddressUseCase(cond AddressUseCaseCond) usecase.AddressUseCase {
	uc := &addressUseCase{
		AddressUseCaseCond: cond,
	}

	return uc
}

func (uc *addressUseCase) Create(ctx context.Context, req vo.AddressCreateReq) (vo.AddressCreateResp, response.Status, error) {
	//配合pool只回傳固定錢包
	wd := configs.App.GetWalletDesposit()
	return vo.AddressCreateResp{
		Address:   wd.AccountID,
		SecretKey: wd.PrivateKey,
		PublicKey: wd.PublicKey,
	}, response.Status{}, nil

	urls, privKey, accountID := uc.ConfigUseCase.GetClientConfig(ctx)

	creatorConfig := client.CreatorConfig{
		PrivateKey: privKey,
		AccountID:  accountID,
	}

	client, err := eos.NewClient(ctx, urls)
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.NewClient error: %s", err)
	}

	newAccount, err := tryCreateNewAccountId(ctx, client)
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("tryCreateNewAccountId error: %s", err)
	}

	newPubKey, newPrivKey, err := eos.GenerateKeyPair()
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.GenerateKeyPair error: %s", err)
	}

	createSuccess := false
	retry := configs.App.GetRetry()
	for i := 0; i < retry; i++ {
		err = client.CreateAccount(ctx, newAccount, newPubKey, creatorConfig)
		if err != nil {
			log.WithFields(log.Fields{
				"newAccount": newAccount,
				"newPubKey":  newPubKey,
				"error":      err,
			}).Error("client.CreateAccount error")

			if strings.Contains(err.Error(), "overdrawn balance") {
				return vo.AddressCreateResp{}, response.CodeBalanceInsufficient, errors.New(response.CodeBalanceInsufficient.Messages)
			}

			if i+1 < retry {
				canRetry, err := eos.IsCPULimitErrorFalsePositive(err)
				if err != nil {
					log.WithFields(log.Fields{
						"newAccount": newAccount,
						"newPubKey":  newPubKey,
						"error":      err,
					}).Error("eos.IsCPULimitErrorFalsePositive error")

					continue
				}

				if !canRetry {
					break
				}
			}
		}

		createSuccess = true
		break
	}

	if !createSuccess {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("eos.CreateAccount error: %s, newAccount: %s, newPubKey: %s", err, newAccount, newPubKey)
	}

	addr := entity.Address{
		MerchantType: domain.MerchantID2Type[req.MerchantID],
		Address:      newAccount,
		ChainType:    req.ChainType,
	}

	_, err = uc.AddressRepo.Create(ctx, addr)
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("AddressRepo.Create error: %s", err)
	}

	result := vo.AddressCreateResp{
		Address:   newAccount,
		SecretKey: newPrivKey,
		PublicKey: newPubKey,
	}

	return result, response.Status{}, nil
}

func (uc *addressUseCase) GetByAddress(ctx context.Context, addr string) (entity.Address, error) {
	return uc.AddressRepo.GetByAddress(ctx, addr)
}

func (uc *addressUseCase) GetList(ctx context.Context) ([]entity.Address, error) {
	return uc.AddressRepo.GetList(ctx)
}

func (uc *addressUseCase) GetBalance(ctx context.Context, req vo.AddressGetBalanceReq) (vo.AddressGetBalanceResp, response.Status, error) {
	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	client, err := eos.NewClient(ctx, urls)
	if err != nil {
		return vo.AddressGetBalanceResp{}, response.CodeInternalError, fmt.Errorf("eos.NewClient error: %s", err)
	}

	tokens, err := uc.TokensUseCase.GetByCryptoType(ctx, req.CryptoType)
	if err != nil && err != gorm.ErrRecordNotFound {
		return vo.AddressGetBalanceResp{}, response.CodeInternalError, fmt.Errorf("TokensUseCase.GetByCryptoType error: %s", err)
	}

	if err == gorm.ErrRecordNotFound {
		return vo.AddressGetBalanceResp{}, response.CodeCryptoNotFound, errors.New(response.CodeCryptoNotFound.Messages)
	}

	tokenBalance, err := client.GetBalanceToken(ctx, req.Address, tokens.ContractAddr, req.CryptoType)
	if err != nil {
		return vo.AddressGetBalanceResp{}, response.CodeInternalError, fmt.Errorf("eos.GetBalanceToken error: %s", err)
	}

	result := vo.AddressGetBalanceResp{
		Balance: tokenBalance,
	}

	return result, response.Status{}, nil
}

func tryCreateNewAccountId(ctx context.Context, client *eos.Eos) (string, error) {
	for i := 0; i < 10; i++ {
		newAccount := eos.RandomEOSName()

		exist, err := client.ExistAccount(ctx, newAccount)
		if err != nil {
			return "", fmt.Errorf("eos.ExistxAccount error: %s, newAccount: %s", err, newAccount)
		}

		if exist {
			continue
		}

		return newAccount, nil
	}
	return "", fmt.Errorf("create new account failed")
}
