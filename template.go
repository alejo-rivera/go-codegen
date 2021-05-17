package codegen

import (
	"bytes"
	"go/types"
	"path/filepath"
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

func ParseTemplate(path string) (*template.Template, error) {
	name := filepath.Base(path)
	return template.New(name).Funcs(templateFunctions).ParseFiles(path)
}

var templateFunctions template.FuncMap

func init() {
	templateFunctions = sprig.TxtFuncMap()
	templateFunctions["singular"] = inflection.Singular
	templateFunctions["plural"] = inflection.Plural
	templateFunctions["typeName"] = typeName
	templateFunctions["pointerType"] = pointerType
	templateFunctions["structFields"] = structFields
	templateFunctions["structField"] = structField
}

func typeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Pkg().Name() + "." + t.Obj().Name()
	case interface{ Elem() types.Type }:
		return typeName(t.Elem())
	default:
		return t.String()
	}
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
