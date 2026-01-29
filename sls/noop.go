package sls

import (
	"fmt"
	"github.com/iWuxc/go-wit/utils"
	"github.com/sirupsen/logrus"
	"os"
)

var _ LogService = (*noopLog)(nil)

type noopLog struct {
	opts *options
}

func NewNoopLog(options ...Option) (func(), error) {
	opts := defaultOptions()
	for _, o := range options {
		o(opts)
	}

	_defaultSLS = &noopLog{opts: opts}
	return func() {}, nil
}

func (n *noopLog) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (n *noopLog) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil
	}

	data := map[string]string{
		"level":    entry.Level.String(),
		"msg":      entry.Message,
		"time":     entry.Time.Format("2006-01-02 15:04:05.999"),
		"data":     utils.ToString(entry.Data),
		"source":   "project",
		"location": "project",
	}

	if source, ok := entry.Data["source"]; ok {
		data["source"] = utils.ToString(source)
	}

	if location, ok := entry.Data["location"]; ok {
		data["location"] = utils.ToString(location)
	}

	if entry.Caller != nil {
		caller, _ := utils.Marshal(map[string]interface{}{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		})
		data["caller"] = string(caller)
	}

	n.Log(data)
	return nil
}

func (n *noopLog) Log(content map[string]string) {
	var topic, source string
	if t, ok := content["topic"]; ok {
		topic = t
		delete(content, "topic")
	}

	if s, ok := content["source"]; ok {
		source = s
		delete(content, "source")
	}

	_ = n.Send(topic, source, content)

	return
}

func (n *noopLog) Send(topic, source string, content map[string]string) error {
	_, _ = fmt.Fprintf(os.Stdout, "sls:noop-log-output | topic: %+v, source: %+v, content: %+v \n", topic, source, content)
	return nil
}

func (n *noopLog) Search(params AliLogParams) (*AliLogContent, error) {
	return nil, nil
}
