package actionlint

import "testing"

func TestExprTypeEquals(t *testing.T) {
	testCases := []struct {
		ty    ExprType
		other ExprType
	}{
		{NullType{}, StringType{}},
		{NumberType{}, StringType{}},
		{BoolType{}, StringType{}},
		{StringType{}, BoolType{}},
	}

	for _, tc := range testCases {
		var l, r ExprType

		l, r = tc.ty, tc.ty
		if !l.Equals(r) {
			t.Errorf("%s should equal to %s", l.String(), r.String())
		}
		l, r = tc.ty, tc.other
		if l.Equals(r) {
			t.Errorf("%s should not equal to %s", l.String(), r.String())
		}
		l, r = tc.ty, AnyType{}
		if !l.Equals(r) {
			t.Errorf("%s should equal to %s", l.String(), r.String())
		}
		l, r = AnyType{}, tc.ty
		if !l.Equals(r) {
			t.Errorf("%s should equal to %s", l.String(), r.String())
		}
	}
}
