package cron

import (
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/libs/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type CronHandlerCond struct {
	dig.In

	R *gin.Engine

	TransUseCase usecase.TransactionUseCase
}

func registerRouterCronHandler(cond CronHandlerCond) {
	router := cronRouter{
		CronHandlerCond: cond,
	}

	cond.R.GET("/cron/listen_block", router.listenBlock)
	cond.R.GET("/cron/tran_confirm", router.runTransactionConfirm)
	cond.R.GET("/cron/tran_notify", router.runTransactionNotify)
	cond.R.GET("/cron/risk_control_notify", router.runRiskControlNotify)
}

type cronRouter struct {
	CronHandlerCond
}

func (r *cronRouter) listenBlock(c *gin.Context) {
	r.TransUseCase.ListenBlock()
	c.JSON(http.StatusOK, response.NewSuccess(nil))
	return

}

func (r *cronRouter) runTransactionConfirm(c *gin.Context) {
	r.TransUseCase.TransactionConfirm()
	c.JSON(http.StatusOK, response.NewSuccess(nil))
	return

}

func (r *cronRouter) runTransactionNotify(c *gin.Context) {
	r.TransUseCase.TransactionNotify()
	c.JSON(http.StatusOK, response.NewSuccess(nil))
	return

}

func (r *cronRouter) runRiskControlNotify(c *gin.Context) {
	r.TransUseCase.RiskControlNotify()
	c.JSON(http.StatusOK, response.NewSuccess(nil))
	return

}
