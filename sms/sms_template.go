package sms

type Template struct {
	Alias    string `json:"alias"`
	SignName string `json:"sign_name"`
	Code     string `json:"code"`
}

// AddTemplate 添加短信模板 .
func AddTemplate(template Template) {
	templates.Store(template.Alias, template)
	templates.Store(template.Code, template)
}

// GetTemplate 获取短信模板 .
func GetTemplate(alias string) (Template, bool) {
	val, ok := templates.Load(alias)
	if ok {
		return val.(Template), true
	}
	return Template{}, false
}
