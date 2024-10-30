package tokens

import (
	"cryp-kaspad/internal/libs/container"
)

func Init() error {
	if err := container.Get().Invoke(registerRouterMonitor); err != nil {
		return err
	}

	return nil
}
