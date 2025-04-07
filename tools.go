package qweb

import (
	"encoding/xml"
	"github.com/casbin/govaluate"
	"github.com/pkg/errors"
	"regexp"
)

func Eval[T any](ctx RenderContext, value string) (T, error) {
	var zero T
	functions := map[string]govaluate.ExpressionFunction{
		"NewLine": func(arguments ...interface{}) (interface{}, error) {
			return "\n", nil
		},
		"not": func(arguments ...interface{}) (interface{}, error) {
			return !arguments[0].(bool), nil
		},
	}
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(value, functions)
	if err != nil {
		return zero, errors.WithStack(err)
	}
	response, err := expr.Evaluate(ctx)
	if err != nil {
		return zero, errors.WithStack(err)
	}
	return response.(T), nil
}

func QAttrs2Attrs(input Attrs) []xml.Attr {
	attrs := make([]xml.Attr, 0)
	for k, v := range input {
		attrs = append(attrs, xml.Attr{
			Name:  xml.Name{Local: k},
			Value: v,
		})
	}
	return attrs
}

func GetNodeIndent(node *Node) string {
	items := node.Parent.Value().Nodes
	for idx, n := range items {
		if n != node {
			continue
		}
		prevNode := items[idx-1]
		if !(prevNode.Name == "::text" && prevNode.Content != "") {
			continue
		}
		pattern := regexp.MustCompile(`(\n +)`)
		match := pattern.FindStringSubmatch(prevNode.Content)
		if len(match) != 2 {
			continue
		}
		return match[1]
	}
	return ""
}

func RemoveLineBreak(node *Node) bool {
	nodes := &node.Nodes
	nodesSize := len(*nodes)
	if nodesSize == 0 {
		return false
	}
	lastNode := (*nodes)[nodesSize-1]
	if IsLineBreakNode(lastNode) {
		*nodes = node.Nodes[:nodesSize-1]
		return true
	}
	return false
}

func IsLineBreakNode(node *Node) bool {
	isText := node.Name == "::text"
	if !isText {
		return false
	}
	pattern := regexp.MustCompile(`\n +`)
	return pattern.MatchString(node.Content)
}
