package cron

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	transactionDeliveryCron "cryp-kaspad/internal/delivery/cron/transaction"
)

func Init() error {
	cronLog := cron.VerbosePrintfLogger(log.StandardLogger())

	// corn 全域的設定，每個加入的 job 都會有的設定，cron.Recover 處理 panic 場景
	c := cron.New(cron.WithSeconds(),
		cron.WithLogger(cronLog),
		cron.WithChain(cron.Recover(cronLog)),
	)

	if err := transactionDeliveryCron.Init(c, cronLog); err != nil {
		return err
	}

	c.Start()

	return nil
}
