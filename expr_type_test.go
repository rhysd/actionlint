package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExprNewMapObjectType(t *testing.T) {
	o := NewMapObjectType(StringType{})
	if o.Props != nil {
		t.Fatalf("props should be nil but %v", o.Props)
	}
	if _, ok := o.Mapped.(StringType); !ok {
		t.Fatalf("mapped type is not string: %v", o.Mapped)
	}
	if o.StrictProps {
		t.Fatalf("map object is not strict object but got %v", o)
	}

	// Mapping to any type is just a loose object
	o = NewMapObjectType(AnyType{})
	if len(o.Props) != 0 {
		t.Fatalf("loose object with no prop info was expected but got props %v", o.Props)
	}
	if o.Mapped != nil {
		t.Fatalf("loose object should not have mapped type but got %v", o.Mapped)
	}
	if o.StrictProps {
		t.Fatalf("loose object is not strict object but got %v", o)
	}
}

func TestExprTypeEquals(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		neq  ExprType
		eq   ExprType
	}{
		{
			what: "null",
			ty:   NullType{},
			neq:  StringType{},
		},
		{
			what: "number",
			ty:   NumberType{},
			neq:  StringType{},
		},
		{
			what: "bool",
			ty:   BoolType{},
			neq:  StringType{},
		},
		{
			what: "string",
			ty:   StringType{},
			neq:  BoolType{},
		},
		{
			what: "object",
			ty:   NewObjectType(),
			neq:  &ArrayType{Elem: AnyType{}},
		},
		{
			what: "strict props object",
			ty:   NewStrictObjectType(),
			neq:  &ArrayType{Elem: AnyType{}},
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
			neq: &ArrayType{Elem: AnyType{}},
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
			neq: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested object prop name",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			neq: &ObjectType{
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
			neq: &ObjectType{
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
			eq: &ObjectType{
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
			eq: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
			},
		},
		{
			what: "map objects",
			ty: &ObjectType{
				Mapped: NullType{},
			},
			eq: &ObjectType{
				Mapped: NullType{},
			},
			neq: &ObjectType{
				Mapped: NumberType{},
			},
		},
		{
			what: "map object equals loose object",
			ty: &ObjectType{
				Mapped: StringType{},
			},
			eq: NewObjectType(),
		},
		{
			what: "loose object equals map object",
			ty:   NewObjectType(),
			eq: &ObjectType{
				Mapped: StringType{},
			},
		},
		{
			what: "map object equals strict object",
			ty: &ObjectType{
				Mapped: StringType{},
			},
			eq: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			neq: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
				StrictProps: true,
			},
		},
		{
			what: "map object equals strict object including any prop",
			ty: &ObjectType{
				Mapped: StringType{},
			},
			eq: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
					"bar": AnyType{},
				},
				StrictProps: true,
			},
			neq: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
					"bar": AnyType{},
				},
				StrictProps: true,
			},
		},
		{
			what: "strict object equals map object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			eq: &ObjectType{
				Mapped: StringType{},
			},
			neq: &ObjectType{
				Mapped: NullType{},
			},
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: StringType{}},
			neq:  NewObjectType(),
		},
		{
			what: "array element type",
			ty:   &ArrayType{Elem: StringType{}},
			neq:  &ArrayType{Elem: BoolType{}},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{Elem: StringType{}},
			},
			neq: &ArrayType{
				Elem: &ArrayType{Elem: BoolType{}},
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
			if tc.neq != nil {
				l, r = tc.ty, tc.neq
				if l.Equals(r) {
					t.Errorf("%s should not equal to %s", l.String(), r.String())
				}
			}
			if tc.eq != nil {
				l, r = tc.ty, tc.eq
				if !l.Equals(r) {
					t.Errorf("%s should equal to %s", l.String(), r.String())
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

func TestExprTypeStringize(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		want string
	}{
		{
			what: "any",
			ty:   AnyType{},
			want: "any",
		},
		{
			what: "null",
			ty:   NullType{},
			want: "null",
		},
		{
			what: "number",
			ty:   NumberType{},
			want: "number",
		},
		{
			what: "bool",
			ty:   BoolType{},
			want: "bool",
		},
		{
			what: "string",
			ty:   StringType{},
			want: "string",
		},
		{
			what: "empty object",
			ty:   NewObjectType(),
			want: "object",
		},
		{
			what: "empty strict props object",
			ty:   NewStrictObjectType(),
			want: "{}",
		},
		{
			what: "strict object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			want: "{foo: string}",
		},
		{
			what: "non-strict object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			want: "object",
		},
		{
			what: "strict props object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
				StrictProps: true,
			},
			want: "{foo: string}",
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: AnyType{}},
			want: "array<any>",
		},
		{
			what: "nested array",
			ty:   &ArrayType{Elem: &ArrayType{BoolType{}, true}},
			want: "array<array<bool>>",
		},
		{
			what: "object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ArrayType{
						Elem: &ObjectType{
							Props: map[string]ExprType{
								"bar": &ArrayType{
									Elem: StringType{},
								},
							},
							StrictProps: true,
						},
					},
				},
				StrictProps: true,
			},
			want: "{foo: array<{bar: array<string>}>}",
		},
		{
			what: "map object",
			ty:   NewMapObjectType(NumberType{}),
			want: "{string => number}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			have := tc.ty.String()
			if have != tc.want {
				t.Fatalf("wanted %q but got %q", tc.want, have)
			}
		})
	}
}

func TestExprTypeFuseSimple(t *testing.T) {
	testCases := []ExprType{
		AnyType{},
		NullType{},
		NumberType{},
		BoolType{},
		StringType{},
		NewObjectType(),
		NewStrictObjectType(),
		NewMapObjectType(NullType{}),
		&ArrayType{Elem: StringType{}},
	}

	for _, ty := range testCases {
		t.Run("any/"+ty.String(), func(t *testing.T) {
			have := ty.Fuse(AnyType{})
			if _, ok := have.(AnyType); !ok {
				t.Errorf("any type into %s was %s while expecting any", ty.String(), have.String())
			}

			have = (AnyType{}).Fuse(ty)
			if _, ok := have.(AnyType); !ok {
				t.Errorf("%s into any type was %s while expecting any", ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("incompatible/"+ty.String(), func(t *testing.T) {
			var in ExprType
			in = NullType{}
			if ty == (NullType{}) {
				in = StringType{} // null is compatible with null so use string instead
			}

			have := ty.Fuse(in)
			if _, ok := have.(AnyType); !ok {
				t.Errorf("incompatible %s type into %s was %s while expecting any", in.String(), ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("self/"+ty.String(), func(t *testing.T) {
			have := ty.Fuse(ty)
			if !cmp.Equal(ty, have) {
				s := ty.String()
				t.Errorf("%s into %s was %s while expecting %s", s, s, have.String(), s)
			}
		})
	}
}

func TestExprTypeFuseComplicated(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		into ExprType
		want ExprType
	}{
		{
			what: "number fuses into string",
			ty:   NumberType{},
			into: StringType{},
			want: StringType{},
		},
		{
			what: "string is fused by number",
			ty:   StringType{},
			into: NumberType{},
			want: StringType{},
		},
		{
			what: "bool fuses into string",
			ty:   BoolType{},
			into: StringType{},
			want: StringType{},
		},
		{
			what: "string is fused by bool",
			ty:   StringType{},
			into: BoolType{},
			want: StringType{},
		},
		{
			what: "object props",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"bar": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
					"bar": StringType{},
				},
			},
		},
		{
			what: "object into strict object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"bar": StringType{},
				},
				StrictProps: true,
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
					"bar": StringType{},
				},
				StrictProps: false,
			},
		},
		{
			what: "strict object into strict object",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
				},
				StrictProps: true,
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"bar": StringType{},
				},
				StrictProps: true,
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
					"bar": StringType{},
				},
				StrictProps: true,
			},
		},
		{
			what: "compatible prop",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
		},
		{
			what: "any prop into prop",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
			},
		},
		{
			what: "prop into any prop",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
			},
		},
		{
			what: "incompatible prop",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
			},
		},
		{
			what: "compatible array element",
			ty: &ArrayType{
				Elem: NumberType{},
			},
			into: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: StringType{},
			},
		},
		{
			what: "incompatible array element",
			ty: &ArrayType{
				Elem: NullType{},
			},
			into: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "any array element into element",
			ty: &ArrayType{
				Elem: AnyType{},
			},
			into: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "array element into any element",
			ty: &ArrayType{
				Elem: StringType{},
			},
			into: &ArrayType{
				Elem: AnyType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "array into array deref",
			ty:   &ArrayType{StringType{}, false},
			into: &ArrayType{StringType{}, true},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "array deref into array",
			ty:   &ArrayType{StringType{}, true},
			into: &ArrayType{StringType{}, false},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "array deref into array deref",
			ty:   &ArrayType{StringType{}, true},
			into: &ArrayType{StringType{}, true},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "object no prop at left hand side",
			ty:   NewObjectType(),
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
		},
		{
			what: "object no prop at right hand side",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
			into: NewObjectType(),
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": StringType{},
				},
			},
		},
		{
			what: "any elem array at left hand side",
			ty:   &ArrayType{AnyType{}, false},
			into: &ArrayType{StringType{}, false},
			want: &ArrayType{AnyType{}, false},
		},
		{
			what: "any elem array at right hand side",
			ty:   &ArrayType{StringType{}, false},
			into: &ArrayType{AnyType{}, false},
			want: &ArrayType{AnyType{}, false},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{
					Elem: NumberType{},
				},
			},
			into: &ArrayType{
				Elem: &ArrayType{
					Elem: StringType{},
				},
			},
			want: &ArrayType{
				Elem: &ArrayType{
					Elem: StringType{},
				},
			},
		},
		{
			what: "nested objects",
			ty: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"foo":  NumberType{},
							"piyo": NumberType{},
						},
					},
					"aaa": NumberType{},
					"ccc": NumberType{},
				},
			},
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"bar":  StringType{},
							"piyo": StringType{},
						},
					},
					"bbb": StringType{},
					"ccc": StringType{},
				},
			},
			want: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"foo":  NumberType{},
							"bar":  StringType{},
							"piyo": StringType{},
						},
					},
					"aaa": NumberType{},
					"bbb": StringType{},
					"ccc": StringType{},
				},
			},
		},
		{
			what: "map object into compatible map object",
			ty:   NewMapObjectType(NumberType{}),
			into: NewMapObjectType(StringType{}),
			want: NewMapObjectType(StringType{}),
		},
		{
			what: "map object into incompatible map object",
			ty:   NewMapObjectType(NumberType{}),
			into: NewMapObjectType(NullType{}),
			want: NewObjectType(),
		},
		{
			what: "map object into compatible object",
			ty:   NewMapObjectType(NumberType{}),
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": NumberType{},
				},
			},
			want: NewMapObjectType(NumberType{}),
		},
		{
			what: "map object into incompatible object",
			ty:   NewMapObjectType(NumberType{}),
			into: &ObjectType{
				Props: map[string]ExprType{
					"foo": BoolType{},
				},
			},
			want: NewObjectType(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			ty := tc.into.Fuse(tc.ty)
			if !cmp.Equal(ty, tc.want) {
				t.Fatalf(
					"%s into %s was %s while expecting %s\ndiff:\n%s",
					tc.ty.String(),
					tc.into.String(),
					ty.String(),
					tc.want.String(),
					cmp.Diff(tc.want, ty),
				)
			}
		})
	}
}

func TestExprTypeFuseCreateNewInstance(t *testing.T) {
	{
		ty := &ArrayType{
			Elem: NumberType{},
		}
		ty2 := ty.Fuse(&ArrayType{
			Elem: StringType{},
		})
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if _, ok := ty.Elem.(NumberType); !ok {
			t.Fatalf("original element type was modified: %v", ty)
		}
	}

	{
		ty := &ObjectType{
			Props: map[string]ExprType{
				"foo": NumberType{},
			},
		}
		ty2 := ty.Fuse(&ObjectType{
			Props: map[string]ExprType{
				"foo": StringType{},
				"bar": BoolType{},
			},
		})
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if len(ty.Props) != 1 {
			t.Fatalf("new prop was added: %v", ty)
		}
		if _, ok := ty.Props["foo"].(NumberType); !ok {
			t.Fatalf("prop type was modified: %v", ty)
		}
	}
}
