package address

import (
	"context"

	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"

	kaspa "cryp-kaspad/internal/libs/kaspa"
	"cryp-kaspad/internal/libs/response"
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/shopspring/decimal"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/util"
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

type configFlags struct {
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
func (uc *addressUseCase) Create(ctx context.Context, req vo.AddressCreateReq) (vo.AddressCreateResp, response.Status, error) {

	cfg, err := parseConfig()
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("parseConfig error: %s", err)
	}
	privateKey, publicKey, err := libkaspawallet.CreateKeyPair(false)
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("libkaspawallet.CreateKeyPair error: %s", err)
	}

	testFlag := uc.ConfigUseCase.ChkTestNet(ctx)
	if testFlag {
		cfg.Testnet = true
		cfg.NetParams().Prefix = util.Bech32PrefixKaspaTest
	}

	addr, err := util.NewAddressPublicKey(publicKey, cfg.NetParams().Prefix)
	if err != nil {
		return vo.AddressCreateResp{}, response.CodeInternalError, fmt.Errorf("NewAddressPublicKey error: %s", err)
	}
	privateStr := fmt.Sprintf("%x", privateKey)
	publicStr := fmt.Sprintf("%x", publicKey)

	result := vo.AddressCreateResp{
		Address:   addr.EncodeAddress(),
		SecretKey: string(privateStr),
		PublicKey: string(publicStr),
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

	c, err := kaspa.StartDeamon(ctx, urls)
	if err != nil {
		return vo.AddressGetBalanceResp{}, response.CodeInternalError, fmt.Errorf("StartDeamon error: %s", err)
	}

	resp, err := c.RpcClient.GetBalanceByAddress(req.Address)
	if err != nil {
		return vo.AddressGetBalanceResp{}, response.CodeInternalError, fmt.Errorf("kaspa.GetBalanceToken error: %s", err)
	}
	satoshis := decimal.NewFromInt(int64(resp.Balance))
	conversionRate := decimal.NewFromInt(1000000000)

	result := vo.AddressGetBalanceResp{
		Balance: satoshis.Div(conversionRate),
	}

	return result, response.Status{}, nil
}
