package actionlint

import (
	"sort"
	"strconv"
	"strings"
)

// RuleMatrix is a rule checker to check 'matrix' field of job.
type RuleMatrix struct {
	RuleBase
}

// NewRuleMatrix creates new RuleMatrix instance.
func NewRuleMatrix() *RuleMatrix {
	return &RuleMatrix{
		RuleBase: RuleBase{name: "matrix"},
	}
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

	rule.checkInclude(m)
	rule.checkExclude(m)
}

func (rule *RuleMatrix) checkMatrixValuesContain(sec string, name string, heystack []*String, needle *String) {
	for _, s := range heystack {
		if s.Value == needle.Value {
			return // found
		}
	}

	qs := make([]string, 0, len(heystack))
	for _, s := range heystack {
		qs = append(qs, strconv.Quote(s.Value))
	}
	sort.Strings(qs)
	rule.errorf(
		needle.Pos,
		"%q in %q section does not exist in matrix %q configuration. available values are %s",
		needle.Value,
		sec,
		name,
		strings.Join(qs, ", "),
	)
}

func (rule *RuleMatrix) checkInclude(m *Matrix) {
	if len(m.Include) == 0 {
		return
	}

	rows := m.Rows
	if len(rows) == 0 {
		// Note: 'include' section without any matrix is still useful
		return
	}

	for _, cfg := range m.Include {
		for name, kv := range cfg {
			r, ok := rows[name]
			if !ok {
				continue
			}
			rule.checkMatrixValuesContain("include", name, r.Values, kv.Value)
		}
	}
}

func (rule *RuleMatrix) checkExclude(m *Matrix) {
	if len(m.Exclude) == 0 {
		return
	}

	rows := m.Rows
	if len(rows) == 0 {
		rule.error(m.Pos, "\"exclude\" section exists but no matrix variation exists")
		return
	}

	for _, cfg := range m.Exclude {
		for k, v := range cfg {
			r, ok := rows[k]
			if !ok {
				qs := make([]string, 0, len(rows))
				for k := range rows {
					qs = append(qs, strconv.Quote(k))
				}
				rule.errorf(
					v.Key.Pos,
					"%q in \"exclude\" section does not exist in matrix. available matrix configurations are %s",
					k,
					strings.Join(qs, ", "),
				)
				continue
			}
			rule.checkMatrixValuesContain("exclude", k, r.Values, v.Value)
		}
	}
}
