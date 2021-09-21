package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExprNewMapObjectType(t *testing.T) {
	o := NewMapObjectType(theStringType)
	if o.Props != nil {
		t.Fatalf("props should be nil but %v", o.Props)
	}
	if _, ok := o.Mapped.(*StringType); !ok {
		t.Fatalf("mapped type is not string: %v", o.Mapped)
	}
	if o.IsStrict() {
		t.Fatalf("map object is not strict object but got %v", o)
	}
}

func TestExprObjectTypeSetStrict(t *testing.T) {
	o := NewEmptyObjectType()
	if o.IsStrict() || !o.IsLoose() {
		t.Fatal("should be loose")
	}
	o.Strict()
	if !o.IsStrict() || o.IsLoose() {
		t.Fatal("should be strict")
	}
	o.Loose()
	if o.IsStrict() || !o.IsLoose() {
		t.Fatal("should be loose")
	}
}

func TestExprAssignableSimple(t *testing.T) {
	testCases := []ExprType{
		theAnyType,
		theNullType,
		theNumberType,
		theBoolType,
		theStringType,
		NewObjectType(map[string]ExprType{"n": theNumberType}),
		NewStrictObjectType(map[string]ExprType{"b": theBoolType}),
		NewMapObjectType(theNullType),
		&ArrayType{Elem: theStringType},
	}

	for _, ty := range testCases {
		s := ty.String()
		t.Run(s, func(t *testing.T) {
			if !ty.Assignable(ty) {
				t.Fatalf("%s is not self-assignable", ty)
			}

			switch ty.(type) {
			case *NullType:
			case *AnyType:
			default:
				if (theNullType).Assignable(ty) {
					t.Fatalf("%s is assignable to null", ty)
				}
			}

			if !(theAnyType).Assignable(ty) {
				t.Fatalf("%s is not assignable to any", ty)
			}
		})
	}
}

func TestExprAssignableObject(t *testing.T) {
	testCases := []struct {
		from, to ExprType
		no       bool
	}{
		{
			from: NewMapObjectType(theNumberType),
			to:   NewMapObjectType(theStringType),
		},
		{
			from: NewEmptyObjectType(),
			to:   NewMapObjectType(theStringType),
		},
		{
			from: NewMapObjectType(theStringType),
			to:   NewEmptyObjectType(),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{
				"a": theNumberType,
				"b": theStringType,
			}),
			to: NewMapObjectType(theStringType),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": theNullType}),
			to:   NewMapObjectType(theStringType),
			no:   true,
		},
		{
			from: NewMapObjectType(theNumberType),
			to: NewStrictObjectType(map[string]ExprType{
				"a": theAnyType,
				"b": theStringType,
			}),
		},
		{
			from: NewMapObjectType(theNumberType),
			to: NewStrictObjectType(map[string]ExprType{
				"a": theNullType,
				"b": theStringType,
			}),
			no: true,
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": theNumberType}),
			to:   NewStrictObjectType(map[string]ExprType{"a": theStringType}),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": theStringType}),
			to:   NewStrictObjectType(map[string]ExprType{"b": theStringType}),
			no:   true,
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": theNullType}),
			to:   NewStrictObjectType(map[string]ExprType{"a": theStringType}),
			no:   true,
		},
	}

	for _, tc := range testCases {
		l, r := tc.to.String(), tc.from.String()
		t.Run(l+" := "+r, func(t *testing.T) {
			if tc.to.Assignable(tc.from) == tc.no {
				not := ""
				if tc.no {
					not = " not"
				}
				t.Fatalf("%s should%s be assignable to %s", r, not, l)
			}
		})
	}
}

func TestExprEqualTypes(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		neq  ExprType
		eq   ExprType
	}{
		{
			what: "null",
			ty:   theNullType,
			neq:  theStringType,
		},
		{
			what: "number",
			ty:   theNumberType,
			neq:  theStringType,
		},
		{
			what: "bool",
			ty:   theBoolType,
			neq:  theStringType,
		},
		{
			what: "string",
			ty:   theStringType,
			neq:  theBoolType,
		},
		{
			what: "object",
			ty:   NewEmptyObjectType(),
			neq:  &ArrayType{Elem: theAnyType},
		},
		{
			what: "strict props object",
			ty:   NewEmptyStrictObjectType(),
			neq:  &ArrayType{Elem: theAnyType},
		},
		{
			what: "nested object",
			ty: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"bar": theStringType,
				}),
			}),
			neq: &ArrayType{Elem: theAnyType},
		},
		{
			what: "nested strict props object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"bar": theStringType,
				}),
			}),
			neq: &ArrayType{Elem: theAnyType},
		},
		{
			what: "nested object prop name",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"bar": theStringType,
			}),
		},
		{
			what: "nested object prop type",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": theBoolType,
			}),
		},
		{
			what: "strict props object and loose object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
			eq: NewObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
		},
		{
			what: "loose object and strict props object",
			ty: NewObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
			eq: NewObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
		},
		{
			what: "map objects",
			ty:   NewMapObjectType(theNullType),
			eq:   NewMapObjectType(theNullType),
			neq:  NewMapObjectType(theNumberType),
		},
		{
			what: "map object equals loose object",
			ty:   NewMapObjectType(theStringType),
			eq:   NewEmptyObjectType(),
		},
		{
			what: "loose object equals map object",
			ty:   NewEmptyObjectType(),
			eq:   NewMapObjectType(theStringType),
		},
		{
			what: "map object equals strict object",
			ty:   NewMapObjectType(theStringType),
			eq: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
		},
		{
			what: "map object equals strict object including any prop",
			ty:   NewMapObjectType(theStringType),
			eq: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
				"bar": theAnyType,
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": theNullType,
				"bar": theAnyType,
			}),
		},
		{
			what: "strict object equals map object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			eq:  NewMapObjectType(theStringType),
			neq: NewMapObjectType(theNullType),
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: theStringType},
			neq:  NewEmptyObjectType(),
		},
		{
			what: "array element type",
			ty:   &ArrayType{Elem: theStringType},
			neq:  &ArrayType{Elem: theBoolType},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{Elem: theStringType},
			},
			neq: &ArrayType{
				Elem: &ArrayType{Elem: theBoolType},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			var l, r ExprType

			l, r = tc.ty, tc.ty
			if !EqualTypes(l, r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			if tc.neq != nil {
				l, r = tc.ty, tc.neq
				if EqualTypes(l, r) {
					t.Errorf("%s should not equal to %s", l.String(), r.String())
				}
			}
			if tc.eq != nil {
				l, r = tc.ty, tc.eq
				if !EqualTypes(l, r) {
					t.Errorf("%s should equal to %s", l.String(), r.String())
				}
			}
			l, r = tc.ty, theAnyType
			if !EqualTypes(l, r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			l, r = theAnyType, tc.ty
			if !EqualTypes(l, r) {
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
			ty:   theAnyType,
			want: "any",
		},
		{
			what: "null",
			ty:   theNullType,
			want: "null",
		},
		{
			what: "number",
			ty:   theNumberType,
			want: "number",
		},
		{
			what: "bool",
			ty:   theBoolType,
			want: "bool",
		},
		{
			what: "string",
			ty:   theStringType,
			want: "string",
		},
		{
			what: "empty object",
			ty:   NewEmptyObjectType(),
			want: "object",
		},
		{
			what: "empty strict props object",
			ty:   NewEmptyStrictObjectType(),
			want: "{}",
		},
		{
			what: "strict object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: "{foo: string}",
		},
		{
			what: "non-strict object",
			ty: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: "object",
		},
		{
			what: "strict props object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: "{foo: string}",
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: theAnyType},
			want: "array<any>",
		},
		{
			what: "nested array",
			ty:   &ArrayType{Elem: &ArrayType{theBoolType, true}},
			want: "array<array<bool>>",
		},
		{
			what: "object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": &ArrayType{
					Elem: NewStrictObjectType(map[string]ExprType{
						"bar": &ArrayType{
							Elem: theStringType,
						},
					}),
				},
			}),
			want: "{foo: array<{bar: array<string>}>}",
		},
		{
			what: "map object",
			ty:   NewMapObjectType(theNumberType),
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

func TestExprTypeMergeSimple(t *testing.T) {
	testCases := []ExprType{
		theAnyType,
		theNullType,
		theNumberType,
		theBoolType,
		theStringType,
		NewEmptyObjectType(),
		NewEmptyStrictObjectType(),
		NewMapObjectType(theNullType),
		&ArrayType{Elem: theStringType},
	}

	opt := cmpopts.EquateEmpty()

	for _, ty := range testCases {
		t.Run("any/"+ty.String(), func(t *testing.T) {
			have := ty.Merge(theAnyType)
			if _, ok := have.(*AnyType); !ok {
				t.Errorf("any type merged with %s was %s while expecting any", ty.String(), have.String())
			}

			have = (theAnyType).Merge(ty)
			if _, ok := have.(*AnyType); !ok {
				t.Errorf("%s merged with any type was %s while expecting any", ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("incompatible/"+ty.String(), func(t *testing.T) {
			var in ExprType
			in = theNullType
			if ty == (theNullType) {
				in = theStringType // null is compatible with null so use string instead
			}

			have := ty.Merge(in)
			if _, ok := have.(*AnyType); !ok {
				t.Errorf("incompatible %s type merged with %s was %s while expecting any", in.String(), ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("self/"+ty.String(), func(t *testing.T) {
			have := ty.Merge(ty)
			if !cmp.Equal(ty, have, opt) {
				s := ty.String()
				t.Errorf("%s merged with %s was %s while expecting %s", s, s, have.String(), s)
			}
		})
	}
}

func TestExprTypeMergeComplicated(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		with ExprType
		want ExprType
	}{
		{
			what: "number merges with string",
			ty:   theNumberType,
			with: theStringType,
			want: theStringType,
		},
		{
			what: "string is merged by number",
			ty:   theStringType,
			with: theNumberType,
			want: theStringType,
		},
		{
			what: "bool merges with string",
			ty:   theBoolType,
			with: theStringType,
			want: theStringType,
		},
		{
			what: "string is merged by bool",
			ty:   theStringType,
			with: theBoolType,
			want: theStringType,
		},
		{
			what: "object props",
			ty: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
			with: NewObjectType(map[string]ExprType{
				"bar": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
				"bar": theStringType,
			}),
		},
		{
			what: "loose object with strict object",
			ty: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
			with: NewStrictObjectType(map[string]ExprType{
				"bar": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
				"bar": theStringType,
			}),
		},
		{
			what: "strict object with strict object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
			with: NewStrictObjectType(map[string]ExprType{
				"bar": theStringType,
			}),
			want: NewStrictObjectType(map[string]ExprType{
				"foo": theNumberType,
				"bar": theStringType,
			}),
		},
		{
			what: "compatible prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
		},
		{
			what: "any prop with prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": theAnyType,
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theAnyType,
			}),
		},
		{
			what: "prop with any prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": theAnyType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theAnyType,
			}),
		},
		{
			what: "incompatible prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": theNullType,
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theAnyType,
			}),
		},
		{
			what: "compatible array element",
			ty: &ArrayType{
				Elem: theNumberType,
			},
			with: &ArrayType{
				Elem: theStringType,
			},
			want: &ArrayType{
				Elem: theStringType,
			},
		},
		{
			what: "incompatible array element",
			ty: &ArrayType{
				Elem: theNullType,
			},
			with: &ArrayType{
				Elem: theStringType,
			},
			want: &ArrayType{
				Elem: theAnyType,
			},
		},
		{
			what: "any array element with element",
			ty: &ArrayType{
				Elem: theAnyType,
			},
			with: &ArrayType{
				Elem: theStringType,
			},
			want: &ArrayType{
				Elem: theAnyType,
			},
		},
		{
			what: "array element with any element",
			ty: &ArrayType{
				Elem: theStringType,
			},
			with: &ArrayType{
				Elem: theAnyType,
			},
			want: &ArrayType{
				Elem: theAnyType,
			},
		},
		{
			what: "array with array deref",
			ty:   &ArrayType{theStringType, false},
			with: &ArrayType{theStringType, true},
			want: &ArrayType{theStringType, false},
		},
		{
			what: "array deref with array",
			ty:   &ArrayType{theStringType, true},
			with: &ArrayType{theStringType, false},
			want: &ArrayType{theStringType, false},
		},
		{
			what: "array deref with array deref",
			ty:   &ArrayType{theStringType, true},
			with: &ArrayType{theStringType, true},
			want: &ArrayType{theStringType, false},
		},
		{
			what: "object no prop at left hand side",
			ty:   NewEmptyObjectType(),
			with: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
		},
		{
			what: "object no prop at right hand side",
			ty: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
			with: NewEmptyObjectType(),
			want: NewObjectType(map[string]ExprType{
				"foo": theStringType,
			}),
		},
		{
			what: "any elem array at left hand side",
			ty:   &ArrayType{theAnyType, false},
			with: &ArrayType{theStringType, false},
			want: &ArrayType{theAnyType, false},
		},
		{
			what: "any elem array at right hand side",
			ty:   &ArrayType{theStringType, false},
			with: &ArrayType{theAnyType, false},
			want: &ArrayType{theAnyType, false},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{
					Elem: theNumberType,
				},
			},
			with: &ArrayType{
				Elem: &ArrayType{
					Elem: theStringType,
				},
			},
			want: &ArrayType{
				Elem: &ArrayType{
					Elem: theStringType,
				},
			},
		},
		{
			what: "nested objects",
			ty: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"foo":  theNumberType,
					"piyo": theNumberType,
				}),
				"aaa": theNumberType,
				"ccc": theNumberType,
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"bar":  theStringType,
					"piyo": theStringType,
				}),
				"bbb": theStringType,
				"ccc": theStringType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"foo":  theNumberType,
					"bar":  theStringType,
					"piyo": theStringType,
				}),
				"aaa": theNumberType,
				"bbb": theStringType,
				"ccc": theStringType,
			}),
		},
		{
			what: "map object with compatible map object",
			ty:   NewMapObjectType(theNumberType),
			with: NewMapObjectType(theStringType),
			want: NewMapObjectType(theStringType),
		},
		{
			what: "map object with incompatible map object",
			ty:   NewMapObjectType(theNumberType),
			with: NewMapObjectType(theNullType),
			want: NewEmptyObjectType(),
		},
		{
			what: "map object with compatible object",
			ty:   NewMapObjectType(theNumberType),
			with: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theNumberType,
			}),
		},
		{
			what: "map object with incompatible object",
			ty:   NewMapObjectType(theNumberType),
			with: NewObjectType(map[string]ExprType{
				"foo": theBoolType,
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": theBoolType,
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			opt := cmpopts.EquateEmpty()
			ty := tc.with.Merge(tc.ty)
			if !cmp.Equal(ty, tc.want, opt) {
				t.Fatalf(
					"%s was merged with %s as %s while expecting %s\ndiff:\n%s",
					tc.ty.String(),
					tc.with.String(),
					ty.String(),
					tc.want.String(),
					cmp.Diff(tc.want, ty, opt),
				)
			}
		})
	}
}

func TestExprTypeMergeCreateNewInstance(t *testing.T) {
	{
		ty := &ArrayType{
			Elem: theNumberType,
		}
		ty2 := ty.Merge(&ArrayType{
			Elem: theStringType,
		})
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if _, ok := ty.Elem.(*NumberType); !ok {
			t.Fatalf("original element type was modified: %v", ty)
		}
	}

	{
		ty := NewObjectType(map[string]ExprType{
			"foo": theNumberType,
		})
		ty2 := ty.Merge(
			NewObjectType(map[string]ExprType{
				"foo": theStringType,
				"bar": theBoolType,
			}),
		)
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if len(ty.Props) != 1 {
			t.Fatalf("new prop was added: %v", ty)
		}
		if _, ok := ty.Props["foo"].(*NumberType); !ok {
			t.Fatalf("prop type was modified: %v", ty)
		}
	}
}
