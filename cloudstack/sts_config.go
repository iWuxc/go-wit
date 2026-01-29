package cloudstack

type CloudStackConfig struct {
	Sts map[string]*StsConfig `json:"sts" mapstructure:"sts"`
}

type StsConfig struct {
	AccessKeyID      string `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret  string `json:"access_key_secret" mapstructure:"access_key_secret"`
	BucketName       string `json:"bucket_name" mapstructure:"bucket_name"`
	RoleArn          string `json:"role_arn" mapstructure:"role_arn"`
	Policy           string `json:"policy" mapstructure:"policy"`
	RoleSessionName  string `json:"role_session_name" mapstructure:"role_session_name"`
	Endpoint         string `json:"endpoint" mapstructure:"endpoint"`
	DurationSeconds  int    `json:"duration_seconds" mapstructure:"duration_seconds"`
	Https            bool   `json:"https" mapstructure:"https"`
	IamAccount       string `json:"iam_account" mapstructure:"iam_account"`
	IamPwd           string `json:"iam_pwd" mapstructure:"iam_pwd"`
	PrimaryAccount   string `json:"primary_account" mapstructure:"primary_account"`
	Region           string `json:"region" mapstructure:"region"`
	S3ForcePathStyle bool   `json:"s3_force_path_style" mapstructure:"s3_force_path_style"`
	Driver           string `json:"driver" mapstructure:"driver"`
	CloudEnv         string `json:"cloud_env" mapstructure:"cloud_env"`
}

func (c *CloudStackConfig) GetBucketName(app string) string {
	return c.Sts[app].BucketName
}

func (c *CloudStackConfig) GetAccessKeyId(app string) string {
	return c.Sts[app].AccessKeyID
}

func (c *CloudStackConfig) GetAccessKeySecret(app string) string {
	return c.Sts[app].AccessKeySecret
}

func (c *CloudStackConfig) GetRoleArn(app string) string {
	return c.Sts[app].RoleArn
}

func (c *CloudStackConfig) GetPolicy(app string) string {
	return c.Sts[app].Policy
}

func (c *CloudStackConfig) GetRoleSessionName(app string) string {
	return c.Sts[app].RoleSessionName
}

func (c *CloudStackConfig) GetDriver(app string) string {
	return c.Sts[app].Driver
}

func (c *CloudStackConfig) GetCloudEnv(app string) string {
	return c.Sts[app].CloudEnv
}

func (c *CloudStackConfig) GetRegion(app string) string {
	return c.Sts[app].Region
}

func (c *CloudStackConfig) GetEndpoint(app string) string {
	return c.Sts[app].Endpoint
}

func (c *CloudStackConfig) GetS3ForcePathStyle(app string) bool {
	return c.Sts[app].S3ForcePathStyle
}

func (c *CloudStackConfig) GetHttps(app string) bool {
	return c.Sts[app].Https
}
