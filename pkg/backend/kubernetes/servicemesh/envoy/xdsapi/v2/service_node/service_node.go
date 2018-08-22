package servicenode

import (
	"reflect"
	"sync"

	"github.com/golang/glog"

	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
)

type ServiceNode struct {
	ID                 string
	latticeServiceName string

	EnvoyNode *envoycore.Node

	lock    sync.Mutex
	deleted bool

	clusters  []envoycache.Resource
	endpoints []envoycache.Resource
	routes    []envoycache.Resource
	listeners []envoycache.Resource
}

func NewServiceNode(id string, envoyNode *envoycore.Node) *ServiceNode {
	return &ServiceNode{
		ID:        id,
		EnvoyNode: envoyNode,
	}
}

func (s *ServiceNode) Domain() string {
	return s.EnvoyNode.GetId()
}

func (s *ServiceNode) Path() (tree.Path, error) {
	tnPath, err := tree.NewPathFromDomain(s.EnvoyNode.GetId())
	if err != nil {
		return "", err
	}
	return tnPath, nil
}

func (s *ServiceNode) ServiceCluster() string {
	return s.EnvoyNode.GetCluster()
}

func (s *ServiceNode) GetLatticeServiceName() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.latticeServiceName
}

func (s *ServiceNode) SetLatticeServiceName(latticeServiceName string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.latticeServiceName = latticeServiceName
}

func (s *ServiceNode) Update(backend xdsapi.Backend) error {
	glog.Info("ServiceNode.Update called")
	// disallow concurrent updates to service state
	s.lock.Lock()
	defer s.lock.Unlock()

	glog.V(4).Infof("Retrieving system services in service node %v", s.ID)
	systemServices, err := backend.SystemServices(s.ServiceCluster())
	if err != nil {
		return err
	}

	glog.V(4).Infof("Retrieving clusters in service node %v", s.ID)
	clusters, err := s.getClusters(systemServices)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Retrieving endpoints in service node %v", s.ID)
	endpoints, err := s.getEndpoints(clusters, systemServices)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Retrieving listeners in service node %v", s.ID)
	listeners, err := s.getListeners(systemServices)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Retrieving routes in service node %v", s.ID)
	routes, err := s.getRoutes(systemServices)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Checking if service node %v envoy config has changed", s.ID)
	if !reflect.DeepEqual(clusters, s.clusters) ||
		!reflect.DeepEqual(endpoints, s.endpoints) ||
		!reflect.DeepEqual(listeners, s.listeners) ||
		!reflect.DeepEqual(routes, s.routes) {
		glog.V(4).Infof("Service node envoy %v config has changed", s.ID)
		s.clusters = clusters
		s.endpoints = endpoints
		s.listeners = listeners
		s.routes = routes
		if !s.deleted {
			glog.V(4).Info("ServiceNode.Update updating XDS cache")
			err := backend.SetXDSCacheSnapshot(s.ID, s.endpoints, s.clusters, s.routes, s.listeners)
			if err != nil {
				return err
			}
		} else {
			glog.Warning("ServiceNode.Update called on deleted node")
		}
	}

	return nil
}

func (s *ServiceNode) Cleanup(backend xdsapi.Backend) error {
	glog.V(4).Info("ServiceNode.Cleanup called")

	s.lock.Lock()
	defer s.lock.Unlock()

	backend.ClearXDSCacheSnapshot(s.ID)
	s.deleted = true

	return nil
}
