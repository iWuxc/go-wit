package sls

import (
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/utils"
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/sirupsen/logrus"
	"time"
)

var _defaultSLS LogService

type LogService interface {
	logrus.Hook
	// Log .
	Log(content map[string]string)
	// Send set log to aliyun log service and return error if any.
	Send(topic, source string, content map[string]string) error
	// Search get log from aliyun log service and return error if any.
	Search(params AliLogParams) (*AliLogContent, error)
}

// Log see more detail https://github.com/aliyun/aliyun-log-go-sdk
type aliyunLog struct {
	producer *producer.Producer
	client   sls.ClientInterface
	opts     *options
}

func (a *aliyunLog) GetProducer() *producer.Producer {
	if a.producer == nil {
		panic("aliyunLog producer is nil, please check configure")
	}

	return a.producer
}

func Log(content map[string]string) {
	_defaultSLS.Log(content)
}

// Log send log to aliyun log service
func (a *aliyunLog) Log(content map[string]string) {
	var topic, source string
	if t, ok := content["topic"]; ok {
		topic = t
		delete(content, "topic")
	}

	if s, ok := content["source"]; ok {
		source = s
		delete(content, "source")
	}

	_ = a.Send(topic, source, content)

	return
}

// Levels .
func (a *aliyunLog) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire .
func (a *aliyunLog) Fire(entry *logrus.Entry) error {
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

	a.Log(data)
	return nil
}

func Send(topic, source string, content map[string]string) error {
	return _defaultSLS.Send(topic, source, content)
}

// Send set log to aliyun log service and return error if any
func (a *aliyunLog) Send(topic, source string, content map[string]string) error {
	var err error
	if a.opts.debug {
		err = a.GetProducer().SendLogWithCallBack(a.opts.project, a.opts.logstore, topic, source, producer.GenerateLog(uint32(time.Now().Unix()), content), new(result))
	} else {
		err = a.GetProducer().SendLog(a.opts.project, a.opts.logstore, topic, source, producer.GenerateLog(uint32(time.Now().Unix()), content))
	}

	if err != nil {
		return errors.Wrap(err, "aliyun log send error")
	}
	return nil
}

// AliLogParams is the params for aliyun log service
type AliLogParams struct {
	Start, End int64  // start, end  查询开始时间,查询结束时间 Unix时间戳格式
	Query      string // query 查询语句或者分析语句  SELECT remote_addr,COUNT(*) as pv GROUP by remote_addr ORDER by pv desc limit 5
	MaxLineNum int64  // 仅当query参数为查询语句时，该参数有效，表示请求返回的最大日志条数。最小值为0，最大值为100，默认值为100。
	Offset     int64  // 仅当query参数为查询语句时，该参数有效，表示查询开始行。默认值为0。
	Reverse    bool   // 用于指定返回结果是否按日志时间戳降序返回日志，精确到分钟级别。 true：按照日志时间戳降序返回日志。 false（默认值）：按照日志时间戳升序返回日志。
}

type AliLogContent struct {
	Total   int64               `json:"total"`
	Logs    []map[string]string `json:"logs"`
	Content string              `json:"content"`
}

func Search(params AliLogParams) (*AliLogContent, error) {
	return _defaultSLS.Search(params)
}

// Search return log from aliyun log service
func (a *aliyunLog) Search(params AliLogParams) (*AliLogContent, error) {
	if params.Start > params.End || params.Start < 1 {
		return nil, errors.ErrInvalidTimeRange
	}

	resp, err := a.client.GetLogs(a.opts.project, a.opts.logstore, "", params.Start, params.End, params.Query, params.MaxLineNum, params.Offset, params.Reverse)
	if err != nil {
		return nil, err
	}

	return &AliLogContent{
		Total:   resp.Count,
		Logs:    resp.Logs,
		Content: resp.Contents,
	}, nil
}

// NewAliyunLog new an aliyun logger with options.
func NewAliyunLog(options ...Option) (func(), error) {
	opts := defaultOptions()
	for _, o := range options {
		o(opts)
	}

	if opts.endpoint == "" || opts.accessKey == "" || opts.accessSecret == "" {
		return nil, nil
	}

	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.AllowLogLevel = "error"
	if opts.debug {
		producerConfig.AllowLogLevel = "debug"
	}
	producerConfig.Endpoint = opts.endpoint
	producerConfig.AccessKeyID = opts.accessKey
	producerConfig.AccessKeySecret = opts.accessSecret
	producerInst := producer.InitProducer(producerConfig)

	client := sls.CreateNormalInterface(opts.endpoint, opts.accessKey, opts.accessSecret, opts.securityToken)

	_defaultSLS = &aliyunLog{
		producer: producerInst,
		client:   client,
		opts:     opts,
	}

	producerInst.Start()
	return func() {
		_ = client.Close()
		producerInst.SafeClose()
	}, nil
}

// GetAliLogService return default aliyun log service
func GetAliLogService() LogService {
	return _defaultSLS
}
