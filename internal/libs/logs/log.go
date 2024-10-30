package logs

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/common/version"

	"cryp-kaspad/configs"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var DefaultJsonFormatter = &log.JSONFormatter{
	DataKey: "field",
	FieldMap: log.FieldMap{
		log.FieldKeyTime: "timestamp",
	},
}

func SetLogEnv(ctx context.Context) error {
	level, err := log.ParseLevel(configs.App.GetLogLevel())
	if err != nil {
		return err
	}

	log.SetLevel(level)
	log.SetFormatter(DefaultJsonFormatter)
	log.SetReportCaller(true)

	switch configs.App.GetLogOutPutType() {
	case "file":
		if err := logAddHookRotateLogs(DefaultJsonFormatter); err != nil {
			return err
		}
	case "elasticsearch":
		conf := elasticsearchConfig{
			Index:           configs.App.GetServiceName(),
			Address:         configs.App.GetElasticsearchAddress(),
			Formatter:       DefaultJsonFormatter,
			RoutineSize:     uint64(3),
			RoutineChanSize: uint64(10),
		}

		if err := AddElasticsearchHook(ctx, conf, WithVersionField(version.Version)); err != nil {
			return err
		}
	}

	logAddHockHideData()

	return nil
}

func logAddHookRotateLogs(formatter log.Formatter) error {
	writer, err := rotatelogs.New(
		configs.App.GetLogOutPutFile()+".%Y-%m-%d-%H-%M",

		rotatelogs.WithLinkName(configs.App.GetLogOutPutFile()),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		return err
	}

	log.AddHook(lfshook.NewHook(lfshook.WriterMap{
		log.PanicLevel: writer,
		log.FatalLevel: writer,
		log.ErrorLevel: writer,
		log.WarnLevel:  writer,
		log.InfoLevel:  writer,
		log.DebugLevel: writer,
		log.TraceLevel: writer,
	}, formatter))

	return nil
}

func ReloadSetLogLevel() {
	level, err := log.ParseLevel(configs.App.GetLogLevel())
	if err != nil {
		log.WithFields(log.Fields{
			"newLogLevel": configs.App.GetLogLevel(),
			"oldLogLevel": log.GetLevel().String(),
			"err":         err,
		}).Error("ReloadSetLogLevel")
		return
	}

	log.SetLevel(level)
}

func logAddHockHideData() {
	h := &hideData{}
	log.AddHook(h)
}

type hideData struct {
}

func (h *hideData) Levels() []log.Level {
	return log.AllLevels
}

func (h *hideData) Fire(entry *log.Entry) error {
	h.hideSecretKey(entry)
	return nil
}

func (h *hideData) hideSecretKey(entry *log.Entry) {
	data, ok := entry.Data["SecretKey"].(string)
	if ok && len(data) > 0 {
		entry.Data["SecretKey"] = "hide data is secret"
	}

	data, ok = entry.Data["secret_key"].(string)
	if ok && len(data) > 0 {
		entry.Data["secret_key"] = "hide data is secret"
	}

	reqData, ok := entry.Data["req"].(string)
	if ok {
		if strings.Index(reqData, "SecretKey") == -1 &&
			strings.Index(reqData, "secret_key") == -1 {
			return
		}

		splitKey := ""
		if strings.Index(reqData, "SecretKey") > -1 {
			splitKey = "SecretKey"
		}

		if strings.Index(reqData, "secret_key") > -1 {
			splitKey = "secret_key"
		}

		var str strings.Builder
		temp := strings.Split(reqData, splitKey)
		for i, v := range temp {
			if i == 1 {
				index := strings.Index(v, " ")

				str.WriteString(splitKey)
				str.WriteString(": hide data is secret")
				str.WriteString(v[index:])

				continue
			}

			str.WriteString(v)
		}

		entry.Data["req"] = str.String()
	}
}
