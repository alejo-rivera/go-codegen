package codegen

import (
	"bytes"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
)

func RunTemplate(
	template *template.Template,
	aStruct *types.Named,
	args map[string]string,
	info TypeInfo,
) (string, error) {
	c := &TemplateContext{
		Args:         args,
		StructName:   aStruct.Obj().Name(),
		TemplateName: template.Name(),
		PackageName:  aStruct.Obj().Pkg().Name(),
		PackagePath:  aStruct.Obj().Pkg().Path(),
		Struct:       aStruct,
		info:         info,
	}
	var result bytes.Buffer
	if err := template.Execute(&result, c); err != nil {
		return "", err
	}
	return result.String(), nil
}

type TypeInfo interface {
	AddImport(pkg string)
	GetType(fullName string) (types.Type, error)
}

type TemplateContext struct {
	Args         map[string]string
	StructName   string
	TemplateName string
	PackageName  string
	PackagePath  string
	Struct       *types.Named

	info TypeInfo
}

// For a function to be callable from a template, it must return something.
func (c *TemplateContext) AddImport(name string) string {
	c.info.AddImport(name)
	return ""
}

func (c *TemplateContext) AddImportType(t types.Type) (string, error) {
	switch t := t.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg != nil {
			c.AddImport(pkg.Path())
		}
	case *types.Map:
		if _, err := c.AddImportType(t.Key()); err != nil {
			return "", errors.Wrapf(err, "importing map key type '%s' of type %T", t, t)
		}
		if _, err := c.AddImportType(t.Elem()); err != nil {
			return "", errors.Wrapf(err, "importing element type '%s' of type %T", t, t)
		}
	case *types.Interface:
		// A bare `interface{}` we don't need to import, but for an interface
		// literal, we need to import its embedded interfaces and function
		// parameters and return types.
		for i := 0; i < t.NumEmbeddeds(); i++ {
			e := t.EmbeddedType(i)
			if _, err := c.AddImportType(e); err != nil {
				return "", errors.Wrapf(
					err,
					"importing embedded type '%s' of type %T from interface '%s'",
					e, e, t,
				)
			}
		}
		for i := 0; i < t.NumExplicitMethods(); i++ {
			m := t.ExplicitMethod(i)
			mt := m.Type().(*types.Signature)
			if _, err := c.AddImportType(mt); err != nil {
				return "", errors.Wrapf(
					err,
					"importing method of type %T from interface method '%s'",
					mt, m.Name(),
				)
			}
		}
	case *types.Struct:
		// A named struct type will be handled above by `types.Named`, for a struct
		// type literal, we need to import its field types.
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			if _, err := c.AddImportType(f.Type()); err != nil {
				return "", errors.Wrapf(
					err,
					"importing struct field %s, '%s' of type %T",
					f.Name(), f.Type(), f.Type(),
				)
			}
		}
	case *types.Tuple:
		// Used for function parameters and results
		for i := 0; i < t.Len(); i++ {
			if _, err := c.AddImportType(t.At(i).Type()); err != nil {
				return "", errors.Wrapf(
					err,
					"importing tuple index %d, '%s' of type %T",
					i, t, t,
				)
			}
		}
	case *types.Signature:
		p := t.Params()
		if _, err := c.AddImportType(p); err != nil {
			return "", errors.Wrapf(
				err,
				"importing params '%s' of type %T from function '%s'",
				p, p, t.String(),
			)
		}
		r := t.Results()
		if _, err := c.AddImportType(r); err != nil {
			return "", errors.Wrapf(
				err,
				"importing results '%s' of type %T from function '%s'",
				r, r, t.String(),
			)
		}

	case interface{ Elem() types.Type }: // Array, Slice, Pointer, Channel
		return c.AddImportType(t.Elem())
	case *types.Basic:
		// No need to import
	default:
		return "", errors.Errorf("couldn't add import for '%s' of type %T", t, t)
	}
	return "", nil
}

func (c *TemplateContext) Arg(name string) string {
	return c.Args[name]
}

func (c *TemplateContext) HasArg(name string) bool {
	_, has := c.Args[name]
	return has
}

func (c *TemplateContext) RequireArg(name string) (string, error) {
	arg, has := c.Args[name]
	if !has {
		return "", errors.Errorf("required arg %s not found", name)
	}
	return arg, nil
}

func (c *TemplateContext) DefaultArg(name, defaultVal string) string {
	arg, has := c.Args[name]
	if !has {
		return defaultVal
	}
	return arg
}

func (c *TemplateContext) Implements(aType types.Type, interfaceName string) (bool, error) {
	t, err := c.info.GetType(interfaceName)
	if err != nil {
		return false, err
	}
	i, ok := t.(*types.Interface)
	if !ok {
		return false, errors.Errorf("%s is not an interface", interfaceName)
	}
	return types.Implements(aType, i), nil
}

func (c *TemplateContext) TypeString(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p == nil {
			return ""
		} else if p.Path() == c.PackagePath {
			return ""
		} else {
			return p.Name()
		}
	})
}

func ParseTemplate(path string) (*template.Template, error) {
	name := filepath.Base(path)
	return template.New(name).Funcs(templateFunctions).ParseFiles(path)
}

var templateFunctions template.FuncMap

func init() {
	templateFunctions = sprig.TxtFuncMap()
	templateFunctions["singular"] = inflection.Singular
	templateFunctions["plural"] = inflection.Plural
	templateFunctions["catNoSpace"] = catNoSpace
	templateFunctions["typeName"] = typeName
	templateFunctions["isExported"] = isExported
	templateFunctions["pointerType"] = pointerType
	templateFunctions["structFields"] = structFields
	templateFunctions["structField"] = structField
}

func catNoSpace(ss ...string) string {
	return strings.Join(ss, "")
}

func typeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg != nil {
			return t.Obj().Pkg().Name() + "." + t.Obj().Name()
		} else {
			return t.Obj().Name()
		}
	case interface{ Elem() types.Type }:
		return typeName(t.Elem())
	default:
		return t.String()
	}
}

func isExported(name string) bool {
	return token.IsExported(name)
}

func structFields(t types.Type) []*types.Var {
	s := structFromType(t)
	if s == nil {
		return nil
	}
	var fields []*types.Var
	for i := 0; i < s.NumFields(); i++ {
		fields = append(fields, s.Field(i))
	}
	return fields
}

func structField(t types.Type, fieldName string) *types.Var {
	s := structFromType(t)
	if s == nil {
		return nil
	}

	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if f.Name() == fieldName {
			return f
		}
	}

	// If we can't find it on the struct directly try checking nested structs with
	// a lookup (will only be able to find exported fields in nested structs).
	obj, _, _ := types.LookupFieldOrMethod(s, true, nil, fieldName)
	switch obj := obj.(type) {
	case *types.Var:
		return obj
	default:
		return nil
	}
}

func pointerType(t types.Type) *types.Pointer {
	return types.NewPointer(t)
}

func structFromType(t types.Type) *types.Struct {
	switch t := t.(type) {
	case *types.Struct:
		return t
	case *types.Named:
		return structFromType(t.Underlying())
	case interface{ Elem() types.Type }:
		return structFromType(t.Elem())
	default:
		return nil
	}
}
