# Validator

自定义表单校验, 可替换 gin 框架自带的 validator(版本过低, 不支持中文校验提示)

go-kit >= v1.1.9 支持自定义表单校验

## 基本使用

```go
package main

import (
	"github.com/iWuxc/go-wit/validator"
	"github.com/gin-gonic/gin/binding"
)

func main() {
	binding.Validator = validator.NewValidator(newDemoValidate())
}

type demoValidate struct{}

func newDemoValidate() *demoValidate {
	return &demoValidate{}
}

func (d *demoValidate) Validators() []validator.CustomRule {return []validator.CustomRule{}}
```
