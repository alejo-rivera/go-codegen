package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	codegen "github.com/CyborgMaster/go-codegen"
)

var versionFlag = flag.Bool("v", false, "prints the version number")

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
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("expected one or more go files are arguments")
	}

	for _, arg := range args {
		arg, err := filepath.Abs(arg)
		if err != nil {
			log.Fatal(err)
		}

		if err := codegen.ProcessFile(arg); err != nil {
			log.Fatalln(err)
		}
	}
}
