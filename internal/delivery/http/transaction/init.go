package transaction

import (
	"cryp-kaspad/internal/libs/container"
	transactionRepo "cryp-kaspad/internal/repository/transaction"
	transactionUseCase "cryp-kaspad/internal/usecase/transaction"
	"fmt"
)

func Init() error {
	if err := registerContainerWithdraw(); err != nil {
		return err
	}

	if err := registerContainerTransaction(); err != nil {
		return err
	}

	return nil
}

func registerContainerWithdraw() error {
	if err := container.Get().Provide(transactionRepo.NewWithdrawRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(transactionUseCase.NewWithdrawUseCase); err != nil {
		return err
	}

	if err := container.Get().Invoke(func(cond WithdrawHandlerCond) {
		registerRouterWithdraw(cond)
	}); err != nil {
		return fmt.Errorf("Invoke(registerRouterWithdraw), err: %w", err)
	}

	return nil
}

func registerContainerTransaction() error {
	if err := container.Get().Provide(transactionRepo.NewBlockHeightRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(transactionRepo.NewTransactionRepository); err != nil {
		return err
	}

	if err := container.Get().Provide(transactionUseCase.NewTransactionUseCase); err != nil {
		return err
	}

	if err := container.Get().Invoke(func(cond TransactionHandlerCond) {
		registerRouterTransaction(cond)
	}); err != nil {
		return fmt.Errorf("Invoke(registerRouterTransaction), err: %w", err)
	}

	return nil
}
