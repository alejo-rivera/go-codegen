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

func ProcessFile(filePath string) error {
	if !strings.HasSuffix(filePath, ".go") {
		return errors.New(filePath + " does not reference a go file")
	}

	ctx := NewGenContext()

	cfg := &packages.Config{
		Fset: ctx.Fset,
		Mode: packages.NeedName |
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

	if len(ctx.Generated) == 0 {
		return errors.New("No codegen tags detected in file " + filePath)
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
