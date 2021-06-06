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

func (rule *RuleMatrix) checkMatrixValuesContain(sec string, name string, heystack []RawYAMLValue, needle RawYAMLValue) {
	for _, v := range heystack {
		if v.Equals(needle) {
			return // found
		}
	}

	qs := make([]string, 0, len(heystack))
	for _, v := range heystack {
		qs = append(qs, v.String())
	}
	sort.Strings(qs)
	rule.errorf(
		needle.Pos(),
		"%s in %q section does not exist in matrix %q configuration. available values are %s",
		needle.String(),
		sec,
		name,
		strings.Join(qs, ", "),
	)
}

// TODO: This method has bug: it does not consider values in include section. Values can be added to
// combinations of matrix with include section. So it should be considered
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
