package codegen

import (
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jinzhu/inflection"
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

// NewContext initializes a new code generation context.
func NewGenContext() *GenContext {
	return &GenContext{
		Fset:        token.NewFileSet(),
		templates:   make(map[string]*template.Template),
		importsSeen: make(map[string]struct{}),
	}
}

func (ctx *GenContext) TemplateForGenType(genType *types.Named) (*template.Template, error) {
	name := genType.Obj().Name()
	fullName := genType.Obj().Pkg().Path() + "." + name
	fmt.Println("looking for template for " + fullName)
	if template, ok := ctx.templates[fullName]; ok {
		return template, nil
	}

	pos := genType.Obj().Pos()
	fpath := ctx.Fset.Position(pos).Filename
	templateName := name + ".tmpl"
	templatePath := filepath.Join(filepath.Dir(fpath), templateName)
	fmt.Println("looking for template at " + templatePath)

	template, err := template.
		New(templateName).
		Funcs(templateFunctions).
		ParseFiles(templatePath)
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}
	fmt.Println(template)
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
