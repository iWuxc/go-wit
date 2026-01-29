package captcha

import (
	"github.com/mojocn/base64Captcha"
	"time"
)

const (
	defaultCaptchaLen = 5
	defaultWidth      = 200
	defaultHeight     = 80
	defaultMaxSkew    = 0.6
	defaultDotCount   = 8

	defaultMaxRetires  = 5
	defaultLifeExpired = time.Minute * 10
)

// GetCaptcha 生成验证码 .
func GetCaptcha(fs ...Fn) (string, string) {
	opt := options{
		height:   defaultHeight,
		width:    defaultWidth,
		len:      defaultCaptchaLen,
		maxSkew:  defaultMaxSkew,
		dotCount: defaultDotCount,

		maxReties: defaultMaxRetires,
		expire:    defaultLifeExpired,
	}

	for _, f := range fs {
		f(&opt)
	}

	for i := 0; i < opt.maxReties; i++ {
		id, b64, err := genCaptcha(opt)
		if err == nil {
			return id, b64
		}
	}
	return "", ""
}

// Verify 校验 .
func Verify(captchaId, captcha string) bool {
	client := GetCaptchaStore(defaultLifeExpired)
	return client.Verify(captchaId, captcha, true)
}

// genCaptcha 生成验证码 .
func genCaptcha(opt options) (string, string, error) {
	driver := &base64Captcha.DriverDigit{
		Height:   opt.height,
		Width:    opt.width,
		Length:   opt.len,
		MaxSkew:  opt.maxSkew,
		DotCount: opt.dotCount,
	}

	client := GetCaptchaStore(opt.expire)
	id, b64, err := base64Captcha.NewCaptcha(driver, client).Generate()
	return id, b64, err
}
