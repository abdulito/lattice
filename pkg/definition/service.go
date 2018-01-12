package definition

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

type Service interface {
	Interface
	Volumes() []*block.Volume
	Components() []*block.Component
	Resources() block.Resources
}

type ServiceValidator interface {
	Validate(Service) error
}

func NewServiceFromJSON(data []byte) (Service, error) {
	var decoded serviceEncoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	if decoded.Type != TypeService {
		return nil, fmt.Errorf("service type must be %v, got %v", TypeService, decoded.Type)
	}

	s := &service{
		name:       decoded.Name,
		volumes:    decoded.Volumes,
		components: decoded.Components,
		resources:  decoded.Resources,
	}
	return s, nil
}

type serviceEncoder struct {
	Type        string             `json:"type"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Volumes     []*block.Volume    `json:"volumes"`
	Components  []*block.Component `json:"components"`
	Resources   block.Resources    `json:"resources"`
}

type service struct {
	name        string
	description string
	volumes     []*block.Volume
	components  []*block.Component
	resources   block.Resources
}

func (s *service) Type() string {
	return TypeService
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Description() string {
	return s.description
}

func (s *service) Volumes() []*block.Volume {
	return s.volumes
}

func (s *service) Components() []*block.Component {
	return s.components
}

func (s *service) Resources() block.Resources {
	return s.resources
}

func (s *service) MarshalJSON() ([]byte, error) {
	encoder := serviceEncoder{
		Type:        TypeService,
		Name:        s.name,
		Description: s.description,
		Volumes:     s.volumes,
		Components:  s.components,
		Resources:   s.resources,
	}

	return json.Marshal(&encoder)
}
