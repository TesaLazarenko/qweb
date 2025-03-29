package qweb

import (
	"github.com/pkg/errors"
	"maps"
	"weak"
)

var UndefinedNodeError = errors.New("node is undefined")

type Attrs map[string]string

func (na *Attrs) Has(key string) bool {
	_, ok := (*na)[key]
	return ok
}

type Node struct {
	Parent  weak.Pointer[Node]
	Name    string
	Content string
	Attrs   Attrs
	TAttrs  Attrs
	Nodes   []*Node
}

func (node *Node) Copy(target *Node) {
	if target == nil {
		panic(UndefinedNodeError)
	}
	target.Name = node.Name
	target.Content = node.Content
	target.TAttrs = maps.Clone(node.TAttrs)
	target.Attrs = maps.Clone(node.Attrs)
	target.Nodes = make([]*Node, len(node.Nodes))
	for idx, current := range node.Nodes {
		target.Nodes[idx] = current.Clone()
		target.Nodes[idx].Parent = weak.Make(target)
	}
}

func (node *Node) Clone() *Node {
	copyNode := new(Node)
	node.Copy(copyNode)
	return copyNode
}

func (node *Node) Prev() (int, *Node, error) {
	nodeItems := node.Parent.Value().Nodes
	for idx, item := range nodeItems {
		if item != node {
			continue
		}
		if idx-1 < 0 {
			return 0, nil, UndefinedNodeError
		}
		return idx - 1, nodeItems[idx-1], nil
	}
	return 0, nil, UndefinedNodeError
}

func (node *Node) Next() (int, *Node, error) {
	nodeItems := node.Parent.Value().Nodes
	for idx, item := range nodeItems {
		if item != node {
			continue
		}
		if idx+1 > len(nodeItems) {
			return 0, nil, UndefinedNodeError
		}
		return idx + 1, nodeItems[idx+1], nil
	}
	return 0, nil, UndefinedNodeError
}
