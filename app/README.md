# APP

## 基本使用

```go
package main

import (
	"context"
	"github.com/iWuxc/go-wit/app"
	"github.com/iWuxc/go-wit/transport"
	"github.com/iWuxc/go-wit/transport/http"
)

func main() {
	ctx := context.Background()

	// http server
	httpServer := http.NewServer(
		http.Address(":8000"),
	)
	
	App, err := app.NewApp(app.Context(ctx), app.Server([]transport.ServerInterface{httpServer}), app.Version("v0.0.1"))
	if err != nil {
		panic(err)
    }
	App.Run()
}

```
