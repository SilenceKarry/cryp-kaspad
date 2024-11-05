package address

import (
	"context"
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"fmt"
	"strings"

	//"github.com/ethereum/go-ethereum/common"

	"go.uber.org/dig"
)

type TokensUseCaseCond struct {
	dig.In

	TokensRepo repository.TokensRepository
}

type tokensUseCase struct {
	TokensUseCaseCond
}

func NewTokensUseCase(cond TokensUseCaseCond) usecase.TokensUseCase {
	uc := &tokensUseCase{
		TokensUseCaseCond: cond,
	}

	return uc
}

func (uc *tokensUseCase) GetByCryptoType(ctx context.Context, cryptoType string) (entity.Tokens, error) {
	return uc.TokensRepo.GetByCryptoType(ctx, cryptoType)
}

func (uc *tokensUseCase) GetList(ctx context.Context) ([]entity.Tokens, error) {
	return uc.TokensRepo.GetList(ctx)
}

// func (uc *tokensUseCase) GetContractAddress(ctx context.Context) ([]common.Address, error) {
// 	tokens, err := uc.TokensRepo.GetList(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("TokensRepo.GetContractAddress error: %s", err)
// 	}

// 	result := make([]common.Address, 0, len(tokens))

// 	for _, v := range tokens {
// 		if len(v.ContractAddr) > 0 {
// 			result = append(result, common.HexToAddress(v.ContractAddr))
// 		}
// 	}

// 	return result, nil
// }

func (uc *tokensUseCase) GetContractAddr2Tokens(ctx context.Context) (map[string]entity.Tokens, error) {
	tokens, err := uc.TokensRepo.GetList(ctx)
	if err != nil {
		return nil, fmt.Errorf("TokensRepo.GetContractAddress error: %s", err)
	}

	contractAddr2Token := make(map[string]entity.Tokens)

	for _, v := range tokens {
		if len(v.ContractAddr) > 0 {
			contractAddr2Token[strings.ToLower(v.ContractAddr)] = v
		}
	}

	return contractAddr2Token, nil
}

func (uc *tokensUseCase) CreateContractToken(ctx context.Context, request vo.ContractCreateRequest) error {
	token, err := uc.TokensRepo.GetByContractAddr(ctx, request.ContractAddr)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("CreateContractToken TokensRepo.GetByContractAddr error:%s", err)
	}

	if token.ID > 0 {
		return nil
	}

	now := time.Now().Unix()
	entityToken := entity.Tokens{
		CreateTime:   now,
		UpdateTime:   now,
		CryptoType:   request.CryptoType,
		ChainType:    domain.ChainType,
		ContractAddr: request.ContractAddr,
		Decimals:     request.Decimals,
		GasLimit:     request.GasLimit,
		GasPrice:     request.GasPrice,
	}

	jsonStr, isOk := covertJsonStr(request.ContractAbi)
	if !isOk {
		return errors.New("contract abi params error")
	}

	entityToken.ContractAbi = jsonStr
	if jsonStr == "" {
		entityToken.ContractAbi = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"burn","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"symbol","type":"string"},{"internalType":"uint8","name":"decimals","type":"uint8"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bool","name":"mintable","type":"bool"},{"internalType":"address","name":"owner","type":"address"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mint","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"mintable","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	}

	if err := uc.TokensRepo.Create(ctx, entityToken); err != nil {
		return fmt.Errorf("CreateContractToken TokensRepo.Create error:%s", err)
	}

	return nil
}

func (uc *tokensUseCase) UpdateContractToken(ctx context.Context, req vo.ContractUpdateRequest, contractAddr string) error {
	_, err := uc.TokensRepo.GetByContractAddr(ctx, contractAddr)
	if err != nil {
		return fmt.Errorf("UpdateContractToken TokensRepo.GetByContractAddr error:%w", err)
	}

	now := time.Now().Unix()
	entityToken := entity.Tokens{
		UpdateTime: now,
		CryptoType: req.CryptoType,
	}

	if req.GasPrice != nil {
		entityToken.GasPrice = *req.GasPrice
	}

	if req.Decimals != nil {
		entityToken.Decimals = *req.Decimals
	}

	if req.GasLimit != nil {
		entityToken.GasLimit = *req.GasLimit
	}

	if utils.IsNotEmpty(req.ContractAddr) {
		entityToken.ContractAddr = req.ContractAddr
	}

	jsonStr, isOk := covertJsonStr(req.ContractAbi)
	if !isOk {
		return errors.New("contract abi params error")
	}

	if jsonStr != "" {
		entityToken.ContractAbi = jsonStr
	}

	if err := uc.TokensRepo.Update(ctx, entityToken, contractAddr); err != nil {
		return fmt.Errorf("UpdateContractToken error:%s", err)
	}

	return nil
}

func (uc *tokensUseCase) GetByContractAddr(ctx context.Context, contractAddr string) (vo.GetContractResponse, response.Status, error) {
	token, err := uc.TokensRepo.GetByContractAddr(ctx, contractAddr)
	status := response.Status{}

	if err != nil {
		status = response.CodeInternalError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = response.CodeCryptoNotFound
		}
		return vo.GetContractResponse{}, status, errors.New(status.Messages)
	}

	var jsonObj any
	if token.ContractAbi != "" {
		if err := json.Unmarshal([]byte(token.ContractAbi), &jsonObj); err != nil {
			return vo.GetContractResponse{}, response.CodeInternalError, fmt.Errorf("GetByContractAddr json.Unmarshal error:%s", err)
		}
	}

	return vo.GetContractResponse{
		ContractAddr: token.ContractAddr,
		Decimals:     token.Decimals,
		GasLimit:     token.GasLimit,
		GasPrice:     token.GasPrice,
		CryptoType:   token.CryptoType,
		ChainType:    token.ChainType,
		ContractAbi:  jsonObj,
	}, status, nil
}

func covertJsonStr(data any) (string, bool) {
	if data == nil {
		return "", true
	}

	switch reflect.ValueOf(data).Kind() {
	// 這個例外
	case reflect.String:
		value := data.(string)
		if value == "" {
			return "", true
		}
		return "", false
	case reflect.Map:
		obj := data.(map[string]any)
		jsonByte, err := json.Marshal(&obj)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("covertJsonStr reflect.Map")
			return "", false
		}
		return string(jsonByte), true
	case reflect.Slice:
		obj := data.([]any)
		jsonByte, err := json.Marshal(&obj)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("covertJsonStr reflect.Array")
			return "", false
		}
		return string(jsonByte), true
	default:
		return "", false
	}
}
