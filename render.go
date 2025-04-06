package qweb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"maps"
	"reflect"
	"strings"
	"weak"
)

const (
	TOut     = "t-out"
	TForeach = "t-foreach"
	TAs      = "t-as"
	TIf      = "t-if"
	TAttr    = "t-att"
)

type RenderContext map[string]any

type RenderResponse struct {
	Pass     bool
	Skip     bool
	Rendered bool
	Error    error
}

func normalizeTextNodeIndent(src *Node, dst *Node) {
	if src == nil {
		return
	}
	if dst == nil {
		return
	}
	srcNodes := src.Parent.Value().Nodes
	dstNodes := dst.Parent.Value().Nodes
	if !(len(srcNodes) > 0 && len(dstNodes) > 0) {
		return
	}
	lastSrcNode := srcNodes[len(srcNodes)-1]
	lastDstNode := dstNodes[len(dstNodes)-1]
	if lastSrcNode.Name == "::text" && lastDstNode.Name == "::text" {
		dstNodes[len(dstNodes)-1].Content = lastSrcNode.Content
	}
}

func renderOut(ctx RenderContext, src *Node, dst *Node) RenderResponse {
	attrValue, ok := src.TAttrs[TOut]
	if !ok {
		return RenderResponse{
			Pass: true,
		}
	}
	var fineValue string
	val, err := Eval[any](ctx, attrValue)
	if err != nil {
		return RenderResponse{
			Error: err,
		}
	}
	if reflect.TypeOf(val).Kind() == reflect.String {
		fineValue = val.(string)
	} else {
		fineValue = fmt.Sprintf("%v", val)
	}
	if fineValue == "" && src.Name == "t" {
		normalizeTextNodeIndent(src, dst)
		return RenderResponse{
			Skip: true,
		}
	} else if fineValue == "" {
		return RenderResponse{
			Pass: true,
		}
	}
	if src.Name == "t" {
		newNode := &Node{
			Name:    "::text",
			Content: fineValue,
		}
		newNode.Copy(dst)
		return RenderResponse{Rendered: true}
	}
	dst.Content = fineValue
	return RenderResponse{Rendered: true}
}

func renderForeach(ctx RenderContext, src *Node, dst *Node) RenderResponse {
	if !(src.TAttrs.Has(TForeach) && src.TAttrs.Has(TAs)) {
		return RenderResponse{Pass: true}
	}
	each, err := Eval[any](ctx, src.TAttrs[TForeach])
	if err != nil {
		return RenderResponse{Error: err}
	}
	var items []interface{}
	switch reflect.TypeOf(each).Kind() {
	case reflect.Slice:
		items = each.([]interface{})
		break
	case reflect.Float64:
		val := int(each.(float64))
		items = make([]interface{}, val)
		for i := range val {
			items[i] = i
		}
		break
	default:
		return RenderResponse{
			Error: errors.Errorf("invalid type for foreach: %v", reflect.TypeOf(each).Kind()),
		}
	}
	loopCtx := maps.Clone(ctx)
	for idx, item := range items {
		loopCtx[src.TAttrs[TAs]] = item
		newNode := src.Clone()
		delete(newNode.TAttrs, TAs)
		delete(newNode.TAttrs, TForeach)
		newChildNode, stop, err := render(loopCtx, newNode, newNode)
		if err != nil {
			return RenderResponse{Error: err}
		}
		if stop {
			break
		}
		if newChildNode == nil {
			continue
		}
		parentNodes := &dst.Parent.Value().Nodes
		if newChildNode.Name == "t" {
			newChildNodesSize := len(newChildNode.Nodes)
			if newChildNodesSize > 0 {
				start := 1
				end := newChildNodesSize
				if idx == len(items)-1 {
					end = newChildNodesSize - 1
				}
				*parentNodes = append(*parentNodes, newChildNode.Nodes[start:end]...)
			}
		} else {
			*parentNodes = append(*parentNodes, newChildNode)
		}
	}
	return RenderResponse{Rendered: true}
}

func renderAttr(ctx RenderContext, src *Node, dst *Node) error {
	for name, attr := range src.TAttrs {
		if !strings.HasPrefix(name, TAttr) {
			continue
		}
		attrName := strings.Split(name, TAttr+"-")[1]
		val, err := Eval[any](ctx, attr)
		if err != nil {
			return err
		}
		var fineValue string
		switch reflect.TypeOf(val).Kind() {
		case reflect.String:
			fineValue = val.(string)
			break
		case reflect.Float64:
			fineValue = fmt.Sprintf("%v", val)
			break
		default:
			return errors.Errorf("invalid type for attr: %v", reflect.TypeOf(val).Kind())
		}
		dst.Attrs[attrName] = fineValue
	}
	return nil
}

func checkTIf(ctx RenderContext, node *Node) (bool, error) {
	attrValue, ok := node.TAttrs[TIf]
	if !ok {
		return true, nil
	}
	val, err := Eval[any](ctx, attrValue)
	if err != nil {
		return false, err
	}
	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		return len(val.(string)) != 0, nil
	case reflect.Float64:
		return val.(float64) != 0, nil
	case reflect.Bool:
		return val.(bool), nil
	default:
		return false, errors.Errorf("invalid type for if: %v", reflect.TypeOf(val).Kind())
	}
}

func xmlWrite(encoder *xml.Encoder, node *Node, bodyCB func(*Node) error) error {
	startElement := xml.StartElement{
		Name: xml.Name{Local: node.Name},
		Attr: QAttrs2Attrs(node.Attrs),
	}
	// Write start element
	if err := encoder.EncodeToken(startElement); err != nil {
		return errors.WithStack(err)
	}
	if err := bodyCB(node); err != nil {
		return err
	}
	// Write end element
	if err := encoder.EncodeToken(startElement.End()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func render(ctx RenderContext, src *Node, parent *Node) (*Node, bool, error) {
	if valid, err := checkTIf(ctx, src); err != nil || !valid {
		if parent != nil {
			normalizeTextNodeIndent(src, &Node{
				Parent: weak.Make(parent),
			})
		}
		return nil, true, err
	}
	currentNode := &Node{
		Name:    src.Name,
		Attrs:   src.Attrs,
		Content: src.Content,
		Nodes:   []*Node{},
	}
	if parent != nil {
		currentNode.Parent = weak.Make(parent)
	}
	var response RenderResponse
	response = renderForeach(ctx, src, currentNode)
	if response.Error != nil {
		return nil, true, response.Error
	}
	if response.Rendered {
		return nil, false, nil
	}
	if err := renderAttr(ctx, src, currentNode); err != nil {
		return nil, true, err
	}
	response = renderOut(ctx, src, currentNode)
	if response.Error != nil {
		return nil, true, response.Error
	}
	if response.Skip {
		return nil, true, nil
	}
	if response.Rendered {
		return currentNode, false, nil
	}
	for _, childNode := range src.Nodes {
		node, stop, err := render(ctx, childNode, currentNode)
		if err != nil {
			return nil, true, err
		}
		if stop {
			break
		}
		if node == nil {
			continue
		}
		node.Parent = weak.Make(currentNode)
		currentNode.Nodes = append(currentNode.Nodes, node)
	}
	return currentNode, false, nil
}

func Render(ctx RenderContext, src *Node) (*Node, error) {
	node, _, err := render(ctx, src, nil)
	return node, err
}

func renderString(encoder *xml.Encoder, root *Node) error {
	if root.Name == "::text" {
		if err := encoder.EncodeToken(xml.CharData(root.Content)); err != nil {
			return err
		}
		return nil
	}
	if root.Name == "::comment" {
		if err := encoder.EncodeToken(xml.Comment(root.Content)); err != nil {
			return err
		}
		return nil
	}
	err := xmlWrite(encoder, root, func(node *Node) error {
		if node.Content != "" {
			if err := encoder.EncodeToken(xml.CharData(root.Content)); err != nil {
				return err
			}
			return nil
		}
		for _, childNode := range node.Nodes {
			if err := renderString(encoder, childNode); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func RenderString(ctx RenderContext, src *Node) (value string, fErr error) {
	w := &bytes.Buffer{}
	encoder := xml.NewEncoder(w)
	defer func() {
		if err := encoder.Close(); err != nil && fErr == nil {
			fErr = err
		}
	}()
	root, err := Render(ctx, src)
	if err != nil {
		return "", err
	}
	if err := renderString(encoder, root); err != nil {
		return "", err
	}
	if err := encoder.Flush(); err != nil {
		return "", err
	}
	return w.String(), nil
}
