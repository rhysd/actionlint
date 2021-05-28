package actionlint

import (
	"strconv"
	"strings"
)

// RuleMatrix is a rule checker to check 'matrix' field of job.
type RuleMatrix struct {
	RuleBase
}

// NewRuleMatrix creates new RuleMatrix instance.
func NewRuleMatrix() *RuleMatrix {
	return &RuleMatrix{}
}

func (rule *RuleMatrix) VisitJobPre(n *Job) {
	if n.Strategy == nil || n.Strategy.Matrix == nil {
		return
	}

	m := n.Strategy.Matrix

	for _, row := range m.Rows {
		seen := make(map[string]struct{}, len(row.Values))
		for _, v := range row.Values {
			if _, ok := seen[v.Value]; ok {
				rule.errorf(v.Pos, "duplicate value %q in matrix %q", v.Value, row.Name.Value)
			}
			seen[v.Value] = struct{}{}
		}
	}

	rule.checkIncludeExclude("include", m.Include, m)
	rule.checkIncludeExclude("exclude", m.Exclude, m)
}

func (rule *RuleMatrix) checkIncludeExclude(sec string, cfgs []map[string]*MatrixCombination, m *Matrix) {
	if len(cfgs) == 0 {
		return
	}

	rows := m.Rows
	if len(rows) == 0 {
		rule.errorf(m.Pos, "%q section exists but no matrix variation exists", sec)
		return
	}

	for _, cfg := range cfgs {
		for n, c := range cfg {
			r, ok := rows[n]
			if !ok {
				qs := make([]string, 0, len(rows))
				for k := range rows {
					qs = append(qs, strconv.Quote(k))
				}
				rule.errorf(
					c.Key.Pos,
					"%q in %q section does not exist in matrix. available matrix configurations are %s",
					n,
					sec,
					strings.Join(qs, ", "),
				)
				continue
			}
			if !findStringInSequence(r.Values, c.Value) {
				qs := make([]string, 0, len(r.Values))
				for _, s := range r.Values {
					qs = append(qs, strconv.Quote(s.Value))
				}
				rule.errorf(
					c.Value.Pos,
					"%q in %q section but it does not exist in matrix %q configuration. available values are %s",
					c.Value.Value,
					sec,
					r.Name.Value,
					strings.Join(qs, ", "),
				)
			}
		}
	}
}

func findStringInSequence(heystack []*String, needle *String) bool {
	for _, s := range heystack {
		if s.Value == needle.Value {
			return true
		}
	}
	return false
}
