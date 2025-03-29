package qweb

import (
	"encoding/xml"
	"github.com/casbin/govaluate"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

func Eval[T any](ctx RenderContext, value string) (T, error) {
	var zero T
	functions := map[string]govaluate.ExpressionFunction{
		"NewLine": func(arguments ...interface{}) (interface{}, error) {
			buff := make([]string, int(arguments[0].(float64)))
			return strings.Join(buff, "\n"), nil
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
		if idx < 1 {
			continue
		}
		if n != node {
			continue
		}
		prevNode := items[idx-1]
		if !(prevNode.Name == "::text" && prevNode.Content != "") {
			continue
		}
		pattern := regexp.MustCompile(`(\n\s+)$`)
		match := pattern.FindStringSubmatch(prevNode.Content)
		if len(match) != 2 {
			continue
		}
		return match[1]
	}
	return ""
}
