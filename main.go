package main

import (
	"errors"
	"github.com/casbin/govaluate"
	"os"
	"reflect"
)

type QAttrType string

const (
	TOut     = "t-out"
	TForeach = "t-foreach"
	TIf      = "t-if"
)

func TIfRender(ctx RenderContext, value string) (bool, error) {
	expr, err := govaluate.NewEvaluableExpression(value)
	if err != nil {
		return false, err
	}
	response, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	responseType := reflect.TypeOf(response).Kind()
	if responseType != reflect.Bool {
		return false, errors.New("not a bool")
	}
	responseValue := response.(bool)
	return responseValue, nil
}

func TOutRender(ctx RenderContext, value string) (string, error) {
	expr, err := govaluate.NewEvaluableExpression(value)
	if err != nil {
		return "", err
	}
	response, err := expr.Evaluate(ctx)
	if err != nil {
		return "", err
	}
	responseType := reflect.TypeOf(response).Kind()
	if responseType != reflect.String {
		return "", errors.New("not a string")
	}
	responseValue := response.(string)
	return responseValue, nil
}

type RenderContext map[string]any

func Render(root *Node) (string, error) {
	return "", nil
}

func main() {
	file, err := os.Open("./example.xml")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()
	rawData, err := Parse(file)
	if err != nil {
		panic(err)
	}
	// ctx := RenderContext{
	// 	"item":  "Taras",
	// 	"item2": "None",
	// }
	v, err := Render(rawData)
	if err != nil {
		panic(err)
	}
	println(v)
	// rsp, _ := RenderQWeb(ctx, *rawData)
	// v, err := xml.Marshal(rsp)
	// if err != nil {
	// 	panic(err)
	// }
	// val := string(v)
	// println(val)
}
