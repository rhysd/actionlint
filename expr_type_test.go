package actionlint

import "testing"

func TestExprTypeEquals(t *testing.T) {
	testCases := []struct {
		what    string
		ty      ExprType
		other   ExprType
		otherEq ExprType
	}{
		{
			what:  "null",
			ty:    NullType{},
			other: StringType{},
		},
		{
			what:  "number",
			ty:    NumberType{},
			other: StringType{},
		},
		{
			what:  "bool",
			ty:    BoolType{},
			other: StringType{},
		},
		{
			what:  "string",
			ty:    StringType{},
			other: BoolType{},
		},
		{
			what:  "object",
			ty:    NewObjectType(),
			other: &ArrayType{Elem: AnyType{}},
		},
		{
			what:  "strict props object",
			ty:    NewStrictObjectType(),
			other: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"bar": StringType{},
						},
					},
				},
			},
			other: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested strict props object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"bar": StringType{},
						},
						StrictProps: true,
					},
				},
				StrictProps: true,
			},
			other: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested object prop name",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			other: &ObjectType{
				Props: map[string]ExprType{
					"bar": StringType{},
				},
				StrictProps: true,
			},
		},
		{
			what: "nested object prop type",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			other: &ObjectType{
				Props: map[string]ExprType{
					"foo": BoolType{},
				},
				StrictProps: true,
			},
		},
		{
			what: "strict props object and loose object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
				StrictProps: true,
			},
			otherEq: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
			},
		},
		{
			what: "loose object and strict props object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
			},
			otherEq: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
			},
		},
		{
			what:  "array",
			ty:    &ArrayType{Elem: StringType{}},
			other: NewObjectType(),
		},
		{
			what:  "array element type",
			ty:    &ArrayType{Elem: StringType{}},
			other: &ArrayType{Elem: BoolType{}},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{Elem: StringType{}},
			},
			other: &ArrayType{
				Elem: &ArrayType{Elem: BoolType{}},
			},
		},
		{
			what:  "array deref",
			ty:    &ArrayDerefType{Elem: StringType{}},
			other: NewObjectType(),
		},
		{
			what:  "array deref element type",
			ty:    &ArrayDerefType{Elem: StringType{}},
			other: &ArrayDerefType{Elem: BoolType{}},
		},
		{
			what: "nested array derefs",
			ty: &ArrayDerefType{
				Elem: &ArrayDerefType{Elem: StringType{}},
			},
			other: &ArrayDerefType{
				Elem: &ArrayDerefType{Elem: BoolType{}},
			},
		},
		{
			what: "mix of array derefs and arrays",
			ty: &ArrayType{
				Elem: &ArrayDerefType{Elem: StringType{}},
			},
			otherEq: &ArrayDerefType{
				Elem: &ArrayType{Elem: StringType{}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			var l, r ExprType

			l, r = tc.ty, tc.ty
			if !l.Equals(r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			if tc.other != nil {
				l, r = tc.ty, tc.other
				if l.Equals(r) {
					t.Errorf("%s should not equal to %s", l.String(), r.String())
				}
			}
			if tc.otherEq != nil {
				l, r = tc.ty, tc.otherEq
				if !l.Equals(r) {
					t.Errorf("%s should not equal to %s", l.String(), r.String())
				}
			}
			l, r = tc.ty, AnyType{}
			if !l.Equals(r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			l, r = AnyType{}, tc.ty
			if !l.Equals(r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
		})
	}
}
