package transaction

import (
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/libs/container"
	"fmt"

	"github.com/robfig/cron/v3"
)

func Init(c *cron.Cron, cronLog cron.Logger) error {
	if err := registerTransactionCron(c, cronLog); err != nil {
		return err
	}

	return nil
}

func registerTransactionCron(c *cron.Cron, cronLog cron.Logger) error {
	if err := container.Get().Invoke(func(uc usecase.TransactionUseCase) {

	}); err != nil {
		return fmt.Errorf("Invoke(registerTransactionCron), err: %w", err)
	}

	return nil
}
