package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
)

const (
	ComponentTypeContainer        = "container"
	ContainerBuildTypeCommand     = "command_build"
	ContainerBuildTypeDockerImage = "docker_image"
	ContainerBuildTypeDockerBuild = "docker_build"
)

var ContainerType = definition.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeContainer,
}

type Container struct {
	Build *ContainerBuild `json:"build,omitempty"`
	Exec  *ContainerExec  `json:"exec,omitempty"`

	Ports map[int32]ContainerPort `json:"ports,omitempty"`

	HealthCheck *ContainerHealthCheck `json:"health_check,omitempty"`

	Resources *ContainerResources `json:"resources,omitempty"`
}

type ContainerBuild struct {
	CommandBuild *ContainerBuildCommand
	DockerImage  *DockerImage
	DockerBuild  *DockerBuild
}

func (b *ContainerBuild) UnmarshalJSON(data []byte) error {
	var e *containerBuildEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	switch e.Type {
	case ContainerBuildTypeCommand:
		var c *ContainerBuildCommand
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		b.CommandBuild = c
		return nil

	case ContainerBuildTypeDockerImage:
		var i *DockerImage
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}

		b.DockerImage = i
		return nil

	case ContainerBuildTypeDockerBuild:
		var dockerBuild *DockerBuild
		if err := json.Unmarshal(data, &dockerBuild); err != nil {
			return err
		}

		b.DockerBuild = dockerBuild
		return nil

	default:
		return fmt.Errorf("unrecognized container build type: %v", e.Type)
	}
}

func (b *ContainerBuild) MarshalJSON() ([]byte, error) {
	var e interface{}
	switch {
	case b.CommandBuild != nil:
		e = &containerBuildCommandEncoder{
			Type: ContainerBuildTypeCommand,
			ContainerBuildCommand: b.CommandBuild,
		}

	case b.DockerImage != nil:
		e = &containerBuildDockerImageEncoder{
			Type:        ContainerBuildTypeDockerImage,
			DockerImage: b.DockerImage,
		}

	case b.DockerBuild != nil:
		e = &containerBuildDockerBuildEncoder{
			Type:        ContainerBuildTypeDockerBuild,
			DockerBuild: b.DockerBuild,
		}

	default:
		return nil, fmt.Errorf("container build must have a type")
	}

	return json.Marshal(&e)
}

type containerBuildEncoder struct {
	Type string `json:"type"`
}

type ContainerBuildCommand struct {
	Source      *ContainerBuildSource     `json:"source,omitempty"`
	BaseImage   DockerImage               `json:"base_image"`
	Command     []string                  `json:"command,omitempty"`
	Environment ContainerBuildEnvironment `json:"environment,omitempty"`
}

type ContainerBuildEnvironment map[string]string

type containerBuildCommandEncoder struct {
	Type string `json:"type"`
	*ContainerBuildCommand
}

type containerBuildDockerImageEncoder struct {
	Type string `json:"type"`
	*DockerImage
}

type containerBuildDockerBuildEncoder struct {
	Type string `json:"type"`
	*DockerBuild
}

type ContainerBuildSource struct {
	GitRepository *GitRepository `json:"git_repository"`
}

type ContainerExec struct {
	Command     []string             `json:"command"`
	Environment ContainerEnvironment `json:"environment,omitempty"`
}

type ContainerEnvironment map[string]ValueOrSecret

type ContainerPort struct {
	Protocol       string                       `json:"protocol"`
	ExternalAccess *ContainerPortExternalAccess `json:"external_access,omitempty"`
}

func (c ContainerPort) Public() bool {
	return c.ExternalAccess != nil && c.ExternalAccess.Public
}

type ContainerPortExternalAccess struct {
	Public bool `json:"public"`
}

type ContainerHealthCheck struct {
	HTTP *ContainerHealthCheckHTTP `json:"http,omitempty"`
}

type ContainerHealthCheckHTTP struct {
	Path string `json:"path"`
	Port int32  `json:"port"`
}

type ContainerResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}
