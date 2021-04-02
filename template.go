package codegen

import (
	"bytes"
	"go/types"

	"github.com/pkg/errors"
)

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
