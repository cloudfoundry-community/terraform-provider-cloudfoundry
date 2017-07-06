package cfapi

import (
	"code.cloudfoundry.org/cli/cf/api/stacks"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// StackManager -
type StackManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	repo stacks.StackRepository
}

// CCStack -
type CCStack struct {
	ID          string
	Name        string
	Description string
}

// NewStackManager -
func newStackManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (sm *StackManager, err error) {

	sm = &StackManager{
		log: logger,

		config:    config,
		ccGateway: ccGateway,

		apiEndpoint: config.APIEndpoint(),

		repo: stacks.NewCloudControllerStackRepository(config, ccGateway),
	}

	return
}

// FindStackByName -
func (sm *StackManager) FindStackByName(name string) (stack CCStack, err error) {

	var s models.Stack

	s, err = sm.repo.FindByName(name)
	if err == nil {
		stack.ID = s.GUID
		stack.Name = s.Name
		stack.Description = s.Description
	}
	return
}
