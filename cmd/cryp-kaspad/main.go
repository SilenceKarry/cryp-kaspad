package main

import (
	"context"
	"cryp-kaspad/configs"
	deliveryCron "cryp-kaspad/internal/delivery/cron"
	deliveryHttp "cryp-kaspad/internal/delivery/http"
	"cryp-kaspad/internal/libs/container"
	"cryp-kaspad/internal/libs/logs"
	"cryp-kaspad/internal/libs/response"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
	"go.uber.org/dig"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	if err := configs.Start(); err != nil {
		panic(fmt.Sprintf("configs.Start, err: %s", err))
	}

	if err := logs.SetLogEnv(ctx); err != nil {
		panic(fmt.Sprintf("logs.SetLogEnv, err: %s", err))
	}
	configs.SetReloadFunc(logs.ReloadSetLogLevel)

	if configs.App.GetPyroscopeIsRunStart() {
		_, err := profiler.Start(profiler.Config{
			ApplicationName: configs.App.GetServiceName(),
			ServerAddress:   configs.App.GetPyroscopeURL(),
		})

		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("profiler.Start fail")
			return
		}
	}

	log.WithFields(log.Fields{
		"app": fmt.Sprintf("%+v", configs.App),
	}).Debug("check configs app value")

	container.Init()

	if err := container.ProvideInfra(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("container.ProvideInfra")
		return
	}

	if err := deliveryHttp.Init(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("deliveryHttp.Init")
		return
	}

	if err := deliveryCron.Init(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("deliveryCron.Init")
		return
	}

	// runGin
	if err := container.Get().Invoke(func(cond InfraCond) {
		if err := response.RegisterValidator(); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("response.RegisterValidator")
			return
		}

		runGinFlow(cond)
	}); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("runGin")
		return
	}
}

type InfraCond struct {
	dig.In

	R *gin.Engine
}

func runGinFlow(cond InfraCond) {
	srvServerRouter := runGinServerRouter(cond.R)

	log.Info("start service Success")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownTimeout := 10 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelCtx()

	if err := srvServerRouter.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("srvServerRouter.Shutdown")
		return
	}

	timeout := shutdownTimeout / time.Second
	log.WithFields(log.Fields{
		"shutdownTimeout": fmt.Sprintf("%ds", timeout),
	}).Info("server exiting")
}

func runGinServerRouter(r *gin.Engine) *http.Server {
	gin.SetMode(configs.App.GetGinMode())

	srv := &http.Server{
		Addr:    configs.App.GetPort(),
		Handler: r,
	}

	// 優雅關閉功能，無法攔截 kill -9 信號
	// 當服務做 kill 指令時，會攔截信號，並不接受新的 API 請求，
	// 在執行 shutdownTimeout 秒數，讓正在執行 API 功能，盡量執行結束，
	go func(srv *http.Server) {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.WithFields(log.Fields{
				"condition": "err != nil and err != http.ErrServerClosed",
				"err":       err,
			}).Error("srv.ListenAndServe")
			return
		}

		log.WithFields(log.Fields{
			"msg": err.Error(),
		}).Info("srv.ListenAndServe")
	}(srv)

	return srv
}
