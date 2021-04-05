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
		Struct:       aStruct.Underlying().(*types.Struct),
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
	Struct       *types.Struct

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
}

func typeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Pkg().Name() + "." + t.Obj().Name()
	default:
		return t.String()
	}
}

func pointerType(t types.Type) *types.Pointer {
	return types.NewPointer(t)
}
