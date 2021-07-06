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
	window.Call("showError", err.Error()+" on "+when)
}

func encodeErrorAsMap(err *actionlint.Error) map[string]interface{} {
	obj := make(map[string]interface{}, 4)
	obj["message"] = err.Message
	obj["line"] = err.Line
	obj["column"] = err.Column
	obj["kind"] = err.Kind
	return obj
}

func lint(source string) interface{} {
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
		ret = append(ret, encodeErrorAsMap(err))
	}

	window.Call("onCheckCompleted", js.ValueOf(ret))

	return nil
}

func runActionlint(_this js.Value, args []js.Value) interface{} {
	source := args[0].String()
	return lint(source)
}

func main() {
	window.Set("runActionlint", js.FuncOf(runActionlint))
	window.Call("dismissLoading")
	lint(window.Call("getYamlSource").String()) // Show the first result
	select {}
}
