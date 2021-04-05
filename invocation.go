package codegen

import (
	"go/types"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type Invocation struct {
	GenType *types.Named
	Args    map[string]string
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
		genType, ok := aStruct.Field(i).Type().(*types.Named)
		if !ok {
			return nil, errors.New("expected named type for field " + field.Name())
		}

		args, err := parseArgs(genTag)
		if err != nil {
			return nil, errors.Wrap(err, "parsing codgen tag for args")
		}

		invocations = append(invocations, Invocation{
			GenType: genType,
			Args:    args,
		})

		// If the codegen field is itself a struct, then recurse.
		if structType, ok := genType.Underlying().(*types.Struct); ok {
			nested, err := InvocationsForStruct(structType)
			if err != nil {
				return nil, errors.Wrap(err, "nested "+genType.Obj().Name())
			}
			// Pass any args defined by the outer invocation that aren't defined by
			// the inner invocation down.
			for arg, v := range args {
				for _, n := range nested {
					if _, inner := n.Args[arg]; !inner {
						n.Args[arg] = v
					}
				}
			}
			invocations = append(invocations, nested...)
		}
	}

	return invocations, nil
}

func parseArgs(tag string) (map[string]string, error) {
	// The format of the tag of `foo=bar,baz=quux` is almost identical to a query
	// string so we picky back on this implementation, first replacing the commas
	// with semicolons to be compatible with the query string
	values, err := url.ParseQuery(strings.ReplaceAll(tag, ",", ";"))
	if err != nil {
		return nil, err
	}

	// If an arg is specified more than once, the first occurrence wins.
	args := make(map[string]string, len(values))
	for arg, vals := range values {
		val := ""
		if len(vals) > 0 {
			val = vals[0]
		}
		args[arg] = val
	}

	return args, nil
}
