package codegen

import (
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/types/typeutil"
)

// Context represents the context in which a code generation operation is run.
type GenContext struct {
	PackageName string
	Generated   []string

	templates   map[string]*template.Template
	imports     []string
	importsSeen map[string]struct{}
	fset        *token.FileSet
	packages    map[string]*types.Package
}

func NewGenContext(fset *token.FileSet, rootPackage *types.Package) *GenContext {
	allPackages := typeutil.Dependencies(rootPackage)
	packageMap := make(map[string]*types.Package)
	for _, pkg := range allPackages {
		packageMap[pkg.Path()] = pkg
	}
	return &GenContext{
		PackageName: rootPackage.Name(),
		templates:   make(map[string]*template.Template),
		importsSeen: make(map[string]struct{}),
		fset:        fset,
		packages:    packageMap,
	}
}

func (ctx *GenContext) TemplateForGenType(genType *types.Named) (*template.Template, error) {
	genName := genType.Obj().Name()
	fullName := genType.Obj().Pkg().Path() + "." + genName
	if template, ok := ctx.templates[fullName]; ok {
		return template, nil
	}

	pos := genType.Obj().Pos()
	fpath := ctx.fset.Position(pos).Filename
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

func (ctx *GenContext) GetType(fullName string) (types.Type, error) {
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		return nil, errors.Errorf("%s not a fully qualified type name", fullName)
	}
	pkgName := fullName[:lastDot]
	name := fullName[lastDot+1:]

	pkg, ok := ctx.packages[pkgName]
	if !ok {
		return nil, errors.Errorf("package %s not found", pkgName)
	}
	t := pkg.Scope().Lookup(name)
	if t == nil {
		return nil, errors.Errorf("type %s not found in package %s", name, pkgName)
	}

	// `Lookup` returns a `*types.Named`, we need the underlying type
	return t.Type().Underlying(), nil
}
