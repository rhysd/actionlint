package actionlint

import (
	"fmt"
	"strings"
)

// Types

// ExprType is interface for types of values in expression.
type ExprType interface {
	// String returns string representation of the type.
	String() string
	// Assignable returns if other type can be assignable to the type.
	Assignable(other ExprType) bool
	// Equals returns if the type is equal to the other type.
	Equals(other ExprType) bool
	// Fuse merges other type into this type. When other type conflicts with this type, fused
	// result is any type as fallback.
	Fuse(other ExprType) ExprType
}

// AnyType represents type which can be any type. It also indicates that a value of the type cannot
// be type-checked since it's type cannot be known statically.
type AnyType struct{}

func (ty AnyType) String() string {
	return "any"
}

// Assignable returns if other type can be assignable to the type.
func (ty AnyType) Assignable(_ ExprType) bool {
	return true
}

// Equals returns if the type is equal to the other type.
func (ty AnyType) Equals(other ExprType) bool {
	return true
}

// Fuse merges other type into this type. When other type conflicts with this type, fused result is
// any type as fallback.
func (ty AnyType) Fuse(other ExprType) ExprType {
	return ty
}

// NullType is type for null value.
type NullType struct{}

func (ty NullType) String() string {
	return "null"
}

// Assignable returns if other type can be assignable to the type.
func (ty NullType) Assignable(other ExprType) bool {
	switch other.(type) {
	case NullType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty NullType) Equals(other ExprType) bool {
	switch other.(type) {
	case NullType, AnyType:
		return true
	default:
		return false
	}
}

// Fuse merges other type into this type. When other type conflicts with this type, fused result is
// any type as fallback.
func (ty NullType) Fuse(other ExprType) ExprType {
	if _, ok := other.(NullType); ok {
		return ty
	}
	return AnyType{}
}

// NumberType is type for number values such as integer or float.
type NumberType struct{}

func (ty NumberType) String() string {
	return "number"
}

// Equals returns if the type is equal to the other type.
func (ty NumberType) Equals(other ExprType) bool {
	switch other.(type) {
	case NumberType, AnyType:
		return true
	default:
		return false
	}
}

// Assignable returns if other type can be assignable to the type.
func (ty NumberType) Assignable(other ExprType) bool {
	// TODO: Is string of numbers corced into number?
	switch other.(type) {
	case NumberType, AnyType:
		return true
	default:
		return false
	}
}

// Fuse merges other type into this type. When other type conflicts with this type, fused result is
// any type as fallback.
func (ty NumberType) Fuse(other ExprType) ExprType {
	if _, ok := other.(NumberType); ok {
		return ty
	}
	return AnyType{}
}

// BoolType is type for boolean values.
type BoolType struct{}

func (ty BoolType) String() string {
	return "bool"
}

// Assignable returns if other type can be assignable to the type.
func (ty BoolType) Assignable(other ExprType) bool {
	// TODO: Is numbers corced into bool?
	switch other.(type) {
	case BoolType, StringType, NullType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty BoolType) Equals(other ExprType) bool {
	switch other.(type) {
	case BoolType, AnyType:
		return true
	default:
		return false
	}
}

// Fuse merges other type into this type. When other type conflicts with this type, fused result is
// any type as fallback.
func (ty BoolType) Fuse(other ExprType) ExprType {
	if _, ok := other.(BoolType); ok {
		return ty
	}
	return AnyType{}
}

// StringType is type for string values.
type StringType struct{}

func (ty StringType) String() string {
	return "string"
}

// Assignable returns if other type can be assignable to the type.
func (ty StringType) Assignable(other ExprType) bool {
	// Bool and null types also can be coerced into string. But in almost all case, those coercing
	// would be mistakes.
	switch other.(type) {
	case StringType, NumberType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty StringType) Equals(other ExprType) bool {
	switch other.(type) {
	case StringType, AnyType:
		return true
	default:
		return false
	}
}

// Fuse merges other type into this type. When other type conflicts with this type, fused result is
// any type as fallback.
func (ty StringType) Fuse(other ExprType) ExprType {
	switch other.(type) {
	case StringType, NumberType:
		return ty // Consider assignability
	default:
		return AnyType{}
	}
}

// ObjectType is type for objects, which can hold key-values.
type ObjectType struct {
	// Props is map from properties name to their type.
	Props map[string]ExprType
	// StrictProps is flag to check if the properties should be checked strictly. When this flag
	// is set to true, it means that other than properties defined in Props field are not permitted
	// and will cause type error. When this flag is set to false, accessing to unknown properties
	// does not cause type error and will be deducted to any type.
	StrictProps bool
}

// NewObjectType creates new ObjectType instance which allows unknown props. When accessing to
// unknown props, their values will fall back to any.
func NewObjectType() *ObjectType {
	return &ObjectType{map[string]ExprType{}, false}
}

// NewStrictObjectType creates new ObjectType instance which does not allow unknown props.
func NewStrictObjectType() *ObjectType {
	return &ObjectType{map[string]ExprType{}, true}
}

func (ty *ObjectType) String() string {
	len := len(ty.Props)
	if len == 0 && !ty.StrictProps {
		return "object"
	}
	ps := make([]string, 0, len)
	for n, t := range ty.Props {
		ps = append(ps, fmt.Sprintf("%s: %s", n, t.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(ps, "; "))
}

// Assignable returns if other type can be assignable to the type.
func (ty *ObjectType) Assignable(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ObjectType:
		for n, p1 := range ty.Props {
			if p2, ok := other.Props[n]; ok && !p1.Assignable(p2) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty *ObjectType) Equals(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ObjectType:
		if !ty.StrictProps || !other.StrictProps {
			return true
		}
		for n, t := range ty.Props {
			o, ok := other.Props[n]
			if !ok || !t.Equals(o) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Fuse merges two object types into one. When other object has unknown props, they are merged into
// current object. When both have same property, when they are assignable, it remains as-is.
// Otherwise, the property falls back to any type.
// Note that this method modifies itself destructively for efficiency.
func (ty *ObjectType) Fuse(other ExprType) ExprType {
	switch other := other.(type) {
	case *ObjectType:
		for n, ot := range other.Props {
			if t, ok := ty.Props[n]; ok {
				ty.Props[n] = t.Fuse(ot)
			} else {
				ty.Props[n] = ot // New prop
			}
		}
		if !other.StrictProps {
			ty.StrictProps = false
		}
		return ty
	default:
		return AnyType{}
	}
}

// ArrayType is type for arrays.
type ArrayType struct {
	// Elem is type of element of the array.
	Elem ExprType
}

func (ty *ArrayType) String() string {
	return fmt.Sprintf("array<%s>", ty.Elem.String())
}

// Equals returns if the type is equal to the other type.
func (ty *ArrayType) Equals(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Equals(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Equals(other.Elem)
	default:
		return false
	}
}

// Assignable returns if other type can be assignable to the type.
func (ty *ArrayType) Assignable(other ExprType) bool {
	// Note: ArrayType and ArrayDerefType are compatible
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Assignable(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Assignable(other.Elem)
	default:
		return false
	}
}

// Fuse merges two object types into one. When other object has unknown props, they are merged into
// current object. When both have same property, when they are assignable, it remains as-is.
// Otherwise, the property falls back to any type.
// Note that this method modifies itself destructively for efficiency.
func (ty *ArrayType) Fuse(other ExprType) ExprType {
	switch other := other.(type) {
	case *ArrayType:
		ty.Elem = ty.Elem.Fuse(other.Elem)
		return ty
	case *ArrayDerefType:
		ty.Elem = ty.Elem.Fuse(other.Elem)
		return ty
	default:
		return AnyType{}
	}
}

// ArrayDerefType is a type for array element dereference with '.*'. It is distinguished from ArrayType
// for type checker.
// For example, when type of 'a' is '{foo: {bar: int}[]}', 'a.*' is type of 'array_deref<{bar: int}>
// ' and 'a.*.b' is type of 'array_deref<int>'.
type ArrayDerefType struct {
	// Elem is type of element of dereferenced (and filtered) array.
	Elem ExprType
}

func (ty *ArrayDerefType) String() string {
	return fmt.Sprintf("array<%s>", ty.Elem.String())
}

// Assignable returns if other type can be assignable to the type.
func (ty *ArrayDerefType) Assignable(other ExprType) bool {
	// Note: ArrayType and ArrayDerefType are compatible
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Assignable(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Assignable(other.Elem)
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty *ArrayDerefType) Equals(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Equals(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Equals(other.Elem)
	default:
		return false
	}
}

// Fuse merges two object types into one. When other object has unknown props, they are merged into
// current object. When both have same property, when they are assignable, it remains as-is.
// Otherwise, the property falls back to any type.
// Note that this method modifies itself destructively for efficiency.
func (ty *ArrayDerefType) Fuse(other ExprType) ExprType {
	switch other := other.(type) {
	case *ArrayType:
		ty.Elem = ty.Elem.Fuse(other.Elem)
		return ty
	case *ArrayDerefType:
		ty.Elem = ty.Elem.Fuse(other.Elem)
		return ty
	default:
		return AnyType{}
	}
}

// ElemTypeOf returns element type of given type when it is array type.
// When it is any type, it returns any type. Otherwise ti returns nil.
func ElemTypeOf(ty ExprType) (ExprType, bool) {
	switch ty := ty.(type) {
	case AnyType:
		return AnyType{}, true
	case *ArrayType:
		return ty.Elem, true
	case *ArrayDerefType:
		return ty.Elem, true
	default:
		return nil, false
	}
}
