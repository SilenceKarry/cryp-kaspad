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

type WithdrawHandlerCond struct {
	dig.In

	R *gin.Engine

	WithdrawUseCase usecase.WithdrawUseCase
}

func registerRouterWithdraw(cond WithdrawHandlerCond) {
	router := withdrawRouter{
		WithdrawHandlerCond: cond,
	}

	cond.R.POST("/withdraw", router.create)
}

type withdrawRouter struct {
	WithdrawHandlerCond
}

func (r *withdrawRouter) create(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 30*utils.Time30S)
	defer cancelCtx()

	var req vo.WithdrawCreateReq
	if err := response.ShouldBindJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	if respStatus, err := req.Validate(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("req.Validate")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	result, respStatus, err := r.WithdrawUseCase.Create(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("WithdrawUseCase.Create")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}
