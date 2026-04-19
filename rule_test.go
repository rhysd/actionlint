package actionlint

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRuleBaseSetGetConfig(t *testing.T) {
	r := NewRuleBase("", "")
	if r.Config() != nil {
		t.Error("Config must be nil after creating a rule")
	}
	want := &Config{}
	r.SetConfig(want)
	if have := r.Config(); have != want {
		t.Errorf("Wanted config %v but got %v", want, have)
	}
}

func TestRuleBaseErrorfAndErrs(t *testing.T) {
	r := NewRuleBase("dummy name", "dummy description")
	errs := r.Errs()
	if len(errs) > 0 {
		t.Error("no error is expected", errs)
	}
	r.Error(&Pos{Line: 1, Col: 2}, "this is test 1")
	r.Errorf(&Pos{Line: 3, Col: 4}, "this is test %d", 2)
	errs = r.Errs()
	if len(errs) != 2 {
		t.Error("wanted 2 errors but have", errs)
	}
	want := []*Error{
		{
			Message: "this is test 1",
			Line:    1,
			Column:  2,
			Kind:    "dummy name",
		},
		{
			Message: "this is test 2",
			Line:    3,
			Column:  4,
			Kind:    "dummy name",
		},
	}
	if diff := cmp.Diff(errs, want); diff != "" {
		t.Error("unexpected errors from Errs() method:", diff)
	}
}

func TestRuleBaseDebugOutput(t *testing.T) {
	r := NewRuleBase("dummy-name", "")
	r.Debug("this %s output", "is not")

	b := &bytes.Buffer{}
	r.EnableDebug(b)
	r.Debug("this %s output!", "IS")

	have := b.String()
	want := "[dummy-name] this IS output!\n"
	if want != have {
		t.Errorf("wanted %q as debug output but have %q", want, have)
	}
}

func TestRuleBaseNameAndDescription(t *testing.T) {
	r := NewRuleBase("dummy name", "dummy description")
	if r.Name() != "dummy name" {
		t.Errorf("name is unexpected: %q", r.Name())
	}
	if r.Description() != "dummy description" {
		t.Errorf("description is unexpected: %q", r.Description())
	}
}
