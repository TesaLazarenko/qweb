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
	out, err := os.ReadFile(fmt.Sprintf("./assets/test/%s/out.txt", name))
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

func testRender(t *testing.T, name string, ctx *RenderContext) bool {
	asset := readTestAsset(name)
	inpBuffer := bytes.NewBufferString(asset.Template)
	rootNode := new(Node)
	if err := Parse(inpBuffer, rootNode); err != nil {
		t.Errorf("%+v", err)
		return false
	}
	out, err := RenderString(*ctx, rootNode)
	if err != nil {
		t.Errorf("%+v", err)
		return false
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
		testRender(t, "t-out", ctx)
	})
	t.Run("t-if", func(t *testing.T) {
		ctx := &RenderContext{
			"show":  true,
			"items": []string{"a", "b"},
		}
		testRender(t, "t-if", ctx)
	})
	t.Run("t-foreach", func(t *testing.T) {
		ctx := &RenderContext{
			"items": []interface{}{"a", "b", "c"},
		}
		testRender(t, "t-foreach", ctx)
	})
	t.Run("complex_1", func(t *testing.T) {
		ctx := &RenderContext{
			"value":      "Test",
			"emptyValue": "",
			"show":       true,
			"items":      []interface{}{"a", "b", "c"},
		}
		testRender(t, "complex_1", ctx)
	})
}
