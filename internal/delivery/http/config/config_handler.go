package config

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
)

type ConfigParams struct {
	dig.In
	R             *gin.Engine
	ConfigUseCase usecase.ConfigUseCase
}

type configRouter struct {
	ConfigParams
}

func registerConfigRouter(cond ConfigParams) {
	router := configRouter{
		cond,
	}

	subRouter := router.R.Group("/config")
	{
		subRouter.GET("", router.get)
		subRouter.GET("/demo", router.demo)
		subRouter.POST("/set", router.set)
		// subRouter.POST("/sync", router.sync)
	}
}

func (uc configRouter) get(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	config, err := uc.ConfigUseCase.GetAll(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("configRouter.get")

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	mapResult := make(map[string]interface{})
	for i := range config {
		mapResult[config[i].Key] = config[i].Value
	}

	c.JSON(http.StatusOK, response.NewSuccess(mapResult))
}

func (uc configRouter) demo(c *gin.Context) {
	_, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	mainNode := []configs.ViperKey{
		configs.ViperNodeURL,
	}
	vKeysSlice := [][]configs.ViperKey{mainNode}

	nodeKeys := make([][]string, 0)
	for _, vkeys := range vKeysSlice {
		temp := make([]string, 0)
		for _, vkey := range vkeys {
			nodeKey, ok := configs.App.TryGetKey(vkey)
			if !ok {
				continue
			}
			temp = append(temp, nodeKey)
		}
		nodeKeys = append(nodeKeys, temp)
	}

	c.JSON(http.StatusOK, response.NewSuccess(nodeKeys))
}

func (uc configRouter) set(c *gin.Context) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
	defer cancelCtx()

	// var req vo.SetConfig
	var req map[string]string
	if err := c.ShouldBind(&req); err != nil {
		log.WithFields(log.Fields{
			"req": fmt.Sprintf("%+v", req),
			"err": err,
		}).Error("configRouter.set ShouldBind")

		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
		return
	}

	for k := range req {
		if isWhitelisted := configs.App.IsWhitelistedKey(k); !isWhitelisted {
			log.WithFields(log.Fields{
				"key": k,
				"err": "is not whitelisted",
			}).Error("configRouter.set IsWhitelistedKey")
			c.JSON(http.StatusBadRequest, response.NewError(response.CodeParamInvalid))
			return
		}
	}

	for k, v := range req {
		if err := uc.ConfigUseCase.SetOrCreate(ctx, k, v); err != nil {
			log.WithFields(log.Fields{
				"req": fmt.Sprintf("%+v", req),
				"err": err,
			}).Error("configRouter.set SetOrCreate")

			c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
			return
		}
	}

	c.JSON(http.StatusOK, response.NewError(response.Status{}))
}

// func (uc configRouter) sync(c *gin.Context) {
// 	ctx, cancelCtx := context.WithTimeout(context.Background(), utils.Time30S)
// 	defer cancelCtx()

// 	if err := uc.ConfigUseCase.Sync(ctx); err != nil {
// 		log.WithFields(log.Fields{
// 			"err": err,
// 		}).Error("configRouter.sync Sync")

// 		c.JSON(http.StatusBadRequest, response.NewError(response.CodeInternalError))
// 		return
// 	}

// 	c.JSON(http.StatusOK, response.NewError(response.Status{}))
// }
