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

	// TODO: unexported
	Templates map[string]*template.Template

	// TODO: ordered imports
	Imports   map[string]bool
	Generated []string
}

// NewContext initializes a new code generation context.
func NewGenContext() *GenContext {
	return &GenContext{
		Fset:      token.NewFileSet(),
		Templates: map[string]*template.Template{},
		Imports:   map[string]bool{},
	}
}

// Populate fills in the rest of the context based upon the context's
// config.
// func (ctx *Context) Populate() error {
// 	for _, dir := range ctx.SearchPaths {
// 		err := ctx.searchDir(dir)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func (ctx *GenContext) TemplateForGenType(genType *types.Named) (*template.Template, error) {
	name := genType.Obj().Name()
	fullName := genType.Obj().Pkg().Path() + "." + name
	fmt.Println("looking for template for " + fullName)
	if template, ok := ctx.Templates[fullName]; ok {
		return template, nil
	}

	pos := genType.Obj().Pos()
	fpath := ctx.Fset.Position(pos).Filename
	// TODO: remove all references to path, use filepath instead
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
	ctx.Templates[fullName] = template
	return template, nil
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

// func (ctx *GenContext) AddTemplate(dir string) error {
// 	// search directory for every template in the package
// 	pat := path.Join(dir, "*.tmpl")
// 	paths, err := filepath.Glob(pat)

// 	if err != nil {
// 		return err
// 	}

// 	for _, p := range paths {
// 		base := path.Base(p)
// 		name := base[:len(base)-len(".tmpl")]

// 		t, err := template.New(base).Funcs(templateFunctions).ParseFiles(p)
// 		if err != nil {
// 			return err
// 		}

// 		// Add the template with and without the package name
// 		ctx.Templates[name] = t
// 		dirName := path.Base(dir)
// 		ctx.Templates[dirName+"."+name] = t
// 	}

// 	log.Printf("found %d templates in %s", len(paths), dir)

// 	return nil
// }
