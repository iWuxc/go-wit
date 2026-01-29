# 队列

基于 Redis 驱动的快速可重试队列

## 功能介绍

- [x] 任务重试
- [x] 优先级队列
- [x] 允许每个任务的超时和截止日期
- [x] 支持中间件
- [ ] 指标监控
- [ ] 内容监控(队列数量, 运行状态)
- [ ] 多队列驱动
- [ ] 单元测试待完善
- [ ] example 示例待完善

## 使用说明

### Server

```go
package server

import (
	"context"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/queue"
	"github.com/iWuxc/go-wit/queue/server"
)

const (
	TaskName = "task:test"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

func main() {
	mux := queue.NewServeMux()
	mux.Handle(TaskName, NewTestQueue())

	srv := server.NewServer(server.Config{
		Concurrency: 10,
		Queues: map[string]int{
			QueueCritical: 6,
			QueueDefault:  3,
			QueueLow:      1,
		},
	})

	if err := srv.Run(mux); err != nil {
		panic(err)
	}
}

func NewTestTask() *queue.Task {
	return queue.NewTask(TaskName, []byte("go-kit 测试"))
}

type TestQueue struct {
	name string
}

func NewTestQueue() *TestQueue {
	return &TestQueue{name: "go-kit"}
}

func (t *TestQueue) ProcessTask(ctx context.Context, task *queue.Task) error {
	// TODO ...
	log.Info(string(task.Payload()))
	return nil
}

```

### Client

```go
package client

import (
	"github.com/iWuxc/go-wit/queue"
	"github.com/iWuxc/go-wit/queue/client"
	"xxxx.xxx/queues/server"
)

func main() {
	_, err := client.Enqueue(server.NewTestTask(), queue.OptQueue(server.QueueDefault), queue.OptMaxRetry(1))
	if err != nil {
		panic(err)
	}
}
```



