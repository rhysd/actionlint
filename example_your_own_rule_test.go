package actionlint_test

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/rhysd/actionlint"
)

// A rule type to check every steps have their names.
type RuleStepName struct {
	// Embedding RuleBase struct implements the minimal Rule interface.
	actionlint.RuleBase
}

// Reimplement methods in RuleBase. Visit* methods are called on checking workflows.
func (r *RuleStepName) VisitStep(n *actionlint.Step) error {
	// Implement your own check
	if n.Name == nil {
		// RuleBase provides methods to report errors. See RuleBase.Error and RuleBase.Errorf.
		r.Error(n.Pos, "every step must have its name")
	}
	return nil
}

func NewRuleStepName() *RuleStepName {
	return &RuleStepName{
		RuleBase: actionlint.NewRuleBase("step-name", "Checks every step has their own name"),
	}
}

func ExampleLinter_yourOwnRule() {
	// The function set at OnRulesCreated is called after rule instances are created. You can
	// add/remove some rules and return the modified slice. This function is called on linting
	// each workflow files.
	o := &actionlint.LinterOptions{
		OnRulesCreated: func(rules []actionlint.Rule) []actionlint.Rule {
			rules = append(rules, NewRuleStepName())
			return rules
		},
	}

	l, err := actionlint.NewLinter(io.Discard, o)
	if err != nil {
		panic(err)
	}

	f := filepath.Join("testdata", "ok", "minimal.yaml")

	// First return value is an array of lint errors found in the workflow file.
	errs, err := l.LintFile(f, nil)
	if err != nil {
		panic(err)
	}

	// `errs` includes errors like below:
	//
	// testdata/examples/main.yaml:14:9: every step must have its name [step-name]
	//    |
	// 14 |       - uses: actions/checkout@v3
	//    |         ^~~~~
	fmt.Println(len(errs), "lint errors found by actionlint")
	// Output: 1 lint errors found by actionlint
}
