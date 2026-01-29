package cloudstack

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

type CtYunObsStsClient struct {
	StsConfig *StsConfig
}

type securitytoken struct {
	Credential credential `json:"credential"`
}

type credential struct {
	Access        string    `json:"access"`
	ExpiresAt     time.Time `json:"expires_at"`
	Secret        string    `json:"secret"`
	Securitytoken string    `json:"securitytoken"`
}

func NewCtYunObsStsClient(config *StsConfig) (Sts, error) {
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

	return &CtYunObsStsClient{
		StsConfig: config,
	}, nil
}

func (s *CtYunObsStsClient) securityToken(userToken string) (securitytoken, error) {
	jsonString := fmt.Sprintf(`{
    "auth": {
        "identity": {
            "methods": [
                "token"
            ],
            "token": {
                "id": "%s",
              "duration_seconds":%d
            }
        }
    }
}`, userToken, s.StsConfig.DurationSeconds)
	client := resty.NewWithClient(http.DefaultClient)
	response, err := client.R().SetHeader("Content-Type", "application/json").
		SetBody([]byte(jsonString)).SetQueryParams(nil).Post(s.StsConfig.Endpoint + "/v3.0/OS-CREDENTIAL/securitytokens")

	key := securitytoken{}
	if err != nil {
		return key, err
	}

	if response.StatusCode() != http.StatusCreated {
		return key, fmt.Errorf("get /v3.0/OS-CREDENTIAL/securitytokens err %s", response.Body())
	}

	err = json.Unmarshal(response.Body(), &key)
	if err != nil {
		return key, err
	}
	return key, nil
}

func (s *CtYunObsStsClient) userToken() (*string, error) {
	jsonString := fmt.Sprintf(`{
    "auth": {
        "identity": {
            "methods": [
                "password"
            ],
            "password": {
                "user": {
                    "domain": {
                        "name": "%s"
                    },
                    "name": "%s",
                    "password": "%s"
                }
            }
        },
        "scope": {
            "project": {
                "name": "%s"
            }
        }
    }
}`, s.StsConfig.PrimaryAccount, s.StsConfig.IamAccount, s.StsConfig.IamPwd, s.StsConfig.Region)
	client := resty.NewWithClient(http.DefaultClient)
	response, err := client.R().SetHeader("Content-Type", "application/json").
		SetBody([]byte(jsonString)).SetQueryParams(nil).Post(s.StsConfig.Endpoint + "/v3/auth/tokens")

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("get /v3/auth/tokens err %s", response.Body())
	}

	token := response.Header().Get("X-Subject-Token")

	return &token, nil
}
func (s *CtYunObsStsClient) GetSessionToken() (*AssumeOutput, error) {

	token, err := s.userToken()
	if err != nil {
		return &AssumeOutput{}, err
	}

	key, err := s.securityToken(*token)

	if err != nil {
		return &AssumeOutput{}, err
	}

	return &AssumeOutput{
		Credentials: Credentials{
			AccessKeySecret: key.Credential.Secret,
			Expiration:      key.Credential.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			AccessKeyId:     key.Credential.Access,
			SecurityToken:   key.Credential.Securitytoken,
		},
	}, nil
}