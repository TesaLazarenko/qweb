package qweb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"maps"
	"reflect"
	"regexp"
	"strings"
)

const (
	TOut     = "t-out"
	TForeach = "t-foreach"
	TAs      = "t-as"
	TIf      = "t-if"
	TAttr    = "t-att"
)

type RenderContext map[string]any

type Renderer struct {
	Indent  int
	Encoder *xml.Encoder
}

func (r *Renderer) WriteCharData(node *Node) (bool, error) {
	if node.Name != "::text" {
		return false, nil
	}
	err := errors.WithStack(r.Encoder.EncodeToken(xml.CharData(node.Content)))
	return true, err
}

func (r *Renderer) WriteComment(node *Node) (bool, error) {
	if node.Name != "::comment" {
		return false, nil
	}
	err := errors.WithStack(r.Encoder.EncodeToken(xml.Comment(node.Content)))
	return true, err
}

func (r *Renderer) WriteCommentWithValue(msg string) error {
	_, err := r.WriteComment(&Node{
		Name:    "::comment",
		Content: msg,
	})
	return err
}

func (r *Renderer) WriteTOut(ctx RenderContext, node *Node) (bool, error) {
	attrValue, ok := node.TAttrs[TOut]
	if !ok {
		return false, nil
	}
	var fineValue string
	val, err := Eval[any](ctx, attrValue)
	if err != nil {
		return false, err
	}
	if reflect.TypeOf(val).Kind() == reflect.String {
		fineValue = val.(string)
	} else {
		fineValue = fmt.Sprintf("%v", val)
	}
	if fineValue == "" {
		if err := r.WriteCommentWithValue("::render: invisible"); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := r.Encoder.EncodeToken(xml.CharData(fineValue)); err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

func (r *Renderer) WriteForeach(ctx RenderContext, node *Node) (bool, error) {
	if !(node.TAttrs.Has(TForeach) && node.TAttrs.Has(TAs)) {
		return false, nil
	}
	each, err := Eval[any](ctx, node.TAttrs[TForeach])
	if err != nil {
		return false, err
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
		return false, errors.Errorf("invalid type for foreach: %v", reflect.TypeOf(each).Kind())
	}
	loopCtx := maps.Clone(ctx)
	for idx, item := range items {
		loopCtx[node.TAttrs[TAs]] = item
		err := r.Write(node, func(node *Node) error {
			rendered, err := r.WriteTOut(loopCtx, node)
			if err != nil || rendered {
				return err
			}
			for _, node := range node.Nodes {
				if err := r.RenderNode(loopCtx, node); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return false, err
		}
		if idx < len(items)-1 {
			newLineNode := &Node{
				Name:    "::text",
				Content: GetNodeIndent(node),
			}
			if _, err := r.WriteCharData(newLineNode); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (r *Renderer) WriteAttr(ctx RenderContext, node *Node) (bool, error) {
	for name, attr := range node.TAttrs {
		if !strings.HasPrefix(name, TAttr) {
			continue
		}
		attrName := strings.Split(name, TAttr+"-")[1]
		val, err := Eval[any](ctx, attr)
		if err != nil {
			return false, err
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
			return false, errors.Errorf("invalid type for attr: %v", reflect.TypeOf(val).Kind())
		}
		node.Attrs[attrName] = fineValue
	}
	return true, nil
}

func (r *Renderer) CheckTIf(ctx RenderContext, node *Node) (bool, error) {
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

func (r *Renderer) Write(node *Node, bodyCB func(*Node) error) error {
	if node.Name == "t" {
		if err := bodyCB(node); err != nil {
			return err
		}
		return nil
	}
	startElement := xml.StartElement{
		Name: xml.Name{Local: node.Name},
		Attr: QAttrs2Attrs(node.Attrs),
	}
	// Write start element
	if err := r.Encoder.EncodeToken(startElement); err != nil {
		return errors.WithStack(err)
	}
	if err := bodyCB(node); err != nil {
		return err
	}
	// Write end element
	if err := r.Encoder.EncodeToken(startElement.End()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Renderer) RenderNode(ctx RenderContext, node *Node) error {
	validCondition, err := r.CheckTIf(ctx, node)
	if err != nil {
		return err
	}
	if !validCondition {
		if err := r.WriteCommentWithValue("::render: invisible"); err != nil {
			return err
		}
		return nil
	}
	if rendered, err := r.WriteCharData(node); err != nil || rendered {
		return err
	}
	_, err = r.WriteAttr(ctx, node)
	if err != nil {
		return err
	}
	if rendered, err := r.WriteForeach(ctx, node); err != nil || rendered {
		return err
	}
	err = r.Write(node, func(node *Node) error {
		rendered, err := r.WriteTOut(ctx, node)
		if err != nil || rendered {
			return err
		}
		for _, srcNode := range node.Nodes {
			if err := r.RenderNode(ctx, srcNode); err != nil {
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

func Render(w io.Writer, root *Node, ctx RenderContext) error {
	renderer := &Renderer{
		Encoder: xml.NewEncoder(w),
	}
	defer func() {
		if err := renderer.Encoder.Close(); err != nil {
			log.Printf("%+v", err)
		}
	}()
	if err := renderer.RenderNode(ctx, root.Clone()); err != nil {
		return err
	}
	return nil
}

func RenderString(root *Node, ctx RenderContext) (string, error) {
	w := &bytes.Buffer{}
	if err := Render(w, root, ctx); err != nil {
		return "", err
	}
	return w.String(), nil
}

func RemoveComment(data string) string {
	pattern := regexp.MustCompile(`\n( +)?<!--.*?-->`)
	return pattern.ReplaceAllString(data, "")
}
