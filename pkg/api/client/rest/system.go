package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	v1rest "github.com/mlab-lattice/system/pkg/api/v1/rest"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemSubpath = "/systems"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemClient(c rest.Client, baseURL string) *SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemSubpath),
	}
}

func (c *SystemClient) Create(id v1.SystemID, definitionURL string) (*v1.System, error) {
	request := v1rest.CreateSystemRequest{
		ID:            id,
		DefinitionURL: definitionURL,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	body, statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		system := &v1.System{}
		err = rest.UnmarshalBodyJSON(body, &system)
		return system, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) List() ([]v1.System, error) {
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var systems []v1.System
		err = rest.UnmarshalBodyJSON(body, &systems)
		return systems, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Get(id v1.SystemID) (*v1.System, error) {
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		system := &v1.System{}
		err = rest.UnmarshalBodyJSON(body, system)
		return system, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Delete(id v1.SystemID) error {
	body, statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Versions(id v1.SystemID) ([]v1.SystemVersion, error) {
	body, statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v/versions", c.baseURL, id)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var versions []v1.SystemVersion
		err = rest.UnmarshalBodyJSON(body, versions)
		return versions, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Builds(id v1.SystemID) clientv1.BuildClient {
	return newBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Deploys(id v1.SystemID) clientv1.DeployClient {
	return newDeployClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Teardowns(id v1.SystemID) clientv1.TeardownClient {
	return newTeardownClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Services(id v1.SystemID) clientv1.ServiceClient {
	return newServiceClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Secrets(id v1.SystemID) clientv1.SecretClient {
	return newSystemSecretClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}
