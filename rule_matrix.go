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

func isYAMLValueSubset(v, sub RawYAMLValue) bool {
	// When the filter side is dynamically constructed with some expression, it is not possible to statically check if the filter
	// matches the value. To avoid false positives, assume such filter always matches to the value. (#414)
	// ```
	// matrix:
	//   foo: ['a', 'b']
	//   exclude:
	//     foo: ${{ fromJSON('...') }}
	// ```
	if s, ok := sub.(*RawYAMLString); ok && ContainsExpression(s.Value) {
		return true
	}

	switch v := v.(type) {
	case *RawYAMLObject:
		// `exclude` filter can match to objects in matrix as subset of them (#249).
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
		sub, ok := sub.(*RawYAMLObject)
		if !ok {
			return false
		}
		for n, s := range sub.Props {
			if p, ok := v.Props[n]; !ok || !isYAMLValueSubset(p, s) {
				return false
			}
		}
		return true
	case *RawYAMLArray:
		sub, ok := sub.(*RawYAMLArray)
		if !ok {
			return false
		}
		if len(v.Elems) != len(sub.Elems) {
			return false
		}
		for i, v := range v.Elems {
			if !isYAMLValueSubset(v, sub.Elems[i]) {
				return false
			}
		}
		return true
	case *RawYAMLString:
		// When some item is constructed with ${{ }} dynamically, give up checking combinations (#261)
		if ContainsExpression(v.Value) {
			return true
		}
		return v.Equals(sub)
	default:
		return v.Equals(sub)
	}
}

func (rule *RuleMatrix) checkExclude(m *Matrix) {
	if m.Exclude == nil || len(m.Exclude.Combinations) == 0 || (m.Include != nil && m.Include.ContainsExpression()) {
		return
	}

	if len(m.Rows) == 0 && (m.Include == nil || len(m.Include.Combinations) == 0) {
		rule.Error(m.Pos, "\"exclude\" section exists but no matrix variation exists")
		return
	}

	rows := make(map[string][]RawYAMLValue, len(m.Rows))
	ignored := map[string]struct{}{}

	for n, r := range m.Rows {
		if r.Expression != nil {
			ignored[n] = struct{}{}
			continue
		}
		rows[n] = r.Values
	}

	if m.Include != nil {
		for _, c := range m.Include.Combinations {
		Include:
			for n, a := range c.Assigns {
				if _, ok := ignored[n]; ok {
					continue
				}
				row := rows[n]
				for _, v := range row {
					if v.Equals(a.Value) {
						continue Include
					}
				}
				rows[n] = append(row, a.Value)
			}
		}
	}

	for _, c := range m.Exclude.Combinations {
	Exclude:
		for k, a := range c.Assigns {
			if _, ok := ignored[k]; ok {
				continue
			}
			row, ok := rows[k]
			if !ok {
				ss := make([]string, 0, len(rows))
				for k := range rows {
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

			for _, v := range row {
				if isYAMLValueSubset(v, a.Value) {
					continue Exclude
				}
			}

			ss := make([]string, 0, len(row))
			for _, v := range row {
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
