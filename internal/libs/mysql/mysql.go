package mysql

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm/schema"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMysql(username, password, host, port, dbname string, logMode logger.LogLevel,
	maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),

		NamingStrategy: schema.NamingStrategy{
			// 全局禁用表名複數
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 最大閒置連線數
	sqlDB.SetMaxIdleConns(maxIdleConns)
	// 最大連線數
	sqlDB.SetMaxOpenConns(maxOpenConns)
	// 每條連線的存活時間
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

// newLogger 使用 log.StandardLogger 紀錄
func newLogger() logger.Interface {
	return logger.New(log.StandardLogger(), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})
}
