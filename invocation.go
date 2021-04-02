package codegen

import (
	"errors"
	"go/types"
	"reflect"
	"strings"
)

type Invocation struct {
	GenType *types.Named
	Args    []string
}

func InvocationsForStruct(aStruct *types.Struct) ([]Invocation, error) {
	var invocations []Invocation
	for i := 0; i < aStruct.NumFields(); i++ {
		tag := reflect.StructTag(aStruct.Tag(i))
		genTag, ok := tag.Lookup("codegen")
		if !ok {
			continue
		}
		field := aStruct.Field(i)
		if !field.Embedded() {
			return nil, errors.New("codegen tag used on non embedded field " + field.Name())
		}
		genType, ok := aStruct.Field(i).Type().(*types.Named)
		if !ok {
			return nil, errors.New("expected named type for field " + field.Name())
		}

		invocations = append(invocations, Invocation{
			GenType: genType,
			Args:    strings.Split(genTag, ","),
		})
	}

	return invocations, nil
}
