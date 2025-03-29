package main

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"weak"
)

type NodeAttr struct {
	Name  string
	Value string
}

type Node struct {
	Parent   weak.Pointer[Node]
	Name     string
	Content  string
	Attrs    []*NodeAttr
	TAttrs   []*NodeAttr
	Nodes    []*Node
	IsShadow bool
}

func Parse(r io.Reader) (*Node, error) {
	dec := xml.NewDecoder(r)
	level := 0
	root := &Node{}
	rootMap := map[int][]*Node{}
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		_, rootOk := rootMap[level]
		if !rootOk && level > 0 {
			rootMap[level] = make([]*Node, 0, 1)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			level += 1
			attrs := make([]*NodeAttr, 0, len(tok.Attr))
			tAttrs := make([]*NodeAttr, 0)
			for attrId := range tok.Attr {
				attr := &NodeAttr{
					Name:  tok.Attr[attrId].Name.Local,
					Value: tok.Attr[attrId].Value,
				}
				attrs = append(attrs, attr)
				if strings.HasPrefix(attr.Name, "t-") {
					tAttrs = append(tAttrs, attr)
				}
			}
			node := &Node{
				Name:     tok.Name.Local,
				Attrs:    attrs,
				TAttrs:   tAttrs,
				IsShadow: tok.Name.Local == "t",
			}
			if level == 1 {
				root = node
			}
			rootMap[level] = append(rootMap[level], node)
			break
		case xml.CharData:
			text := strings.TrimSpace(string(tok))
			if len(text) == 0 {
				continue
			}
			parent := rootMap[level][len(rootMap[level])-1]
			node := &Node{Content: text, Parent: weak.Make(parent)}
			parent.Nodes = append(parent.Nodes, node)
		case xml.EndElement:
			if level-1 == 0 {
				continue
			}
			parent := rootMap[level-1][len(rootMap[level-1])-1]
			node := rootMap[level][len(rootMap[level])-1]
			node.Parent = weak.Make(parent)
			parent.Nodes = append(parent.Nodes, node)
			level -= 1
			break
		}
	}
	return root, nil
}
