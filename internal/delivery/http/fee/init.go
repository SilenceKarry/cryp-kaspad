package fee

import (
	"cryp-kaspad/internal/libs/container"
	"cryp-kaspad/internal/usecase/fee"
	"fmt"
)

func Init() error {
	if err := registerRouter(); err != nil {
		return err
	}

	return nil
}

func registerRouter() error {
	if err := container.Get().Provide(fee.NewFeeUseCase); err != nil {
		return err
	}

	if err := container.Get().Invoke(registerRouterFee); err != nil {
		return fmt.Errorf("Invoke(registerRouterFee), err: %w", err)
	}

	return nil
}
