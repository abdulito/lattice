package v1

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/definition2/resource"
	"github.com/mlab-lattice/lattice/pkg/definition2/tree"
)

type SystemNode struct {
	parent tree.Node
	path   tree.NodePath

	system *System

	resources map[string]tree.Node
	systems   map[string]*SystemNode
	services  map[string]*ServiceNode
	jobs      map[string]*JobNode
}

func NewSystemNode(system *System, name string, parent tree.Node) (*SystemNode, error) {
	path := tree.RootNodePath()
	if parent != nil {
		path = tree.ChildNodePath(parent.Path(), name)
	}

	node := &SystemNode{
		parent: nil,
		path:   path,

		system: system,

		resources: make(map[string]tree.Node),
		systems:   make(map[string]*SystemNode),
		services:  make(map[string]*ServiceNode),
		jobs:      make(map[string]*JobNode),
	}

	for n, r := range system.Resources {
		resourceNode, err := NewNode(r, n, node)
		if err != nil {
			return nil, err
		}

		node.resources[n] = resourceNode

		switch typedNode := resourceNode.(type) {
		case *JobNode:
			node.jobs[n] = typedNode

		case *ServiceNode:
			node.services[n] = typedNode

		case *SystemNode:
			node.systems[n] = typedNode

		default:
			return nil, fmt.Errorf("unrecognized node type")
		}
	}

	return node, nil
}

func (n *SystemNode) Parent() tree.Node {
	return n.parent
}

func (n *SystemNode) Path() tree.NodePath {
	return n.path
}

func (n *SystemNode) Resource() resource.Interface {
	return n.system
}

func (n *SystemNode) System() *System {
	return n.system
}

func (n *SystemNode) Resources() map[string]tree.Node {
	return n.resources
}

func (n *SystemNode) Jobs() map[string]*JobNode {
	return n.jobs
}

func (n *SystemNode) Services() map[string]*ServiceNode {
	return n.services
}

func (n *SystemNode) Systems() map[string]*SystemNode {
	return n.systems
}
