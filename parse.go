package qweb

import (
	"encoding/xml"
	"github.com/pkg/errors"
	"io"
	"strings"
	"weak"
)

func Parse(r io.Reader, root *Node) error {
	dec := xml.NewDecoder(r)
	level := 0
	rootMap := map[int][]*Node{}
	stopLoop := false
	for !stopLoop {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.WithStack(err)
		}
		_, rootOk := rootMap[level]
		if !rootOk && level > 0 {
			rootMap[level] = make([]*Node, 0, 1)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			level += 1
			node := root
			if level > 1 {
				node = new(Node)
			}
			node.Name = tok.Name.Local
			node.Attrs = make(Attrs)
			node.TAttrs = make(Attrs)
			for _, attr := range tok.Attr {
				name := attr.Name.Local
				if strings.HasPrefix(name, "t-") {
					node.TAttrs[name] = attr.Value
				} else {
					node.Attrs[name] = attr.Value
				}
			}
			rootMap[level] = append(rootMap[level], node)
			break
		case xml.Comment:
			if level == 0 {
				stopLoop = true
				continue
			}
			parent := rootMap[level][len(rootMap[level])-1]
			node := &Node{
				Name:    "::comment",
				Content: string(tok),
				Parent:  weak.Make(parent),
			}
			parent.Nodes = append(parent.Nodes, node)
		case xml.CharData:
			if level == 0 {
				stopLoop = true
				continue
			}
			parent := rootMap[level][len(rootMap[level])-1]
			node := &Node{
				Name:    "::text",
				Content: string(tok),
				Parent:  weak.Make(parent),
			}
			parent.Nodes = append(parent.Nodes, node)
		case xml.EndElement:
			if level == 1 {
				stopLoop = true
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
	return nil
}
