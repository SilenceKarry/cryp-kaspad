package address

import (
	"cryp-kaspad/internal/libs/container"
	"fmt"

	addressRepo "cryp-kaspad/internal/repository/address"
	addressUseCase "cryp-kaspad/internal/usecase/address"
)

func Init() error {
	if err := registerContainerTokens(); err != nil {
		return err
	}

	if err := registerContainerAddress(); err != nil {
		return err
	}

	return nil
}

func registerContainerTokens() error {
	if err := container.Get().Provide(addressRepo.NewTokensRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(addressUseCase.NewTokensUseCase); err != nil {
		return err
	}

	return nil
}

func registerContainerAddress() error {
	if err := container.Get().Provide(addressRepo.NewAddressRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(addressUseCase.NewAddressUseCase); err != nil {
		return err
	}

	if err := container.Get().Invoke(func(cond AddressHandlerCond) {
		registerRouterAddress(cond)
	}); err != nil {
		return fmt.Errorf("Invoke(registerRouterAddress), err: %w", err)
	}

	return nil
}
