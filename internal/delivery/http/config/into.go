package config

import (
	"cryp-kaspad/internal/libs/container"
	configRepository "cryp-kaspad/internal/repository/config"
	configUsecase "cryp-kaspad/internal/usecase/config"
)

func Init() error {
	if err := register(); err != nil {
		return err
	}

	if err := container.Get().Invoke(registerConfigRouter); err != nil {
		return err
	}

	return nil
}

func register() error {
	if err := container.Get().Provide(configRepository.NewConfigRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(configUsecase.NewConfigUseCase); err != nil {
		return err
	}

	return nil
}
