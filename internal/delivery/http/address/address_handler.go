package address

import (
	"context"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type AddressHandlerCond struct {
	dig.In

	R *gin.Engine

	AddressUseCase usecase.AddressUseCase
}

func registerRouterAddress(cond AddressHandlerCond) {
	router := addressRouter{
		AddressHandlerCond: cond,
	}

	cond.R.POST("/address", router.create)
	cond.R.GET("/:address/balance/:cryptoType", router.getBalance)
	cond.R.GET("/address/:address", router.get)
	cond.R.GET("/address", router.getList)
}

type addressRouter struct {
	AddressHandlerCond
}

func (r *addressRouter) create(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var req vo.AddressCreateReq
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

	result, respStatus, err := r.AddressUseCase.Create(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("AddressUseCase.Create")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (r *addressRouter) getList(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	result, err := r.AddressUseCase.GetList(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("AddressUseCase.GetList")

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (r *addressRouter) get(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var address string
	if address = c.Param("address"); address == "" {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeAddressInvalidLength))
		return
	}

	result, err := r.AddressUseCase.GetByAddress(ctx, address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.NewError(response.Status{Messages: "address not found"}))
			return
		}

		log.WithFields(log.Fields{
			"err":     err,
			"address": fmt.Sprintf("%+v", address),
		}).Error("AddressUseCase.Get")

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (r *addressRouter) getBalance(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var req vo.AddressGetBalanceReq
	if respStatus, err := req.Parse(c); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("req.Parse")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	result, respStatus, err := r.AddressUseCase.GetBalance(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("AddressUseCase.GetBalance")

		c.JSON(http.StatusBadRequest, response.NewError(respStatus))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}
