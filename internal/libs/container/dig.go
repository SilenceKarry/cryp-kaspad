package container

import (
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/libs/mysql"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

var (
	c *dig.Container

	once sync.Once
)

func Init() {
	once.Do(func() {
		c = dig.New()
	})
}

func Get() *dig.Container {
	return c
}

func ProvideInfra() error {
	provideFunc := containerProvide{}

	if err := c.Provide(provideFunc.gin); err != nil {
		return fmt.Errorf("c.Provide(provideFunc.ginServerRouter), err: %w", err)
	}

	if err := c.Provide(provideFunc.mysqlMaster, dig.Name("dbM")); err != nil {
		return fmt.Errorf("c.Provide(provideFunc.mysqlMaster), err: %w", err)
	}

	return nil
}

type containerProvide struct {
}

// gin 建立 gin Engine，設定 middleware
func (cp *containerProvide) gin() *gin.Engine {
	return gin.Default()
}

// gorm 建立 gorm.DB 設定，初始化 session 並無實際連線
func (cp *containerProvide) mysqlMaster() (*gorm.DB, error) {
	return mysql.NewMysql(configs.App.GetDBUsername(), configs.App.GetDBPassword(),
		configs.App.GetDBHost(), configs.App.GetDBPort(), configs.App.GetDBName(),
		configs.App.GetGormLogMode(),
		configs.App.GetMaxIdleConns(), configs.App.GetMaxOpenConns(), configs.App.GetConnMaxLifetime())
}
