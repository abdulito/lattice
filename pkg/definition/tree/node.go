package tree

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

// The Node interface represents a Node in the tree of a System definition.
// Note that Nodes are assumed to have an Immutable location in the tree,
// i.e. their parent and children will not change.
type Node interface {
	definition.Interface
	Parent() Node
	Path() NodePath
	Subsystems() map[NodePath]Node
	Services() map[NodePath]*ServiceNode
	NodePools() map[string]block.NodePool
}

func NewNode(d definition.Interface, parent Node) (Node, error) {
	var node Node

	// Dispatch to create proper node
	switch definitionType := d.Type(); definitionType {
	case definition.TypeSystem:
		system, ok := d.(definition.System)
		if !ok {
			return nil, fmt.Errorf("definition.Interface with Type() %v was not a definition.System", definition.TypeSystem)
		}

		systemNode, err := NewSystemNode(system, parent)
		if err != nil {
			return nil, err
		}

		node = Node(systemNode)

	case definition.TypeService:
		service, ok := d.(definition.Service)
		if !ok {
			fmt.Printf("%#v\n", d)
			return nil, fmt.Errorf("definition.Interface with Type() %v was not a definitions.Service", definition.TypeService)
		}

		serviceNode, err := NewServiceNode(service, parent)
		if err != nil {
			return nil, err
		}

		node = Node(serviceNode)

	default:
		return nil, fmt.Errorf("invalid Type() %v", definitionType)
	}

	return node, nil
}

func Walk(n Node, fn func(Node) error) error {
	err := fn(n)
	if err != nil {
		return fmt.Errorf("error walking node %v: %v", n.Path().String(), err)
	}

	for _, subsystem := range n.Subsystems() {
		err := Walk(subsystem, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPath(parent Node, definition definition.Interface) NodePath {
	parentPath := ""
	if parent != nil {
		parentPath = string(parent.Path())
	}

	return NodePath(fmt.Sprintf("%v/%v", parentPath, definition.Name()))
}
