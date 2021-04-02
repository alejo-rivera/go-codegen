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

// ExtractTemplatesFromStruct returns a slice of template types that should be
// invoked upon the provided struct.
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

// ExtractArgs parses the arguments out of a template invocation, using
// the invoking fields tags.
// func ExtractArgs(ctx *Context, stp *ast.StructType, name string) ([]string, error) {
// 	var found *ast.Field

// 	for _, f := range stp.Fields.List {
// 		// fname, err := nameFromFieldType(ctx, f.Type)
// 		fname, err := "test", error(nil)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if name == fname {
// 			found = f
// 		}
// 	}

// 	if found == nil {
// 		return nil, errors.New("Couldn't find template invocation: " + name)
// 	}

// 	if found.Tag == nil {
// 		return nil, nil
// 	}

// 	tag := reflect.StructTag(found.Tag.Value[1 : len(found.Tag.Value)-1])

// 	return strings.Split(tag.Get("template"), ","), nil
// }

// func extractText(ctx *Context, t ast.Expr) (string, error) {
// 	pos := ctx.Fset.Position(t.Pos())
// 	end := ctx.Fset.Position(t.End())

// 	read, err := ioutil.ReadFile(pos.Filename)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(read[pos.Offset:end.Offset]), nil
// }

// func nameFromFieldType(ctx *Context, t ast.Expr) (string, error) {
// 	txt, err := extractText(ctx, t)
// 	if err != nil {
// 		return "", err
// 	}

// 	return txt, nil
// }
