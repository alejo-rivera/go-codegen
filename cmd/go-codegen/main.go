package main

import (
	"flag"
	"log"
	"path/filepath"

	codegen "github.com/CyborgMaster/go-codegen"
)

func main() {
	flag.Parse()

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
