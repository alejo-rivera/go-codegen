package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	codegen "github.com/CyborgMaster/go-codegen"
)

var versionFlag = flag.Bool("v", false, "prints the version number")
var runtimeFlag = flag.Bool("runtime", false, "prints the go runtime version number")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage of go-codegen: [flags] <paths...>")
		flag.PrintDefaults()
		fmt.Fprintln(
			flag.CommandLine.Output(),
			"  paths: files to parse and generate code for.",
		)
	}
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(codegen.Version)
		return
	} else if *runtimeFlag {
		fmt.Println(runtime.Version())
		return

	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("expected one or more go files are arguments")
	}

	filePaths := make([]string, len(args), len(args))
	for i, arg := range args {
		var err error
		filePaths[i], err = filepath.Abs(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := codegen.ProcessFile(filePaths...); err != nil {
		log.Fatalln(err)
	}
}
