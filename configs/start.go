package configs

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	App *configApp

	reloadFunc []func()
)

func SetReloadFunc(f func()) {
	reloadFunc = append(reloadFunc, f)
}

func Start() error {
	if err := parseFlag(); err != nil {
		return err
	}

	log.Println("configName: ", viper.GetString("configName"))
	log.Println("configPath: ", viper.GetString("configPath"))

	viper.SetConfigName(viper.GetString("configName"))
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(viper.GetString("configPath"))

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	App = newConfigApp()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		App.reload()

		for _, f := range reloadFunc {
			f()
		}

		log.WithFields(log.Fields{
			"app": fmt.Sprintf("%+v", App),
		}).Println("reload configs app value")
	})

	return nil
}

func parseFlag() error {
	pflag.String("configName", "config", "config file name")
	pflag.String("configPath", "", "config file path")

	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return err
	}

	return nil
}
