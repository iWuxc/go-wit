package cloudstack

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
)

type AliYunStsClient struct {
	StsConfig *StsConfig
	Client    *sts.Client
}

func NewAliYunStsClient(config *StsConfig) (Sts, error) {
	client, err := sts.NewClientWithAccessKey(config.Region,
		config.AccessKeyID, config.AccessKeySecret)

	if err != nil {
		return nil, err
	}

	return &AliYunStsClient{
		StsConfig: config,
		Client:    client,
	}, err
}

func (s *AliYunStsClient) GetSessionToken() (*AssumeOutput, error) {
	//构建请求对象。
	request := sts.CreateAssumeRoleRequest()

	if s.StsConfig.Https {
		request.Scheme = "https"
	}
	duration := s.StsConfig.DurationSeconds
	//Token有效期最小值为900秒，最大值为要扮演角色的MaxSessionDuration时间。  https://help.aliyun.com/document_detail/371864.html
	if duration < 900 {
		duration = 3600
	}
	request.DurationSeconds = requests.NewInteger(duration)

	//设置参数。
	request.RoleArn = s.StsConfig.RoleArn
	request.RoleSessionName = s.StsConfig.RoleSessionName

	//发起请求，并得到响应。
	response, err := s.Client.AssumeRole(request)
	if err != nil {
		return &AssumeOutput{}, err
	}

	return &AssumeOutput{
		RequestId: response.RequestId,
		Credentials: Credentials{
			AccessKeySecret: response.Credentials.AccessKeySecret,
			Expiration:      response.Credentials.Expiration,
			AccessKeyId:     response.Credentials.AccessKeyId,
			SecurityToken:   response.Credentials.SecurityToken,
		},
	}, nil

}
