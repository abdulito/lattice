package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type ServiceNode struct {
	parent  tree.Node
	path    tree.Path
	service *Service
}

func NewServiceNode(service *Service, name string, parent tree.Node) *ServiceNode {
	return &ServiceNode{
		parent:  parent,
		path:    parent.Path().Child(name),
		service: service,
	}
}

func (n *ServiceNode) Path() tree.Path {
	return n.path
}

func (n *ServiceNode) Value() interface{} {
	return n.service
}

func (n *ServiceNode) Parent() tree.Node {
	return n.parent
}

func (n *ServiceNode) Children() map[string]tree.Node {
	return nil
}

func (n *ServiceNode) Service() *Service {
	return n.service
}
