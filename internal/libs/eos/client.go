package eos

import (
	"context"
	"fmt"
	"time"

	eosclient "cryp-kaspad/internal/libs/eos/client"

	"github.com/ethereum/go-ethereum/log"
)

type Eos struct {
	Client *eosclient.Client
}

func NewClient(ctx context.Context, urls []string) (*Eos, error) {
	for _, url := range urls {
		client, err := eosclient.Dial(url)
		if err != nil {
			log.Warn("eosclient.Dial failed", "url", url, "error", err)
			continue
		}

		if !ping(client) {
			log.Warn("eosclient.Dial ping failed", "url", url)
			continue
		}

		return &Eos{Client: client}, nil
	}

	return nil, fmt.Errorf("eosclient.Dial all url failed, urls: %v", urls)
}

func ping(client *eosclient.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetInfo(ctx)
	return err == nil
}
