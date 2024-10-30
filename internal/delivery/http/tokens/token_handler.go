package tokens

import (
	"context"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type TokenRouterParams struct {
	dig.In
	R             *gin.Engine
	TokensUseCase usecase.TokensUseCase
}
type tokenRouter struct {
	TokenRouterParams
}

func registerRouterMonitor(cond TokenRouterParams) {
	router := tokenRouter{
		TokenRouterParams: cond,
	}

	groupRouter := router.R.Group("/token")
	{
		groupRouter.GET("/:contractAddr", router.get)
		groupRouter.POST("/create", router.create)
		groupRouter.POST("/:contractAddr/update", router.update)
		groupRouter.GET("/", router.getList)
		groupRouter.GET("", router.getList)
	}
}

func (uc tokenRouter) get(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	contractAddr := c.Param("contractAddr")
	if contractAddr == "" {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid))
		return
	}

	result, status, err := uc.TokensUseCase.GetByContractAddr(ctx, contractAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("token router TokensUseCase.GetByContractAddr")

		c.JSON(http.StatusBadRequest, response.NewError(status))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (uc tokenRouter) getList(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	result, err := uc.TokensUseCase.GetList(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("token router TokensUseCase.GetByContractAddr")

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(result))
}

func (uc tokenRouter) create(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var req vo.ContractCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("token router ShouldBindJSON")
		errField := getErrField(err)
		if errField != "" {
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(","+errField+" invalid")))
			return
		}

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	if req.GasPrice.IsZero() {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(",gasPrice invalid")))
		return
	}

	if err := uc.TokensUseCase.CreateContractToken(ctx, req); err != nil {
		if strings.Contains("contract abi params error", err.Error()) {
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid))
			return
		}

		log.WithFields(log.Fields{
			"err": err,
			"req": fmt.Sprintf("%+v", req),
		}).Error("token router TokensUseCase.CreateContractToken")
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(nil))
}

func (uc tokenRouter) update(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	var req vo.ContractUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("token router update ShouldBindJSON")
		errField := getErrField(err)
		if errField != "" {
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(","+errField+" invalid")))
			return
		}

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	contractAddr := c.Param("contractAddr")
	if contractAddr == "" {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg("contractAddr not empty")))
		return
	}

	if req.GasPrice != nil && req.GasPrice.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(",gasPrice invalid")))
		return
	}

	if req.Decimals != nil && *req.Decimals <= 0 {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(",decimals invalid")))
		return
	}

	if req.GasLimit != nil && *req.GasLimit <= 0 {
		c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid.WithMsg(",gasLimit invalid")))
		return
	}

	if err := uc.TokensUseCase.UpdateContractToken(ctx, req, contractAddr); err != nil {
		if strings.Contains("contract abi params error", err.Error()) {
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid))
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeCryptoNotFound))
			return
		}

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccess(nil))
}

func getErrField(err error) string {
	var errStr string
	switch err.(type) {
	case validator.ValidationErrors:
		paramsMap := map[string]string{
			"ContractAddr": "contractAddr",
			"Decimals":     "decimals",
			"GasLimit":     "gasLimit",
			"GasPrice":     "gasPrice",
			"ContractAbi":  "contractAbi",
			"CryptoType":   "cryptoType",
		}
		for _, fe := range err.(validator.ValidationErrors) {
			errStr = paramsMap[fe.Field()]
			break
		}
	case *json.UnmarshalTypeError:
		jsonErr := err.(*json.UnmarshalTypeError)
		errStr = jsonErr.Field

	default:
		if strings.Contains(err.Error(), "to decimal") {
			errStr = "gasPrice"
		}
	}

	return errStr
}
