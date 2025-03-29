package qweb

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

type Asset struct {
	Template string
	Output   string
}

func readTestAsset(name string) *Asset {
	out, err := os.ReadFile(fmt.Sprintf("./assets/test/%s/out.xml", name))
	if err != nil {
		panic(err)
	}
	tmpl, err := os.ReadFile(fmt.Sprintf("./assets/test/%s/tmpl.xml", name))
	if err != nil {
		panic(err)
	}
	asset := &Asset{
		Template: string(tmpl),
		Output:   string(out),
	}
	return asset
}

func compareOutput(t *testing.T, asset *Asset, val string) {
	if val != asset.Output {
		t.Errorf("Output is:\n%s\nExpected:\n%s", val, asset.Output)
		return
	}
}

func render(t *testing.T, name string, ctx *RenderContext, comment bool) bool {
	asset := readTestAsset(name)
	inpBuffer := bytes.NewBufferString(asset.Template)
	rootNode := new(Node)
	if err := Parse(inpBuffer, rootNode); err != nil {
		t.Errorf("%+v", err)
		return false
	}
	out, err := RenderString(rootNode, *ctx)
	if err != nil {
		t.Errorf("%+v", err)
		return false
	}
	if !comment {
		out = RemoveComment(out)
	}
	compareOutput(t, asset, out)
	return true
}

func TestRender(t *testing.T) {
	t.Run("t-out", func(t *testing.T) {
		ctx := &RenderContext{
			"value":      "Test",
			"emptyValue": "",
		}
		render(t, "t-out", ctx, false)
	})
	t.Run("t-if", func(t *testing.T) {
		ctx := &RenderContext{
			"show":  true,
			"items": []string{"a", "b"},
		}
		render(t, "t-if", ctx, false)
	})
	t.Run("t-foreach", func(t *testing.T) {
		ctx := &RenderContext{
			"items": []interface{}{"a", "b", "c"},
		}
		render(t, "t-foreach", ctx, false)
	})
}
