# Log

输出日志

## 基本使用

```go
package main

import (
	"github.com/iWuxc/go-wit/log"
)

func main() {
	log.Debug("hi.")
	log.Printf("%s", "hello, world")
}
```

## 替换默认 Log (输出到新的日志文件)

```go
package main

import (
	logger "github.com/iWuxc/go-wit/log"
)

var log = logger.NewLogger("name",
	logger.SetOutput("path.log", 1),
	logger.SetOutPutLevel("debug"),
)

func main() {
	log.Printf("%s", "hello, world")
}
```