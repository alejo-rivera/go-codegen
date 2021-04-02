package codegen

import (
	"go/token"
	"go/types"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

// Context represents the context in which a code generation operation is run.
type GenContext struct {
	Fset        *token.FileSet
	PackageName string
	Generated   []string

	templates   map[string]*template.Template
	imports     []string
	importsSeen map[string]struct{}
}

func NewGenContext() *GenContext {
	return &GenContext{
		Fset:        token.NewFileSet(),
		templates:   make(map[string]*template.Template),
		importsSeen: make(map[string]struct{}),
	}
}

func (ctx *GenContext) TemplateForGenType(genType *types.Named) (*template.Template, error) {
	genName := genType.Obj().Name()
	fullName := genType.Obj().Pkg().Path() + "." + genName
	if template, ok := ctx.templates[fullName]; ok {
		return template, nil
	}

	pos := genType.Obj().Pos()
	fpath := ctx.Fset.Position(pos).Filename
	templatePath := filepath.Join(filepath.Dir(fpath), genName+".tmpl")
	template, err := ParseTemplate(templatePath)
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}
	ctx.templates[fullName] = template
	return template, nil
}

func (ctx *GenContext) AddImport(pkg string) {
	if _, seen := ctx.importsSeen[pkg]; seen {
		return
	}
	ctx.imports = append(ctx.imports, pkg)
	ctx.importsSeen[pkg] = struct{}{}
}

func (ctx *GenContext) Imports() []string {
	i := make([]string, len(ctx.imports))
	copy(i, ctx.imports)
	return i
}
