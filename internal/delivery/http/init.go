package http

import (
	addressDeliveryHttp "cryp-kaspad/internal/delivery/http/address"
	configDeliveryHttp "cryp-kaspad/internal/delivery/http/config"
	cronDeliveryHttp "cryp-kaspad/internal/delivery/http/cron"
	feeDeliverHttp "cryp-kaspad/internal/delivery/http/fee"
	monitorDeliveryHttp "cryp-kaspad/internal/delivery/http/monitor"
	tokensDeliveryHttp "cryp-kaspad/internal/delivery/http/tokens"
	transactionDeliveryHttp "cryp-kaspad/internal/delivery/http/transaction"
	crypNotify "cryp-kaspad/internal/libs/cryp-notify"
)

func Init() error {
	crypNotify.Start()

	if err := configDeliveryHttp.Init(); err != nil {
		return err
	}

	if err := addressDeliveryHttp.Init(); err != nil {
		return err
	}

	if err := transactionDeliveryHttp.Init(); err != nil {
		return err
	}

	if err := monitorDeliveryHttp.Init(); err != nil {
		return err
	}

	if err := cronDeliveryHttp.Init(); err != nil {
		return err
	}

	if err := feeDeliverHttp.Init(); err != nil {
		return err
	}

	if err := tokensDeliveryHttp.Init(); err != nil {
		return err
	}

	return nil
}
