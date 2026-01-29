package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type demoValidate struct{}

func newDemoValidate() *demoValidate {
	return &demoValidate{}
}

func (d *demoValidate) Validators() []CustomRule {
	return []CustomRule{
		{
			Name: "mobile",
			Rule: mobile,
			Msg:  "{0}必须是一个有效的手机号",
		},
		{
			Name: "userEnName",
			Rule: userEnName,
		},
		{
			Name: "userName",
			Rule: userName,
		},
		{
			Name: "userEmail",
			Rule: userEmail,
		},
		{
			Name: "userPass",
			Rule: userPass,
		},
		{
			Name: "checkTaxId",
			Rule: checkTaxId,
			Msg:  "{0}必须是一个有效的纳税人识别号",
		},
	}
}

// -----------------------  自定义表单验证  -----------------------------------------

// 手机号校验
func mobile(fl validator.FieldLevel) bool {
	re := `^1[3456789]\d{9}$`
	r := regexp.MustCompile(re)
	return r.MatchString(fl.Field().String())
}

// UserEnName 用户账户名校验,只能包含英文字母或数字字符,下划线
func userEnName(fl validator.FieldLevel) bool {
	reg := `^[0-9a-zA-Z_]{1,}$`
	return regexp.MustCompile(reg).MatchString(fl.Field().String())
}

// UserName 用户真实姓名校验，只能包含汉字、英文或空格
func userName(fl validator.FieldLevel) bool {
	reg := "^[\u4e00-\u9fa5a-zA-Z\\s]{1,}$"
	return regexp.MustCompile(reg).MatchString(fl.Field().String())
}

// UserEmail 用户邮箱校验,只能使用目前 binggan isheji microdreams 的公司邮箱
func userEmail(fl validator.FieldLevel) bool {
	reg := `^[A-Za-z]{1,}@(binggan\.com|isheji\.com|microdreams\.com)$`
	return regexp.MustCompile(reg).MatchString(fl.Field().String())
}

// UserPass 用户密码校验，只能包含英文字母或数字字符,下划线,6-12位
func userPass(fl validator.FieldLevel) bool {
	reg := `^[0-9a-zA-Z_]{6,12}$`
	return regexp.MustCompile(reg).MatchString(fl.Field().String())
}

// 纳税人识别号校验
func checkTaxId(fl validator.FieldLevel) bool {
	var regArr = []string{
		`^[\da-zA-Z]{10,15}$`,
		`^\d{6}[\da-zA-Z]{10,12}$`,
		`^[a-zA-Z]\d{6}[\da-zA-Z]{9,11}$`,
		`^[a-zA-Z]{2}\d{6}[\da-zA-Z]{8,10}$`,
		`^\d{14}[\d][\da-zA-Z]{4,5}$`,
		`^\d{17}[\d][\da-zA-Z]{1,2}$`,
		`^[a-zA-Z]\d{14}[\d][\da-zA-Z]{3,4}$`,
		`^[a-zA-Z]\d{17}[\d][\da-zA-Z]{0,1}$`,
		`^[\d]{6}[\da-zA-Z]{13,14}$`,
		`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`,
	}
	for _, re := range regArr {
		r := regexp.MustCompile(re)
		if r.MatchString(fl.Field().String()) {
			return true
		}
	}
	return false
}
