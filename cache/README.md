# Cache

## 基本使用

### 内存驱动的缓存对象

```go
package main

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/cache"
	"time"
)

func main() {
	// 新初始化一个内存驱动的缓存对象
	c, err := cache.NewCache("memory", "")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	if err := c.Set(context.Background(), "foo", "bar", time.Minute); err != nil {
		// TODO . something
	}
	v, err := c.Get(context.Background(), "foo")
	if err != nil {
		// TODO . something
	}
	fmt.Println(v) // bar
	
	c.Delete(context.Background(), "foo")
}
```

### Redis 驱动的缓存对象

```go
package main

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/cache"
	"time"
)

func main() {
	// 新初始化一个内存驱动的缓存对象
	c, err := cache.NewCache("redis", "redis://localhost:6379/3?dial_timeout=3&db=1&read_timeout=6s&max_retries=2")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	if err := c.Set(context.Background(), "foo", "bar", time.Minute); err != nil {
		// TODO . something
	}
	v, err := c.Get(context.Background(), "foo")
	if err != nil {
		// TODO . something
	}
	fmt.Println(v) // bar
	
	c.Delete(context.Background(), "foo")
}
```
