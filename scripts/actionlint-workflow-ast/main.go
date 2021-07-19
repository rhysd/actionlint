package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kr/pretty"
	"github.com/rhysd/actionlint"
)

func main() {
	var src []byte
	var err error
	if len(os.Args) <= 1 {
		src, err = ioutil.ReadAll(os.Stdin)
	} else {
		if os.Args[1] == "-h" || os.Args[1] == "-help" || os.Args[1] == "--help" {
			fmt.Println("Usage: go run ./scripts/actionlint-workflow-ast {workflow_file}")
			os.Exit(0)
		}
		src, err = ioutil.ReadFile(os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w, errs := actionlint.Parse(src)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	pretty.Println(w)
}
