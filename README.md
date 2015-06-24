# `go-codegen`, a simple code generation system

`go-codegen` is a simple template-based code generation system for go.  By
annotating structs with special-format fields, go-codegen will generate code
based upon templates provided alongside your package.

## Example usage

go-codegen works by building a catalog of available code templates, and then any
anonymous fields on a struct whose type match a template name will be invoked
with a struct that provides information about the struct that the template is
being invoked upon.

Take for example, the following go file:

```go
package main

import "fmt"

//go:generate go-codegen

// cmd is a template.  Blank interfaces are good to use for targeting templates
// as they do not affect the compiled package.
type cmd interface {
	Execute() (interface{}, error)
	MustExecute() interface{}
}

type HelloCommand struct {
	// HelloCommand needs to have the `cmd` template invoked upon it.
	// By mixing in cmd, we tell go-codegen so.
	cmd
	Name string
}

func (cmd *HelloCommand) Execute() (interface{}, error) {
	return "Hello, " + cmd.Name, nil
}

type GoodbyeCommand struct {
	cmd
	Name string
}

func (cmd *GoodbyeCommand) Execute() (interface{}, error) {
	return "Goodbye, " + cmd.Name, nil
}

func main() {
	var c cmd
	c = &HelloCommand{Name: "You"}
	fmt.Println(c.MustExecute())
	c = &GoodbyeCommand{Name: "You"}
	fmt.Println(c.MustExecute())
}

```

Notice that `HelloCommand` doesn't have a `MustExecute` method.  This code will
be generated by `go-codegen`.  Now we need to write the `cmd` template.

Create a new file named `cmd.tmpl` in the same package:

```go-template
// MustExecute behaves like Execute, but panics if an error occurs.
func (cmd *{{ .Name }}) MustExecute() interface{} {
	result, err := cmd.Execute()

  if err != nil {
    panic(err)
  }

  return result
}
```

Notice the `{{ .Name }}` expression:  It's just normal go template code.

Now, given both files, lets generate the code.  Run `go-codegen` in the package
directory, and you'll see a file called `main_generated.go` Whose content looks
like:

```go
package main

// MustExecute behaves like Execute, but panics if an error occurs.
func (cmd *HelloCommand) MustExecute() interface{} {
	result, err := cmd.Execute()

	if err != nil {
		panic(err)
	}

	return result
}

// MustExecute behaves like Execute, but panics if an error occurs.
func (cmd *GoodbyeCommand) MustExecute() interface{} {
	result, err := cmd.Execute()

	if err != nil {
		panic(err)
	}

	return result
}
```



### Arguments

TODO

### Adding Imports

TODO

## Finding templates

At present, a template is searched for within the same directory as a struct
invoking the template, using a name of the form `TemplateName.tmpl`.  For
example, invoking go-codegen in a directory that has a file named "FooBar.tmpl"
will cause that template to be invoked for any struct that has an anonymous
field with a type of `FooBar`.  `FooBar` may be any type: interface, struct or
otherwise.

## Notes on template invocation

TODO
