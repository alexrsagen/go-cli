# go-cli
A simple CLI library written in Go.

The library is intended to be used for a whole application, meaning only one CLI is possible for the whole app.

![Demo GIF](https://raw.githubusercontent.com/alexrsagen/go-cli/master/demo.gif)

## Features
- Command autocompletion
- Bash-like command history
- Easy to use!

## Usage
```go
import "github.com/alexrsagen/go-cli"
```

Get started by taking a look at the example usage in [Exec](#exec).

### CommandList
```go
type CommandList map[string]*Command
```

This is a collection of commands stored by name.

Example command list:
```go
var list cli.CommandList
list = cli.CommandList{
    "command": &cli.Command{
        Description: "Example command",
        Handler: func(args []string) {
            cli.Println("Example command ran!")
        },
    },
    "submenu": &cli.Command{
        Description: "A nested menu of commands",
        Handler: func(args []string) {
            cli.SetList(list["submenu"].List)
            cli.SetPrefix("my-cli(submenu)# ")
        },
        List: cli.CommandList{
            "command": &cli.Command{
                Description: "Example command",
                Handler: func(args []string) {
                    cli.Println("Submenu example command ran!")
                },
            },
            "return": &cli.Command{
                Description: "Return from submenu context",
                Handler: func(args []string) {
                    cli.SetList(list)
                    cli.SetPrefix("my-cli# ")
                },
            },
        },
    },
    "another_submenu": &cli.Command{
        Description: "Another nested menu of commands",
        Handler: func(args []string) {
            cli.SetList(list["submenu"].List)
            cli.SetPrefix("my-cli(submenu)# ")
        },
        List: cli.CommandList{
            "command": &cli.Command{
                Description: "Example command",
                Handler: func(args []string) {
                    cli.Println("Another submenu example command ran!")
                },
            },
            "return": &cli.Command{
                Description: "Return from submenu context",
                Handler: func(args []string) {
                    cli.SetList(list)
                    cli.SetPrefix("my-cli# ")
                },
            },
        },
    },
}
```

### CommandHandler
```go
type CommandHandler func(args []string)
```
This defines the prototype for a function ran when executing a [Command](#command)

### Command
```go
type Command struct {
    Description string
    Arguments   []string
    Handler     CommandHandler
    List        CommandList
}
```

This is a structure for storing a single command item. You cannot store the name of a command inside itself. Use a [CommandList](#commandlist) to store commands by name.

A [Command](#command) containing other commands may not have a handler set. **If you do this, it will result in a runtime panic.**

Example command item:
```go
var item *cli.Command
item = &cli.Command{
    Description: "Example command",
    Handler: func(args []string) {
        cli.Println("Example command ran!")
    },
}
```

### Exec
```go
func Exec(path []string) bool
```

This function attempts to execute a single command, and returns true if the command executed.

Example usage:
```go
import "os"
import "github.com/alexrsagen/go-cli"

func main() {
    cli.SetList(cli.CommandList{
        "command": &cli.Command{
            Description: "Example command",
            Handler: func(args []string) {
                cli.Println("Example command ran!")
            },
        }
    })

    // Default prefix is "# ", but you can change it like so:
    // cli.SetPrefix("my-cli# ")

    // Executes a command directly, if one is given in arguments.
    // Otherwise creates a CLI.
    if len(os.Args) > 1 {
        if !cli.Exec(os.Args[1:]) {
            os.Exit(1)
        }
    } else {
        err := cli.Run()
        if err != nil {
            panic(err) // Received an error event from termbox
        }
    }
}
```

### Field
```go
type Field struct {
	DisplayName, Input string
	Mask               rune
	Format             *regexp.Regexp
}
```

This is a structure containing a single form field.

### FieldCategory
```go
type FieldCategory struct {
	DisplayName string
	Fields      FieldList
}
```

This is a [FieldList](#fieldlist) with a title.

### FieldList
```go
type FieldList []*Field
```

This is a collection of form fields.

### FieldCategoryList
```go
type FieldCategoryList []*FieldCategory
```

This is a collection of field categories.

### Form
```go
func (fl FieldList) Form()
func (fcl FieldCategoryList) Form()
```

These functions render a series of input fields to be filled before returning. Should be used within a [CommandHandler](#commandhandler). The FieldCategoryList `Form()` function also renders its category titles.

Example usage:
```go
func myHandler(args []string) {
    field1 := &cli.Field{DisplayName: "Field 1"}
    field2 := &cli.Field{DisplayName: "Field 2"}
    field3 := &cli.Field{DisplayName: "Field 3"}
    field4 := &cli.Field{DisplayName: "Field 4"}
    field5 := &cli.Field{DisplayName: "Field 5"}
    field6 := &cli.Field{DisplayName: "Field 6"}
    
    fcl := cli.FieldCategoryList{
        &cli.FieldCategory{
            DisplayName: "Category 1",
            Fields: cli.FieldList{
                field1,
                field2,
                field3,
            },
        },
        &cli.FieldCategory{
            DisplayName: "Category 2",
            Fields: cli.FieldList{
                field4,
                field5,
                field6,
            },
        },
    }
    
    fcl.Form()
    
    cli.Println("Form filled!\n")
    cli.Printf("%s: %s\n", field1.DisplayName, field1.Input)
    cli.Printf("%s: %s\n", field2.DisplayName, field2.Input)
    cli.Printf("%s: %s\n", field3.DisplayName, field3.Input)
    cli.Printf("%s: %s\n", field4.DisplayName, field4.Input)
    cli.Printf("%s: %s\n", field5.DisplayName, field5.Input)
    cli.Printf("%s: %s\n", field6.DisplayName, field6.Input)
}
```

### SetPrefix
```go
func SetPrefix(s string)
```

This function sets the CLI input prefix string.

Example usage: see [Exec](#exec)

### SetList
```go
func SetList(l CommandList)
```

This function sets the CLI command list.

Example usage: see [Exec](#exec)

### Run
```go
func Run() error
```

This function sets up a new CLI on the process tty.

Example usage: see [Exec](#exec)

### Printf
```go
func Printf(format string, a ...interface{})
```

A wrapper around [fmt.Sprintf](https://golang.org/pkg/fmt/#Sprintf).

The point of the wrapper function is to be able to correctly write the output to the terminal created by [termbox](https://github.com/nsf/termbox-go). Falls back to calling [fmt.Printf](https://golang.org/pkg/fmt/#Printf) directly when a terminal is has not been started.

### Println
```go
func Println(a ...interface{})
```

A wrapper around [fmt.Sprintln](https://golang.org/pkg/fmt/#Sprintln).

The point of the wrapper function is to be able to correctly write the output to the terminal created by [termbox](https://github.com/nsf/termbox-go). Falls back to calling [fmt.Println](https://golang.org/pkg/fmt/#Println) directly when a terminal is has not been started.