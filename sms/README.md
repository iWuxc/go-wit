# SMS 
发送短信

- 支持多模板短信发送
- 支持 按别名(Alias)/按模板CODE(SMS_15305****) 发送模板短信
- 支持多参数
- 支持频率限制, 默认 同一号码每分钟只能发送一次, 同一号码每天最多发送十次

## 基本使用

```go
package main

import (
	"github.com/iWuxc/go-wit/sms"
)

func main() {
	if err := sms.Init("xxxxx", "xxxxxx", "xxxxxx"); err != nil {
		panic(err)
    }
	
	sms.AddTemplate(sms.Template{
		Alias:    "login",
		SignName: "micros",
		Code:     "SMS_1230511111",
	})
	
	if err := sms.SendSMS("login", "18888888888", map[string]string{"code": "1234"}); err != nil {
		// handle error.
    }
}
```