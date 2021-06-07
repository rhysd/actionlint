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

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleMatrix) VisitJobPre(n *Job) {
	if n.Strategy == nil || n.Strategy.Matrix == nil {
		return
	}

	m := n.Strategy.Matrix

	for _, row := range m.Rows {
		rule.checkDuplicateInRow(row)
	}

	// Note:
	// Any new value can be set in the section as new combination. It can add new value to existing
	// column also.
	//
	// matrix:
	//   os: [ubuntu-latest, macos-latest]
	//   include:
	//     - os: windows-latest
	//       sh: pwsh

	rule.checkExclude(m)
}

func (rule *RuleMatrix) checkDuplicateInRow(row *MatrixRow) {
	seen := make([]RawYAMLValue, 0, len(row.Values))
	for _, v := range row.Values {
		ok := true
		for _, p := range seen {
			if p.Equals(v) {
				rule.errorf(
					v.Pos(),
					"duplicate value %s is found in matrix %q. the same value is at %s",
					v.String(),
					row.Name.Value,
					p.Pos().String(),
				)
				ok = false
				break
			}
		}
		if ok {
			seen = append(seen, v)
		}
	}
}

func findYAMLValueInArray(heystack []RawYAMLValue, needle RawYAMLValue) bool {
	for _, v := range heystack {
		if v.Equals(needle) {
			return true
		}
	}
	return false
}

func (rule *RuleMatrix) checkExclude(m *Matrix) {
	if len(m.Exclude) == 0 {
		return
	}

	rows := m.Rows
	if len(rows) == 0 && len(m.Include) == 0 {
		rule.error(m.Pos, "\"exclude\" section exists but no matrix variation exists")
		return
	}

	vals := make(map[string][]RawYAMLValue, len(rows))
	for name, row := range rows {
		vals[name] = row.Values
	}
	for _, cs := range m.Include {
		for n, c := range cs {
			vs, ok := vals[n]
			if !ok {
				vals[n] = []RawYAMLValue{c.Value}
				continue
			}
			if !findYAMLValueInArray(vs, c.Value) {
				vals[n] = append(vs, c.Value)
			}
		}
	}

	for _, cs := range m.Exclude {
		for k, c := range cs {
			vs, ok := vals[k]
			if !ok {
				qs := make([]string, 0, len(vals))
				for k := range vals {
					qs = append(qs, strconv.Quote(k))
				}
				sort.Strings(qs)
				rule.errorf(
					c.Key.Pos,
					"%q in \"exclude\" section does not exist in matrix. available matrix configurations are %s",
					k,
					strings.Join(qs, ", "),
				)
				continue
			}

			if findYAMLValueInArray(vs, c.Value) {
				continue
			}

			qs := make([]string, 0, len(vs))
			for _, v := range vs {
				qs = append(qs, v.String())
			}
			sort.Strings(qs)
			rule.errorf(
				c.Value.Pos(),
				"value %s in \"exclude\" does not exist in matrix %q combinations. possible values are %s",
				c.Value.String(),
				k,
				strings.Join(qs, ", "),
			)
		}
	}
}
