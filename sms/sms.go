package sms

import (
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/pkg/errors"
	"sync"
)

const smsLoggerTemplate = `{time} {channel}: "{method} {uri} HTTP/{version}" {code} {cost} {hostname}`

var (
	client    *dysmsapi.Client
	templates sync.Map

	ERRTemplateNotFount = errors.New("SMS template not fount")
	ERRMobileLength     = errors.New("Not support multi mobiles")
)

func Init(regionID, accessKeyID, accessKeySecret string, temps ...Template) error {
	var err error
	if regionID == "" || accessKeyID == "" || accessKeySecret == "" {
		return errors.New("短信配置不正确，access_key_id:" + accessKeyID + ", access_key_secret:" + accessKeySecret)
	}

	client, err = dysmsapi.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return errors.Wrap(err, "短信配置不正确")
	}
	// SMS log
	client.SetLogger("info", "SMS", log.GetInstance().Out(), smsLoggerTemplate)

	for _, template := range temps {
		AddTemplate(template)
	}

	return nil
}

// SendSMS 发送短信 .
// @Param string alias 短信模板别名
// @Param string phoneNumbers 接收短信的手机号码
// @Param map[string]string args 短信模板参数
// @Param limits RateLimitInterface 短信发送频率限制
func SendSMS(alias, phoneNumbers string, args map[string]string, limits ...RateLimitInterface) error {
	template, ok := GetTemplate(alias)
	if !ok {
		return ERRTemplateNotFount
	}
	params, err := utils.StructToJson(args)
	if err != nil {
		return errors.Wrap(err, "marshal params")
	}

	if len(phoneNumbers) > 11 {
		return ERRMobileLength
	}

	if len(limits) == 0 {
		limits = append(limits, defaultRateLimit())
	}

	// 校验短信发送频率
	if err := limits[0].Limit(phoneNumbers); err != nil {
		return err
	}

	if err := send(phoneNumbers, template.SignName, template.Code, params); err != nil {
		return err
	}

	return nil
}

// send 发送短信 .
func send(mobile, sign, code, params string) error {
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = mobile
	request.SignName = sign
	request.TemplateCode = code
	request.TemplateParam = params
	if _, e := client.SendSms(request); e != nil {
		return e
	}

	return nil
}
