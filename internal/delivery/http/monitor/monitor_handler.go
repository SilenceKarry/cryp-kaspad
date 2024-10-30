package monitor

import (
	"cryp-kaspad/internal/libs/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type MonitorHandlerCond struct {
	dig.In

	R *gin.Engine
}

func registerRouterMonitor(cond MonitorHandlerCond) {
	router := monitorRouter{
		MonitorHandlerCond: cond,
	}

	cond.R.GET("/health", router.getHealth)
}

type monitorRouter struct {
	MonitorHandlerCond
}

func (r *monitorRouter) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, response.NewSuccess(nil))
}
