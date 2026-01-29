package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/iWuxc/go-wit/log"
	"github.com/gin-gonic/gin/binding"
	zhongwen "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	trans ut.Translator
	_     binding.StructValidator = (*defaultValidator)(nil)
)

type rule map[string]validator.Func
type ruleMsg map[string]validator.RegisterTranslationsFunc

// CustomRule 自定义校验数据
type CustomRule struct {
	Name string
	Rule validator.Func
	Msg  string
}

type Validate interface {
	Validators() []CustomRule
}

type defaultValidator struct {
	vv       Validate
	once     sync.Once
	validate *validator.Validate
}

func NewValidator(v ...Validate) *defaultValidator {
	zh := zhongwen.New()
	uni := ut.New(zh, zh)
	trans, _ = uni.GetTranslator("zh")

	if len(v) == 0 {
		v = append(v, newDemoValidate())
	}
	return &defaultValidator{vv: v[0]}
}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyInit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// 注册自定义验证规则
		for _, r := range v.vv.Validators() {
			if len(r.Name) == 0 || r.Rule == nil {
				continue
			}

			if err := v.validate.RegisterValidation(r.Name, r.Rule); err != nil {
				log.Errorf("Failed to register validator: %s", err.Error())
			}

			if len(r.Msg) == 0 {
				continue
			}

			if err := v.validate.RegisterTranslation(r.Name, trans, func(ut ut.Translator) error {
				if err := trans.Add(r.Name, r.Msg, false); err != nil {
					return err
				}
				return nil
			}, translate); err != nil {
				log.Errorf("Failed to register customer rule msg validator: %s", err.Error())
			}
		}

		if err := zhTranslations.RegisterDefaultTranslations(v.validate, trans); err != nil {
			log.Errorf("Failed to register validator translation: %s", err.Error())
		}

		v.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			if f := field.Tag.Get("field"); f != "" {
				return f
			}

			return field.Name
		})

	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func Translate(errs validator.ValidationErrors) string {
	var errList []string
	for _, e := range errs {
		// can translate each error one at a time.
		errList = append(errList, e.Translate(trans))
	}
	return strings.Join(errList, "|")
}

// translate 自定义字段的翻译方法
func translate(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field())
	if err != nil {
		panic(fe.(error).Error())
	}
	return msg
}
