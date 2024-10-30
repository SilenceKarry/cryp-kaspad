package transaction

import (
	"context"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type TransactionHandlerCond struct {
	dig.In

	R *gin.Engine

	TransUseCase usecase.TransactionUseCase
}

func registerRouterTransaction(cond TransactionHandlerCond) {
	router := transactionRouter{
		TransactionHandlerCond: cond,
	}

	cond.R.GET("/tx/:txHash", router.getTxHash)
	cond.R.GET("/blockHeight", router.getBlockHeight)
	cond.R.POST("/block/:blockNumber/transaction", router.createTransactionByBlockNumber)
}

type transactionRouter struct {
	TransactionHandlerCond
}

func (r *transactionRouter) getTxHash(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var req vo.TransGetTxHashReq
	if respStatus, err := req.Parse(c); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("req.Parse")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	result, respStatus, err := r.TransUseCase.GetByTxHash(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("TransUseCase.GetByTxHash")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (r *transactionRouter) getBlockHeight(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	result, respStatus, err := r.TransUseCase.GetBlockHeight(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("TransUseCase.GetBlockHeight")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (r *transactionRouter) createTransactionByBlockNumber(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	//Get Parameters.
	var req vo.CreateTransactionByBlockNumberReq
	if respStatus, err := req.Parse(c); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("req.Parse")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	list, respStatus, err := r.TransUseCase.CreateTransactionByBlockNumber(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("TransUseCase.CreateTransactionByBlockNumber")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(list))
}
