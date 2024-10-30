package logs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"
)

type Option func(*elasticsearch)

type elasticsearchConfig struct {
	Address  []string
	Username string
	Password string
	Index    string

	Formatter logrus.Formatter

	// goroutines size
	RoutineSize uint64

	// chan buffer
	RoutineChanSize uint64
}

type elasticsearch struct {
	client *elastic.Client
	ctx    context.Context
	index  string

	formatter logrus.Formatter

	routines    []chan *logrus.Entry
	routineSize uint64
	routinesNum uint64

	data map[string]interface{}
}

func AddElasticsearchHook(ctx context.Context, conf elasticsearchConfig, options ...Option) error {
	client, err := elastic.NewClient(elastic.Config{
		Addresses: conf.Address,
		Username:  conf.Username,
		Password:  conf.Password,
	})
	if err != nil {
		return err
	}

	if conf.Formatter == nil {
		conf.Formatter = DefaultJsonFormatter
	}

	els := &elasticsearch{
		ctx:         ctx,
		client:      client,
		formatter:   conf.Formatter,
		index:       conf.Index,
		routineSize: conf.RoutineSize,
		data:        make(map[string]interface{}),
	}

	for _, option := range options {
		option(els)
	}

	els.routines = make([]chan *logrus.Entry, conf.RoutineSize)
	for i := uint64(0); i < conf.RoutineSize; i++ {
		c := make(chan *logrus.Entry, conf.RoutineChanSize)
		els.routines[i] = c
		go els.worker(i, c)
	}

	logrus.AddHook(els)

	return nil
}

func (e *elasticsearch) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (e *elasticsearch) Fire(entry *logrus.Entry) error {
	for key, value := range e.data {
		entry.Data[key] = value
	}

	num := atomic.AddUint64(&e.routinesNum, 1) % e.routineSize
	e.routines[num] <- entry
	return nil
}

func (e *elasticsearch) worker(index uint64, entry chan *logrus.Entry) {
	defer func() {
		if err := recover(); err != nil {
			b, _ := e.formatter.Format(logrus.WithError(fmt.Errorf("recover elasticsearch worker: %w", err)))
			log.Println(string(b))
			return
		}

		close(entry)

		for _ = range entry {
		}
	}()

	for {
		select {
		case value := <-entry:
			body, err := e.formatter.Format(value)
			if err != nil {
				break
			}

			req := esapi.IndexRequest{
				Index:   e.index,
				Body:    bytes.NewReader(body),
				Timeout: time.Second * 30,
			}

			if _, err = req.Do(e.ctx, e.client); err != nil {
				b, _ := e.formatter.Format(logrus.WithError(fmt.Errorf("elasticsearch index request error: %w", err)))
				log.Println(string(b))
			}
		case <-e.ctx.Done():
			log.Println("close elasticsearch log worker(%d)", index)
			return
		}
	}
}

func WithVersionField(version string) Option {
	return func(o *elasticsearch) {
		o.data["version"] = version
	}
}
