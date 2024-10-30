package cron

import (
	"cryp-kaspad/internal/libs/container"
	"fmt"
)

func Init() error {
	if err := registerCronContrainer(); err != nil {
		return err
	}

	return nil
}

func registerCronContrainer() error {
	if err := container.Get().Invoke(func(cond CronHandlerCond) {
		registerRouterCronHandler(cond)
	}); err != nil {
		return fmt.Errorf("Invoke(registerRouterCronHandler), err: %w", err)
	}

	return nil
}
