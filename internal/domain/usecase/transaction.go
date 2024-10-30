package usecase

import (
	"context"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
)

type TransactionUseCase interface {
	GetByTxHash(ctx context.Context, req vo.TransGetTxHashReq) (vo.TransGetTxHashResp, response.Status, error)
	GetBlockHeight(ctx context.Context) (vo.BlockHeightGetResp, response.Status, error)

	ListenBlock()
	TransactionConfirm()
	TransactionNotify()
	RiskControlNotify()

	CreateTransactionByBlockNumber(ctx context.Context, req vo.CreateTransactionByBlockNumberReq) ([]entity.Transaction, response.Status, error)
}
