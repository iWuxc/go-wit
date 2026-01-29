
## Errors
> 错误统一维护

### 基本使用

通过 Prometheus 注册项目中的错误, 并在 Grafana 大盘中展示

```go
package main

import (
	 "github.com/iWuxc/go-wit/errors"
	 "github.com/iWuxc/go-wit/log"
)

func main()  {
    var err error
	
	err = errors.NewStatError("main", "error test", "error")
	
	log.Printf("err: %s", err)
}
```