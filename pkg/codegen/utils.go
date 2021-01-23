package codegen

import (
	"fmt"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func WrapTypeInPointer(t Type) Type {
	if isPointerType(t) {
		return t
	}
	return &PointerType{Type: t}
}

func isPointerType(t Type) bool {
	switch x := t.(type) {
	case *PointerType:
		return true
	case *NamedType:
		return isPointerType(x.Decl.Type)
	default:
		return false
	}
}

func PrimitiveTypeFromJSONSchemaType(jsType string) (Type, error) {
	switch jsType {
	case schemas.TypeNameString:
		return PrimitiveType{"string"}, nil
	case schemas.TypeNameNumber:
		return PrimitiveType{"float64"}, nil
	case schemas.TypeNameInteger:
		return PrimitiveType{"int"}, nil
	case schemas.TypeNameBoolean:
		return PrimitiveType{"bool"}, nil
	case schemas.TypeNameNull:
		return NullType{}, nil
	case schemas.TypeNameObject, schemas.TypeNameArray:
		return nil, fmt.Errorf("unexpected type %q here", jsType)
	}
	return nil, fmt.Errorf("unknown JSON Schema type %q", jsType)
}

func MergeType(a, b Type) (Type, error) {
	if a == nil {
		return b, nil
	}

	if b == nil {
		return a, nil
	}

	aStruct, isAStruct := a.(*StructType)
	bStruct, isBStruct := b.(*StructType)

	if isAStruct && isBStruct {
		return mergeStructTypes(aStruct, bStruct), nil
	}

	aNamedType, isANamedType := a.(*NamedType)
	bNamedType, isBNamedType := b.(*NamedType)

	if isANamedType && isBNamedType {
		return nil, fmt.Errorf("a and b cannot be NamedTypes, a is '%s' and b is '%s'", aNamedType.GetName(), bNamedType.GetName())
	}

	if isAStruct && isBNamedType {
		subStructType, isSubTypeStruct := bNamedType.Decl.Type.(*StructType)
		if !isSubTypeStruct {
			return nil, fmt.Errorf("b named type '%s' has an invalid type decl type '%T'", bNamedType.GetName(), bNamedType.Decl.Type)
		}

		return mergeStructTypes(subStructType, aStruct), nil
	}

	if isBStruct && isANamedType {
		subStructType, isSubTypeStruct := aNamedType.Decl.Type.(*StructType)
		if !isSubTypeStruct {
			return nil, fmt.Errorf("a named type '%s' has an invalid type decl type '%T'", aNamedType.GetName(), aNamedType.Decl.Type)
		}

		aNamedType.Decl.Type = mergeStructTypes(subStructType, aStruct)
		return aNamedType, nil
	}

	return nil, fmt.Errorf("unable to merge types %T and %T", a, b)
}

func mergeStructTypes(a, b *StructType) *StructType {
	result := &StructType{
		Fields:             append([]StructField{}, a.Fields...),
		RequiredJSONFields: append([]string{}, a.RequiredJSONFields...),
	}

	for _, field := range b.Fields {
		result.AddField(field)
	}

	if len(b.RequiredJSONFields) > 0 {
		result.RequiredJSONFields = append(result.RequiredJSONFields, a.RequiredJSONFields...)
	}

	return result
}
