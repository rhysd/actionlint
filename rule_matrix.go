package actionlint

import "strings"

// RuleMatrix is a rule checker to check 'matrix' field of job.
type RuleMatrix struct {
	RuleBase
}

// NewRuleMatrix creates new RuleMatrix instance.
func NewRuleMatrix() *RuleMatrix {
	return &RuleMatrix{
		RuleBase: RuleBase{
			name: "matrix",
			desc: "Checks for matrix combinations in \"matrix:\"",
		},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleMatrix) VisitJobPre(n *Job) error {
	if n.Strategy == nil || n.Strategy.Matrix == nil || n.Strategy.Matrix.Expression != nil {
		return nil
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
	return nil
}

func (rule *RuleMatrix) checkDuplicateInRow(row *MatrixRow) {
	if row.Values == nil {
		return // Give up when ${{ }} is specified
	}
	seen := make([]RawYAMLValue, 0, len(row.Values))
	for _, v := range row.Values {
		ok := true
		for _, p := range seen {
			if p.Equals(v) {
				rule.Errorf(
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

func filterMatchesYAMLValue(v, filter RawYAMLValue) bool {
	switch v := v.(type) {
	case *RawYAMLObject:
		// `exclude` and `include` filter can match to objects in matrix as subset of them (#249).
		// For example,
		//
		// matrix:
		//   os:
		//     - { name: Ubuntu, matrix: ubuntu }
		//     - { name: Windows, matrix: windows }
		//   arch:
		//     - { name: ARM, matrix: arm }
		//     - { name: Intel, matrix: intel }
		//   exclude:
		//     - os: { matrix: windows }
		//       arch: { matrix: arm }
		//
		// The `exclude` filters out `{ os: { name: Windows, matrix: windows }, arch: {name: ARM, matrix: arm } }`
		switch filter := filter.(type) {
		case *RawYAMLObject:
			for n, f := range filter.Props {
				if p, ok := v.Props[n]; !ok || !filterMatchesYAMLValue(p, f) {
					return false
				}
			}
			return true
		default:
			return false
		}
	default:
		return v.Equals(filter)
	}
}

func filterMatchesMatrixRow(row []RawYAMLValue, filter RawYAMLValue) bool {
	for _, v := range row {
		if filterMatchesYAMLValue(v, filter) {
			return true
		}
	}
	return false
}

func (rule *RuleMatrix) checkExclude(m *Matrix) {
	if m.Exclude == nil || len(m.Exclude.Combinations) == 0 || (m.Include != nil && m.Include.ContainsExpression()) {
		return
	}

	if len(m.Rows) == 0 && (m.Include == nil || len(m.Include.Combinations) == 0) {
		rule.Error(m.Pos, "\"exclude\" section exists but no matrix variation exists")
		return
	}

	vals := make(map[string][]RawYAMLValue, len(m.Rows))
	ignored := map[string]struct{}{}
Outer:
	for name, row := range m.Rows {
		if row.Expression != nil {
			ignored[name] = struct{}{}
			continue
		}
		// When some item is constructed with ${{ }} dynamically, give up checking combination values (#261)
		for _, y := range row.Values {
			if s, ok := y.(*RawYAMLString); ok && isExprAssigned(s.Value) {
				ignored[name] = struct{}{}
				continue Outer
			}
		}
		vals[name] = row.Values
	}

	if m.Include != nil {
		for _, combi := range m.Include.Combinations {
			for n, a := range combi.Assigns {
				if _, ok := ignored[n]; ok {
					continue
				}
				vs, ok := vals[n]
				if !ok {
					vals[n] = []RawYAMLValue{a.Value}
					continue
				}
				if !filterMatchesMatrixRow(vs, a.Value) {
					vals[n] = append(vs, a.Value)
				}
			}
		}
	}

	for _, combi := range m.Exclude.Combinations {
		for k, a := range combi.Assigns {
			if _, ok := ignored[k]; ok {
				continue
			}
			vs, ok := vals[k]
			if !ok {
				ss := make([]string, 0, len(vals))
				for k := range vals {
					ss = append(ss, k)
				}
				rule.Errorf(
					a.Key.Pos,
					"%q in \"exclude\" section does not exist in matrix. available matrix configurations are %s",
					k,
					sortedQuotes(ss),
				)
				continue
			}

			if filterMatchesMatrixRow(vs, a.Value) {
				continue
			}

			ss := make([]string, 0, len(vs))
			for _, v := range vs {
				ss = append(ss, v.String())
			}
			rule.Errorf(
				a.Value.Pos(),
				"value %s in \"exclude\" does not match in matrix %q combinations. possible values are %s",
				a.Value.String(),
				k,
				strings.Join(ss, ", "), // Note: do not use quotesBuilder
			)
		}
	}
}
