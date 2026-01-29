package cloudstack

import (
	"fmt"
)

const awsSts = "aws"
const ossSts = "aliyun"
const ctYunObsSts = "ctyunobs"

type AssumeOutput struct {
	RequestId       string          `json:"RequestId" xml:"RequestId"`
	AssumedRoleUser AssumedRoleUser `json:"AssumedRoleUser" xml:"AssumedRoleUser"`
	Credentials     Credentials     `json:"Credentials" xml:"Credentials"`
}

type Credentials struct {
	AccessKeySecret string `json:"AccessKeySecret" xml:"AccessKeySecret"`
	Expiration      string `json:"Expiration" xml:"Expiration"`
	AccessKeyId     string `json:"AccessKeyId" xml:"AccessKeyId"`
	SecurityToken   string `json:"SecurityToken" xml:"SecurityToken"`
}

type AssumedRoleUser struct {
	AssumedRoleId string `json:"AssumedRoleId" xml:"AssumedRoleId"`
	Arn           string `json:"Arn" xml:"Arn"`
}

type Sts interface {
	GetSessionToken() (*AssumeOutput, error)
}

type Client struct {
	config *StsConfig
	Sts    Sts
}

func NewClient(config *StsConfig) (*Client, error) {
	var err error
	var sts Sts
	client := &Client{
		config: config,
	}
	switch config.Driver {
	case ossSts:
		sts, err = NewAliYunStsClient(config)
	case awsSts:
		sts, err = NewAwsStsClient(config)
	case ctYunObsSts:
		sts, err = NewCtYunObsStsClient(config)
	default:
		return nil, fmt.Errorf("sts driver not support yet %s", config.Driver)
	}

	if err != nil {
		return nil, err
	}

	client.Sts = sts

	return client, nil
}

func (c *Client) Name() ServiceName {
	return StsService
}

func (c *Client) Driver() string {
	return c.config.Driver
}

func (c *Client) Client() interface{} {
	return c.Sts
}
