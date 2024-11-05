package config

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"fmt"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/config"
	_ "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"gorm.io/gorm"

	_ "github.com/jessevdk/go-flags"
)

type ConfigCond struct {
	dig.In

	ConfigRepo repository.Config
}

type configUseCase struct {
	ConfigCond
}
type BalanceConf struct {
	DaemonAddress string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to"`
	Verbose       bool   `long:"verbose" short:"v" description:"Verbose: show addresses with balance"`
	config.NetworkFlags
}

func NewConfigUseCase(cond ConfigCond) usecase.ConfigUseCase {
	return &configUseCase{
		cond,
	}
}

func (uc *configUseCase) SetOrCreate(ctx context.Context, key string, value string) error {
	if IsWhitelistedKey := configs.App.IsWhitelistedKey(key); !IsWhitelistedKey {
		return fmt.Errorf("%v is not whitelisted", key)
	}

	config, err := uc.ConfigRepo.Get(ctx, key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := uc.ConfigRepo.Create(ctx, key, value); err != nil {
				return fmt.Errorf("SetOrCreate ConfigRepo.Create(%s)", err)
			}

			return nil
		}

		return fmt.Errorf("SetOrCreate ConfigRepo.Get(%s)", err)
	}

	config.Key = key
	config.Value = value
	config.UpdateTime = time.Now().Unix()
	if err := uc.ConfigRepo.Update(ctx, config); err != nil {
		return fmt.Errorf("SetOrCreate ConfigRepo.Update (%s)", err)
	}

	return nil
}
func (uc *configUseCase) GetAll(ctx context.Context) ([]entity.Config, error) {
	configs, err := uc.ConfigRepo.GetAll(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return configs, nil
}

func (uc *configUseCase) GetWithKeys(ctx context.Context, keys ...configs.ViperKey) (map[configs.ViperKey]string, error) {
	config, err := uc.ConfigRepo.GetAll(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	tempMap := make(map[string]string)
	for _, c := range config {
		tempMap[c.Key] = c.Value
	}

	configMap := make(map[configs.ViperKey]string)
	for _, k := range keys {
		nodeKey, ok := configs.App.TryGetKey(k)
		if !ok {
			return nil, fmt.Errorf("try use not whitelisted key")
		}
		v, ok := tempMap[nodeKey]
		if !ok {
			return nil, fmt.Errorf("db not found for key %v", nodeKey)
		}
		configMap[k] = v
	}

	return configMap, nil
}

func (uc *configUseCase) GetWithVKey(ctx context.Context, vkey configs.ViperKey) (interface{}, error) {
	key, ok := configs.App.TryGetKey(vkey)
	if !ok {
		return nil, fmt.Errorf("try use not whitelisted key")
	}
	value, err := configs.App.GetValueWithKey(key)
	if err != nil {
		return nil, err
	}
	configSlice, err := uc.ConfigRepo.GetAll(ctx)
	if err != nil {
		return value, nil
	}
	for _, v := range configSlice {
		if v.Key == key {
			switch vkey {

			default:
				return v.Value, nil
			}
		}
	}
	return value, nil
}

func (uc *configUseCase) tryGetKeys(ctx context.Context, keys ...configs.ViperKey) (map[configs.ViperKey]interface{}, error) {
	config, err := uc.ConfigRepo.GetAll(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	tempMap := make(map[string]string)
	for _, c := range config {
		tempMap[c.Key] = c.Value
	}

	configMap := make(map[configs.ViperKey]interface{})
	for _, k := range keys {
		nodeKey, ok := configs.App.TryGetKey(k)
		if !ok {
			return nil, fmt.Errorf("try use not whitelisted key")
		}
		v, ok := tempMap[nodeKey]
		if !ok {
			continue
		}
		configMap[k] = v
	}

	return configMap, nil
}

func (uc *configUseCase) GetDBConfig(ctx context.Context, defaultMap map[configs.ViperKey]interface{}) map[configs.ViperKey]interface{} {
	var keys []configs.ViperKey
	for k := range defaultMap {
		keys = append(keys, k)
	}

	//先在DB搜尋
	configMap, err := uc.tryGetKeys(ctx, keys...)
	if err != nil {
		logrus.Errorf("GetDBConfig GetWithKeys error: %v", err)
	}

	if configMap == nil {
		configMap = map[configs.ViperKey]interface{}{}
	}

	//若DB沒有，則使用預設值
	for _, k := range keys {
		if _, ok := configMap[k]; !ok {
			configMap[k] = defaultMap[k]
		}
	}

	return configMap
}

func (uc *configUseCase) GetClientConfig(ctx context.Context) ([]string, string, string) {
	defaultMap := map[configs.ViperKey]interface{}{
		configs.ViperNodeURL:    configs.App.GetNodeURL(),
		configs.ViperPrivateKey: configs.App.GetPrivateKey(),
		configs.ViperAccountID:  configs.App.GetAccountID(),
	}

	configMap := uc.GetDBConfig(ctx, defaultMap)

	if _, ok := configMap[configs.ViperNodeURL].([]string); !ok {
		configMap[configs.ViperNodeURL] = []string{configMap[configs.ViperNodeURL].(string)}
	}

	return configMap[configs.ViperNodeURL].([]string), configMap[configs.ViperPrivateKey].(string), configMap[configs.ViperAccountID].(string)
}

func (uc *configUseCase) GetNodeUrl(ctx context.Context) []string {
	defaultMap := map[configs.ViperKey]interface{}{
		configs.ViperNodeURL: configs.App.GetNodeURL(),
	}

	configMap := uc.GetDBConfig(ctx, defaultMap)

	if _, ok := configMap[configs.ViperNodeURL].([]string); !ok {
		configMap[configs.ViperNodeURL] = []string{configMap[configs.ViperNodeURL].(string)}
	}

	return configMap[configs.ViperNodeURL].([]string)
}

func (uc *configUseCase) ChkTestNet(ctx context.Context) bool {
	testFlag := configs.App.ChkTestNet()
	return testFlag
}
