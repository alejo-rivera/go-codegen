package codegen

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func ProcessFile(filePaths ...string) error {
	patterns := make([]string, len(filePaths), len(filePaths))
	for i, filePath := range filePaths {
		if !strings.HasSuffix(filePath, ".go") {
			return errors.New(filePath + " does not reference a go file")
		}
		patterns[i] = fmt.Sprint("file=", filePath)
	}

	fset := token.NewFileSet()
	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedDeps |
			packages.NeedFiles,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return errors.Wrap(err, "parsing file")
	}

	filePathToPkg, err := generatePathToPackageMap(filePaths, pkgs)
	if err != nil {
		return errors.Wrapf(err, "failed to map file paths to packages")
	}

	for filePath, pkg := range filePathToPkg {
		structs := findStructsInFile(filePath, pkg, fset)

		ctx := NewGenContext(fset, pkg.Types)
		for _, s := range structs {
			if err := processStruct(s, ctx); err != nil {
				return errors.Wrapf(err, "processing struct %s", s.Obj().Name())
			}
		}

		if len(ctx.Generated()) == 0 {
			return errors.New("No codegen tags detected in file " + filePath)
		}

		base := filePath[:len(filePath)-len(".go")]
		genPath := base + "_generated.go"
		if err := Output(ctx, genPath); err != nil {
			return errors.Wrap(err, "writing generated code to "+genPath)
		}
		fmt.Printf("Wrote %s.\n", genPath)
	}

	return nil
}

func processStruct(aStruct *types.Named, ctx *GenContext) error {
	invocations, err := InvocationsForStruct(aStruct.Underlying().(*types.Struct))
	if err != nil {
		return errors.Wrap(err, "extracting template invocations")
	}

	for _, invocation := range invocations {
		if err := ctx.RunTemplate(invocation, aStruct); err != nil {
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

func generatePathToPackageMap(filePaths []string, pkgs []*packages.Package) (map[string]*packages.Package, error) {
	filePathToPkg := make(map[string]*packages.Package, len(filePaths))
FilePathLoop:
	for _, filePath := range filePaths {
		for _, pkg := range pkgs {
			for _, pkgFile := range pkg.GoFiles {
				if filePath == pkgFile {
					filePathToPkg[filePath] = pkg
					continue FilePathLoop
				}
			}
		}
		return nil, errors.New("could not find package for file, " + filePath)
	}
	return filePathToPkg, nil
}
