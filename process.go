package codegen

import (
	"fmt"
	"go/token"
	"go/types"
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
	ctx := NewGenContext()

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

// ProcessFilePath runs the code gen engine against a single file.
func ProcessFilePath(filePath string) error {
	ctx := NewGenContext()

	cfg := &packages.Config{
		Fset: ctx.Fset,
		// TODO: make sure these all are needed.
		Mode: packages.NeedName |
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
	pkg := pkgs[0]
	ctx.PackageName = pkg.Name
	structs := findStructsInFile(filePath, pkg, ctx.Fset)

	for _, s := range structs {
		if err := processStruct(s, ctx); err != nil {
			return errors.Wrapf(err, "processing struct %s", s.Obj().Name())
		}
	}

	base := filePath[:len(filePath)-len(".go")]
	genPath := base + "_generated.go"
	if err := Output(ctx, genPath); err != nil {
		return errors.Wrap(err, "writing generated code to "+genPath)
	}
	fmt.Printf("Wrote %s.\n", genPath)
	return nil
}

func processStruct(aStruct *types.Named, ctx *GenContext) error {
	invocations, err := InvocationsForStruct(aStruct.Underlying().(*types.Struct))
	if err != nil {
		return errors.Wrap(err, "extracting template invocations")
	}

	for _, invocation := range invocations {
		if err := RunTemplate(invocation, aStruct, ctx); err != nil {
			return errors.Wrap(err, "running template")
		}
	}

	// for _, templateName := range templates {
	// 	err := RunTemplate(ctx, templateName, tsp.Name.Name, stp)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	return nil
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
