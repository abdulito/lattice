package v1

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Interface interface {
	Status() (bool, error)

	Systems() SystemClient
}

type SystemClient interface {
	Create(id v1.SystemID, definitionURL string) (*v1.System, error)
	List() ([]v1.System, error)
	Get(v1.SystemID) (*v1.System, error)
	Delete(v1.SystemID) error

	Builds(v1.SystemID) BuildClient
	Deploys(v1.SystemID) DeployClient
	Teardowns(v1.SystemID) TeardownClient
	Services(v1.SystemID) ServiceClient
	Secrets(v1.SystemID) SecretClient
}

type BuildClient interface {
	Create(version string) (*v1.Build, error)
	List() ([]v1.Build, error)
	Get(v1.BuildID) (*v1.Build, error)
}

type DeployClient interface {
	CreateFromBuild(v1.BuildID) (*v1.Deploy, error)
	CreateFromVersion(string) (*v1.Deploy, error)
	List() ([]v1.Deploy, error)
	Get(v1.DeployID) (*v1.Deploy, error)
}

type TeardownClient interface {
	Create() (*v1.Teardown, error)
	List() ([]v1.Teardown, error)
	Get(v1.TeardownID) (*v1.Teardown, error)
}

type ServiceClient interface {
	List() ([]v1.Service, error)
	Get(v1.ServiceID) (*v1.Service, error)
}

type SecretClient interface {
	List() ([]v1.Secret, error)
	Get(path tree.NodePath, name string) (*v1.Secret, error)
	Set(path tree.NodePath, name, value string) error
	Unset(path tree.NodePath, name string) error
}
