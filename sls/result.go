package sls

import (
	"github.com/iWuxc/go-wit/utils"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
)

type result struct {}

func (r *result) Success(res *producer.Result) {
	utils.DebugInfo("Send Log To AliSLS Success", res)
}

func (r *result) Fail(res *producer.Result) {
	utils.DebugInfo("Send Log To AliSLS Fail", res)
}
