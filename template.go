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

func RunTemplate(invocation Invocation, aStruct *types.Named, ctx *GenContext) error {
	template, err := ctx.TemplateForGenType(invocation.GenType)
	if err != nil {
		return errors.Wrap(err, "getting template")
	}

	c := &TemplateContext{
		Args:         invocation.Args,
		StructName:   aStruct.Obj().Name(),
		TemplateName: template.Name(),
		PackageName:  ctx.PackageName,
		Struct:       aStruct.Underlying().(*types.Struct),
		ctx:          ctx,
	}
	var result bytes.Buffer
	if err := template.Execute(&result, c); err != nil {
		return err
	}
	ctx.Generated = append(ctx.Generated, result.String())
	return nil
}

type TemplateContext struct {
	Args         []string
	StructName   string
	TemplateName string
	PackageName  string
	Struct       *types.Struct

	ctx *GenContext
}

// For a function to be callable from a template, it must return something.
func (c *TemplateContext) AddImport(name string) string {
	c.ctx.AddImport(name)
	return ""
}

func (c *TemplateContext) Implements(aType types.Type, interfaceName string) (bool, error) {
	t, err := c.ctx.GetType(interfaceName)
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
}

func typeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Pkg().Name() + "." + t.Obj().Name()
	default:
		return t.String()
	}
}
