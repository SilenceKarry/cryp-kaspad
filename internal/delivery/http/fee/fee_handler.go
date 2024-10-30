package fee

import (
	"context"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"net/http"
)

type FeeHandlerCond struct {
	dig.In

	R *gin.Engine

	Fee usecase.FeeUseCase
}

type router struct {
	FeeHandlerCond
}

func registerRouterFee(cond FeeHandlerCond) {
	router := router{
		FeeHandlerCond: cond,
	}

	cond.R.GET("/fee/:crypto_type", router.getFee)
}

func (t *router) getFee(c *gin.Context) {
	req := vo.FeeReq{}
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*utils.Time30S)
	defer cancelCtx()
	if respStatus, err := req.Parse(c); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("req.Parse")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	resp, status, err := t.Fee.GetFee(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("Fee.GetFee")

		c.JSON(http.StatusBadRequest, response.NewError(status))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(resp))

}
