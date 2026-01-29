package cloudstack

import (
	"fmt"
)

type ServiceName string

const (
	StsService ServiceName = "sts" // sts
)

var cloudStackMgr = NewManager()

const DefaultApp = "default"

type StackService interface {
	Driver() string
	Name() ServiceName
	Client() interface{}
}

type Manager struct {
	services map[ServiceName]map[string]StackService
}

func NewManager() *Manager {
	mgr := &Manager{
		services: map[ServiceName]map[string]StackService{},
	}
	return mgr
}

func Mgr() *Manager {
	return cloudStackMgr
}
func (mg *Manager) Register(env string, servicename ServiceName, service StackService) {
	if mg.services[servicename] == nil {
		mg.services[servicename] = map[string]StackService{}
	}
	mg.services[servicename][env] = service
}

func (mg *Manager) Get(name ServiceName) StackService {
	return mg.services[name][DefaultApp]
}

func (mg *Manager) GetBy(servicename ServiceName, appname string) StackService {
	return mg.services[servicename][appname]
}

func Init(config *CloudStackConfig) error {

	if config == nil || config.Sts == nil {
		return fmt.Errorf("cloud stack config init err")
	}

	for k, value := range config.Sts {
		stsClient, err := NewClient(value)
		if err != nil {
			return err
		}
		Mgr().Register(k, stsClient.Name(), stsClient)
	}
	return nil
}
