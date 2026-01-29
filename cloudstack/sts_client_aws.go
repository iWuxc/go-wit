package cloudstack

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	sts2 "github.com/aws/aws-sdk-go/service/sts"
)

type AwsStsClient struct {
	StsConfig *StsConfig
	Client    *sts2.STS
}

func NewAwsStsClient(config *StsConfig) (Sts, error) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, "")

	c := aws.Config{
		Region:           aws.String(config.Region),
		Endpoint:         aws.String(config.Endpoint),
		S3ForcePathStyle: aws.Bool(config.S3ForcePathStyle),
		Credentials:      creds,
	}

	if config.Https == false {
		c.DisableSSL = aws.Bool(true)
	}

	return &AwsStsClient{
		StsConfig: config,
		Client:    sts2.New(session.New(&c)),
	}, nil

}

func (s *AwsStsClient) GetSessionToken() (*AssumeOutput, error) {

	duration := s.StsConfig.DurationSeconds

	if duration < 900 {
		duration = 3600
	}
	input := &sts2.AssumeRoleInput{
		DurationSeconds: aws.Int64(int64(duration)),
		RoleArn:         aws.String(s.StsConfig.RoleArn),
		RoleSessionName: aws.String(s.StsConfig.RoleSessionName),
		//Policy:          aws.String(s.StsConfig.Policy),
	}
	if len(s.StsConfig.Policy) > 0 {
		input.Policy = aws.String(s.StsConfig.Policy)
	}

	//发起请求，并得到响应。
	response, err := s.Client.AssumeRole(input)
	if err != nil {
		return &AssumeOutput{}, err
	}

	return &AssumeOutput{
		Credentials: Credentials{
			AccessKeySecret: *response.Credentials.SecretAccessKey,
			Expiration:      response.Credentials.Expiration.Format("2006-01-02T15:04:05Z"),
			AccessKeyId:     *response.Credentials.AccessKeyId,
			SecurityToken:   *response.Credentials.SessionToken,
		},
	}, nil

}
