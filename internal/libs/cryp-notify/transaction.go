package cryp_notify

import (
	"bytes"
	"context"
	"cryp-kaspad/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/moul/http2curl"

	"cryp-kaspad/internal/domain"

	"github.com/go-resty/resty/v2"
)

var (
	Transaction *transaction
)

func initTransaction() {
	Transaction = &transaction{}
}

type transaction struct {
}

func (trans *transaction) CreateTransactionNotify(ctx context.Context, host string, req CreateTransactionNotifyReq) (string, int, error) {
	type CommonResult struct {
		Status struct {
			Code int    `json:"code"`
			Msg  string `json:"messages"`
		} `json:"status"`

		Data interface{} `json:"data"`
	}

	url := fmt.Sprintf("%s/transaction", host)

	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, err
	}

	result := CommonResult{}

	client := resty.New()
	res, err := client.R().SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&result).
		Post(url)
	if err != nil {
		return "", 0, err
	}

	curl, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(string(body)))
	if err != nil {
		return "", 0, err
	}

	curl.Header.Set("Content-Type", "application/json")

	command, err := http2curl.GetCurlCommand(curl)
	if err != nil {
		return "", 0, err
	}

	if res.StatusCode() != http.StatusOK {
		return command.String(), 0, fmt.Errorf("call post transaction api fail, res.Status: %s", res.Status())
	}

	if result.Status.Code != 0 {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"curl":        command.String(),
			"resp":        fmt.Sprintf("%+v", result),
		}).Error("createTransactionNotify fail, result.Code != 0")

		return command.String(), domain.TxNotifyStatusFail, nil
	}

	log.WithFields(log.Fields{
		utils.LogUUID:    ctx.Value(utils.LogUUID),
		"curl":           command.String(),
		"resp":           fmt.Sprintf("%+v", result),
		"notify_tx_hash": req.TxHash,
	}).Warn("notify transaction info")

	return command.String(), domain.TxNotifyStatusSuccess, nil
}
