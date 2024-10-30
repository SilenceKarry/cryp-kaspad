package monitor

import (
	"cryp-kaspad/internal/libs/container"
	"fmt"
)

func Init() error {
	if err := registerContainerMonitor(); err != nil {
		return err
	}

	return nil
}

func registerContainerMonitor() error {
	if err := container.Get().Invoke(func(cond MonitorHandlerCond) {
		registerRouterMonitor(cond)
	}); err != nil {
		return fmt.Errorf("Invoke(registerRouterMonitor), err: %w", err)
	}

	return nil
}
