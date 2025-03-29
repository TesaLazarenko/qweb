package qweb

import (
	"testing"
	"weak"
)

func newRootNode() *Node {
	root := &Node{
		Name: "container",
	}
	root.Nodes = []*Node{
		{
			Parent: weak.Make(root),
			Name:   "item1",
		},
		{
			Parent: weak.Make(root),
			Name:   "item2",
		},
		{
			Parent: weak.Make(root),
			Name:   "item3",
		},
	}
	return root
}

func TestNode(t *testing.T) {
	t.Run("copy", func(t *testing.T) {
		rootNode := newRootNode()
		newNode := new(Node)
		rootNode.Copy(newNode)
		if newNode == nil {
			t.Error("new node should not be nil")
			return
		}
	})
	t.Run("clone", func(t *testing.T) {
		rootNode := newRootNode()
		newNode := rootNode.Clone()
		if newNode == nil {
			t.Error("new node should not be nil")
			return
		}
	})
	t.Run("prev", func(t *testing.T) {
		rootNode := newRootNode()
		target := rootNode.Nodes[1]
		pos, node, err := target.Prev()
		if err != nil {
			t.Error(err)
			return
		}
		if pos != 0 {
			t.Error("prev index should be 0")
			return
		}
		if node == nil {
			t.Error("node should not be nil")
			return
		}
	})
	t.Run("next", func(t *testing.T) {
		rootNode := newRootNode()
		target := rootNode.Nodes[1]
		pos, node, err := target.Next()
		if err != nil {
			t.Error(err)
			return
		}
		if pos != 2 {
			t.Error("next index should be 2")
			return
		}
		if node == nil {
			t.Error("node should not be nil")
			return
		}
	})
}
