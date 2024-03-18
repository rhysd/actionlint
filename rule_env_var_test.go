package actionlint

import "testing"

func testValidateEnvVarName(t *testing.T, name string) []*Error {
	t.Helper()
	n := &String{Value: name, Pos: &Pos{}}
	e := &Env{
		Vars: map[string]*EnvVar{
			name: {n, &String{Value: "", Pos: &Pos{}}},
		},
	}
	w := &Workflow{Env: e}
	r := NewRuleEnvVar()
	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}
	return r.Errs()
}

func TestRuleEnvVarCheckValidName(t *testing.T) {
	valids := []string{
		"foo_bar",
		"FOO_BAR",
		"foo-bar",
		"_",
		"-",
	}

	for _, n := range valids {
		t.Run(n, func(t *testing.T) {
			if errs := testValidateEnvVarName(t, n); len(errs) > 0 {
				t.Fatalf("env var %q should be valid but got errors: %v", n, errs)
			}
		})
	}
}

func TestRuleEnvVarCheckInvalidName(t *testing.T) {
	invalids := []string{
		"a b",
		"a=b",
		"a=b",
	}

	for _, n := range invalids {
		t.Run(n, func(t *testing.T) {
			if errs := testValidateEnvVarName(t, n); len(errs) == 0 {
				t.Fatalf("env var %q should be invalid but got no error", n)
			}
		})
	}
}

func TestRuleEnvVarSkipEdgeCaseEnv(t *testing.T) {
	w := &Workflow{Env: nil}
	r := NewRuleEnvVar()
	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}
	if errs := r.Errs(); len(errs) > 0 {
		t.Fatal("no error should be detected when `Env` is nil but got", errs)
	}

	w = &Workflow{
		Env: &Env{
			Expression: &String{Value: ""},
		},
	}
	r = NewRuleEnvVar()
	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}
	if errs := r.Errs(); len(errs) > 0 {
		t.Fatal("no error should be detected when `Env` is constructed by expression", errs)
	}

}
