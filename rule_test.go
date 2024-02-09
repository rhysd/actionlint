package actionlint

import "testing"

func TestRuleBaseConfig(t *testing.T) {
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
