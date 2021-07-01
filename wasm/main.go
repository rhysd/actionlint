package main

import (
	"io/ioutil"
	"syscall/js"

	"github.com/rhysd/actionlint"
)

var (
	document = js.Global().Get("document")
	window   = js.Global().Get("window")
)

func fail(err error, when string) {
	window.Call("onCheckFailed", err.Error()+" on "+when)
}

func encodeErrorAsValue(err *actionlint.Error) js.Value {
	obj := make(map[string]interface{}, 4)
	obj["message"] = err.Message
	obj["line"] = err.Line
	obj["column"] = err.Column
	obj["kind"] = err.Kind
	return js.ValueOf(obj)
}

func onButtonClick(this js.Value, args []js.Value) interface{} {
	source := window.Call("getYamlSource").String()

	opts := actionlint.LinterOptions{}
	linter, err := actionlint.NewLinter(ioutil.Discard, &opts)
	if err != nil {
		fail(err, "creating linter instance")
		return nil
	}

	errs, err := linter.Lint("test.yaml", []byte(source), nil)
	if err != nil {
		fail(err, "applying lint rules")
		return nil
	}

	ret := make([]interface{}, 0, len(errs))
	for _, err := range errs {
		ret = append(ret, encodeErrorAsValue(err))
	}

	window.Call("onCheckCompleted", js.ValueOf(ret))

	return nil
}

func main() {
	cb := js.FuncOf(onButtonClick)
	document.Call("getElementById", "button").Call("addEventListener", "click", cb)
	select {}
}
