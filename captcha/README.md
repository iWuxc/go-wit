# Captcha

验证码

## 基本使用

```go
package main

import (
	"github.com/iWuxc/go-wit/captcha"
	"github.com/iWuxc/go-wit/log"
	"time"
)

const (
	captchaLen    = 4
	captchaWidth  = 200
	captchaHeight = 80

	captchaExpire = time.Minute * 10
)

func main() {
	// 生成新的验证码
	id, img := captcha.GetCaptcha(
		// 设置验证码宽
		captcha.WithWidth(captchaWidth),
		// 设置验证码高
		captcha.WithHeight(captchaHeight),
		// 设置验证码长度
		captcha.WithLength(captchaLen),
		// 获取验证码失败的重试次数
		captcha.WithMaxRetires(2),
		// 设置验证码有效期
		captcha.WithExpire(captchaExpire),
	)

	log.Printf("captcha id: %s, bas64: %s", id, img)
}

```
