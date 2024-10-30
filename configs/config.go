package configs

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm/logger"

	"github.com/spf13/viper"
)

const (
	envServiceName = "env.serviceName"

	envLogLevel                      = "env.log.level"
	envLogOutput                     = "env.log.output"
	envLogOutputFile                 = "env.log.outputs.file"
	envLogOutputElasticsearchAddress = "env.log.outputs.elasticsearch"

	envDebug = "env.debug"
	envPort  = "env.port"

	pyroscopeIsRunStart = "pyroscope.isRunStart"
	pyroscopeURL        = "pyroscope.url"

	dbGormLogMode = "db.gorm.logMode"

	dbMasterHost     = "db.master.host"
	dbMasterPort     = "db.master.port"
	dbMasterUsername = "db.master.username"
	dbMasterPassword = "db.master.password"
	dbMasterName     = "db.master.name"

	dbMasterMaxIdleConns    = "db.master.maxIdleConns"
	dbMasterMaxOpenConns    = "db.master.maxOpenConns"
	dbMasterConnMaxLifetime = "db.master.connMaxLifetime"

	nodeURL      = "node.url"
	nodeConfirm  = "node.confirm"
	nodeGasPrice = "node.gasPrice"
	AccountID    = "node.accountID"
	PrivateKey   = "node.privateKey"
	Retry        = "node.retry"

	notifyOPDevURL = "notify.opDevURL"
	notifyOPPreURL = "notify.opPreURL"
	notifyOPURL    = "notify.opURL"
	notifyQAURL    = "notify.qaURL"

	goroutineSizeBlock = "goroutineSize.Block"
	goroutineSizeTrans = "goroutineSize.Trans"

	walletDesposit = "wallet.deposit"
)

func newConfigApp() *configApp {
	viper.SetDefault(envLogLevel, "warn")
	viper.SetDefault(envLogOutput, "file")
	viper.SetDefault(envLogOutputFile, "/var/log/layout_demo_log")

	viper.SetDefault(envPort, "8080")

	viper.SetDefault(dbGormLogMode, "warn")

	viper.SetDefault(dbMasterMaxIdleConns, 10)
	viper.SetDefault(dbMasterMaxOpenConns, 20)
	viper.SetDefault(dbMasterConnMaxLifetime, 30)

	whitelistedKeys := map[string]interface{}{
		nodeURL:    nil,
		PrivateKey: nil,
		AccountID:  nil,
	}

	whitelistedKeyToViper := map[ViperKey]string{
		ViperNodeURL:    nodeURL,
		ViperPrivateKey: PrivateKey,
		ViperAccountID:  AccountID,
	}

	var wd WalletDesposit
	if err := viper.UnmarshalKey(walletDesposit, &wd); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %s", err))
	}

	return &configApp{
		serviceName: viper.GetString(envServiceName),

		whitelistedKeys:       whitelistedKeys,
		whitelistedKeyToViper: whitelistedKeyToViper,

		logLevel:                   viper.GetString(envLogLevel),
		logOutPut:                  viper.GetString(envLogOutput),
		logOutFile:                 viper.GetString(envLogOutputFile),
		logOutElasticsearchAddress: viper.GetStringSlice(envLogOutputElasticsearchAddress),

		debug: viper.GetBool(envDebug),
		port:  ":" + viper.GetString(envPort),

		pyroscopeIsRunStart: viper.GetBool(pyroscopeIsRunStart),
		pyroscopeURL:        viper.GetString(pyroscopeURL),

		gormLogMode: viper.GetString(dbGormLogMode),

		dbHost:     viper.GetString(dbMasterHost),
		dbPort:     viper.GetString(dbMasterPort),
		dbUsername: viper.GetString(dbMasterUsername),
		dbPassword: viper.GetString(dbMasterPassword),
		dbName:     viper.GetString(dbMasterName),

		maxIdleConns:    viper.GetInt(dbMasterMaxIdleConns),
		maxOpenConns:    viper.GetInt(dbMasterMaxOpenConns),
		connMaxLifetime: viper.GetDuration(dbMasterConnMaxLifetime),

		nodeURL:      viper.GetStringSlice(nodeURL),
		nodeConfirm:  viper.GetInt64(nodeConfirm),
		nodeGasPrice: viper.GetFloat64(nodeGasPrice),
		PrivateKey:   viper.GetString(PrivateKey),
		AccountID:    viper.GetString(AccountID),
		Retry:        viper.GetInt(Retry),

		notifyOPDevURL: viper.GetString(notifyOPDevURL),
		notifyOPPreURL: viper.GetString(notifyOPPreURL),
		notifyOPURL:    viper.GetString(notifyOPURL),
		notifyQAURL:    viper.GetString(notifyQAURL),

		goroutineSizeBlock: viper.GetInt(goroutineSizeBlock),
		goroutineSizeTrans: viper.GetInt(goroutineSizeTrans),

		WalletDesposit: wd,
	}
}

type ViperKey int

const (
	ViperNodeURL ViperKey = iota
	ViperPrivateKey
	ViperAccountID
)

type WalletDesposit struct {
	AccountID  string `mapstructure:"accountID"`
	PrivateKey string `mapstructure:"privateKey"`
	PublicKey  string `mapstructure:"publicKey"`
}

type configApp struct {
	sync.RWMutex

	serviceName string

	whitelistedKeys       map[string]interface{}
	whitelistedKeyToViper map[ViperKey]string

	logLevel                   string
	logOutPut                  string
	logOutFile                 string
	logOutElasticsearchAddress []string

	debug bool
	port  string

	pyroscopeIsRunStart bool
	pyroscopeURL        string

	gormLogMode string

	dbHost     string
	dbPort     string
	dbUsername string
	dbPassword string
	dbName     string

	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration

	nodeURL      []string
	nodeConfirm  int64
	nodeGasPrice float64
	PrivateKey   string
	AccountID    string
	Retry        int

	notifyOPDevURL string
	notifyOPPreURL string
	notifyOPURL    string
	notifyQAURL    string

	goroutineSizeBlock int
	goroutineSizeTrans int

	WalletDesposit WalletDesposit
}

func (c *configApp) reload() {
	c.Lock()
	defer c.Unlock()

	c.logLevel = viper.GetString(envLogLevel)
}

func (c *configApp) GetServiceName() string {
	return c.serviceName
}

func (c *configApp) GetLogLevel() string {
	c.RLock()
	defer c.RUnlock()

	return c.logLevel
}

func (c *configApp) GetLogOutPutType() string {
	return c.logOutPut
}

func (c *configApp) GetLogOutPutFile() string {
	return c.logOutFile
}

func (c *configApp) GetElasticsearchAddress() []string {
	return c.logOutElasticsearchAddress
}

func (c *configApp) GetDebug() bool {
	return c.debug
}

func (c *configApp) GetGinMode() string {
	if c.GetDebug() {
		return gin.DebugMode
	}

	return gin.ReleaseMode
}

func (c *configApp) GetPort() string {
	return c.port
}

func (c *configApp) GetPyroscopeIsRunStart() bool {
	return c.pyroscopeIsRunStart
}

func (c *configApp) GetPyroscopeURL() string {
	return c.pyroscopeURL
}

func (c *configApp) GetGormLogMode() logger.LogLevel {
	logMode := strings.ToLower(c.gormLogMode)

	switch logMode {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	}

	return logger.Warn
}

func (c *configApp) GetDBHost() string {
	return c.dbHost
}

func (c *configApp) GetDBPort() string {
	return c.dbPort
}

func (c *configApp) GetDBUsername() string {
	return c.dbUsername
}

func (c *configApp) GetDBPassword() string {
	return c.dbPassword
}

func (c *configApp) GetDBName() string {
	return c.dbName
}

func (c *configApp) GetMaxIdleConns() int {
	return c.maxIdleConns
}

func (c *configApp) GetMaxOpenConns() int {
	return c.maxOpenConns
}

func (c *configApp) GetConnMaxLifetime() time.Duration {
	return c.connMaxLifetime * time.Second
}

func (c *configApp) GetNodeURL() []string {
	return c.nodeURL
}

func (c *configApp) GetNodeConfirm() int64 {
	return c.nodeConfirm
}

func (c *configApp) GetNodeGasPrice() float64 {
	return c.nodeGasPrice
}

func (c *configApp) GetPrivateKey() string {
	return c.PrivateKey
}

func (c *configApp) GetAccountID() string {
	return c.AccountID
}

func (c *configApp) GetRetry() int {
	return c.Retry
}

func (c *configApp) GetNotifyOPDevURL() string {
	return c.notifyOPDevURL
}

func (c *configApp) GetNotifyOPPreURL() string {
	return c.notifyOPPreURL
}

func (c *configApp) GetNotifyOPURL() string {
	return c.notifyOPURL
}

func (c *configApp) GetNotifyQAURL() string {
	return c.notifyQAURL
}

func (c *configApp) GetGoroutineSizeBlock() int {
	return c.goroutineSizeBlock
}

func (c *configApp) GetGoroutineSizeTrans() int {
	return c.goroutineSizeTrans
}

func (c *configApp) GetWalletDesposit() WalletDesposit {
	return c.WalletDesposit
}

func (c *configApp) IsWhitelistedKey(key string) bool {
	_, b := c.whitelistedKeys[key]
	return b
}

func (c *configApp) TryGetKey(vKey ViperKey) (string, bool) {
	key, b := c.whitelistedKeyToViper[vKey]
	return key, b
}

func (c *configApp) GetValueWithKey(key string) (interface{}, error) {

	var value interface{}
	switch key {
	case nodeURL:
		value = c.nodeURL[0]
	default:
		return nil, fmt.Errorf("unwhitelisted key: %s", key)
	}
	return value, nil
}

func (c *configApp) Update(key string, value interface{}) error {
	c.Lock()
	defer c.Unlock()

	// 更新配置的值
	switch key {
	case nodeURL:
		c.nodeURL = value.([]string)
	default:
		return fmt.Errorf("unwhitelisted key: %s", key)
	}
	return nil
}
