package codegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func Process(arg string) error {
	if strings.HasSuffix(arg, ".go") {
		return ProcessFilePath(arg)
	}
	return ProcessDir(arg)
}

// ProcessDir runs the code gen engine against all files in `dir`.
func ProcessDir(dir string) error {
	ctx := NewContext()

	// pkgs, err := parser.ParseDir(ctx.Fset, ctx.Dir, nil, 0)
	// if err != nil {
	// 	return err
	// }

	// for _, pkg := range pkgs {
	// 	for _, file := range pkg.Files {
	// 		err := processFile(ctx, file)

	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return Output(ctx, "main_generated.go")
}

// ProcessFilePath runs the code gen engine against a single file, at `p`
func ProcessFilePath(filePath string) error {
	// ctx := NewContext()

	cfg := &packages.Config{
		Fset: token.NewFileSet(),
		Mode: packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		return errors.Wrap(err, "parsing file")
	}
	if l := len(pkgs); l != 1 {
		return errors.New("expected only 1 package, found " + strconv.Itoa(l))
	}
	structs := findStructsInFile(filePath, pkgs[0], cfg.Fset)
	for _, s := range structs {
		fmt.Printf("%s\n", s)
	}
	// scope := pkg.Types.Scope()
	// for _, name := range scope.Names() {
	// 	object := scope.Lookup(name)
	// 	_, ok := object.Type().Underlying().(*types.Struct)
	// 	if !ok {
	// 		continue
	// 	}
	// 	fpath := cfg.Fset.Position(object.Pos()).Filename
	// 	if fpath != filePath {
	// 		continue
	// 	}
	// }

	return nil
	// fmt.Println("lookup:", t)
	// fmt.Printf("type: %T\n", t)
	// for i := 0; i < t.NumFields(); i++ {
	// 	f := t.Field(i)
	// 	fmt.Println("field:", f)
	// 	fmt.Println("tag:", t.Tag(i))
	// 	fmt.Println("type:", f.Type())
	// }

	// f := t.Field(0)
	// fmt.Println("field:", f)
	// fmt.Println("type:", f.Type())
	// fmt.Printf("typetype: %T\n", f.Type())
	// pos := f.Type().(*types.Named).Obj().Pos()
	// fpath := pkgs[0].Fset.Position(pos).Filename
	// fmt.Println("fst:", path.Dir(fpath))
	// return nil

	// file, err := parser.ParseFile(ctx.Fset, p, nil, 0)
	// if err != nil {
	// 	return err
	// }

	// err = processFile(ctx, file)
	// if err != nil {
	// 	return err
	// }

	// base := path.Base(p)
	// name := base[:len(base)-len(".go")]
	// return Output(ctx, name+"_generated.go")
}

func findStructsInFile(
	filePath string,
	pkg *packages.Package,
	fset *token.FileSet,
) []*types.Named {
	var structs []*types.Named
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		namedType, ok := object.Type().(*types.Named)
		if !ok {
			continue
		}

		// fmt.Printf("object: %s\n", object)
		// fmt.Printf("type: %T\n", object)
		// fmt.Printf("object: %s\n", object.Type())
		// fmt.Printf("type: %T\n", object.Type())
		if _, ok = namedType.Underlying().(*types.Struct); !ok {
			continue
		}
		fpath := fset.Position(object.Pos()).Filename
		if fpath == filePath {
			structs = append(structs, namedType)
		}
	}
	return structs
}

func processFile(ctx *Context, file *ast.File) error {
	// ctx.PackageName = file.Name.Name

	for _, decl := range file.Decls {
		err := processDecl(ctx, decl)
		if err != nil {
			return err
		}
	}

	return nil
}

func processDecl(ctx *Context, decl ast.Decl) error {
	gdp, ok := decl.(*ast.GenDecl)

	if !ok {
		return nil
	}

	for _, spec := range gdp.Specs {
		err := processSpec(ctx, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

func processSpec(ctx *Context, spec ast.Spec) error {
	tsp, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}

	stp, ok := tsp.Type.(*ast.StructType)
	if !ok {
		return nil
	}

	templates, err := ExtractTemplatesFromType(ctx, stp)
	if err != nil {
		return err
	}

	for _, templateName := range templates {
		err := RunTemplate(ctx, templateName, tsp.Name.Name, stp)
		if err != nil {
			return err
		}
	}

	return nil
}

func test(p string) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, "file="+p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	// Print the names of the source files
	// for each package listed on the command line.
	for _, pkg := range pkgs {
		fmt.Println(pkg.ID, pkg.GoFiles)
	}
	scope := pkgs[0].Types.Scope()
	fmt.Println("scope:", scope)
	fmt.Println("names:", scope.Names())
	t := scope.Lookup("Component").Type().Underlying().(*types.Struct)
	fmt.Println("lookup:", t)
	fmt.Printf("type: %T\n", t)
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		fmt.Println("field:", f)
		fmt.Println("tag:", t.Tag(i))
		fmt.Println("type:", f.Type())
	}

	f := t.Field(0)
	fmt.Println("field:", f)
	fmt.Println("type:", f.Type())
	fmt.Printf("typetype: %T\n", f.Type())
	pos := f.Type().(*types.Named).Obj().Pos()
	fpath := pkgs[0].Fset.Position(pos).Filename
	fmt.Println("fst:", path.Dir(fpath))
}
