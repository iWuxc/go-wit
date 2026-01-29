package sls

type options struct {
	accessKey    string
	accessSecret string
	endpoint     string
	project      string
	logstore     string

	securityToken string
	debug         bool
}

func defaultOptions() *options {
	return &options{
		project:  "projectName",
		logstore: "app",
	}
}

// WithEndpoint sets the endpoint of the SLS service.
func WithEndpoint(endpoint string) Option {
	return func(alc *options) {
		alc.endpoint = endpoint
	}
}

// WithProject sets the project name of the SLS service.
func WithProject(project string) Option {
	return func(alc *options) {
		alc.project = project
	}
}

// WithLogstore sets the logstore name of the SLS service.
func WithLogstore(logstore string) Option {
	return func(alc *options) {
		alc.logstore = logstore
	}
}

// WithAccessKeyID sets the access key of the SLS service.
func WithAccessKeyID(ak string) Option {
	return func(alc *options) {
		alc.accessKey = ak
	}
}

// WithAccessKeySecret sets the access secret of the SLS service.
func WithAccessKeySecret(as string) Option {
	return func(alc *options) {
		alc.accessSecret = as
	}
}

// WithClientSecurityToken sets the security token of the SLS service.
func WithClientSecurityToken(securityToken string) Option {
	return func(alc *options) {
		alc.securityToken = securityToken
	}
}

// WithDebug sets the debug mode of the SLS service.
func WithDebug(debug bool) Option {
	return func(alc *options) {
		alc.debug = debug
	}
}

type Option func(alc *options)
