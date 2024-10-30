package cryp_notify

import (
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain"
)

var (
	MerchantType2URL = make(map[int]string)
)

func Start() {
	MerchantType2URL = map[int]string{
		domain.MerchantTypeOPDev: configs.App.GetNotifyOPDevURL(),
		domain.MerchantTypeOPPre: configs.App.GetNotifyOPPreURL(),
		domain.MerchantTypeOP:    configs.App.GetNotifyOPURL(),
		domain.MerchantTypeQA:    configs.App.GetNotifyQAURL(),
	}

	initTransaction()
}
