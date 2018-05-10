# go-cli
A simple CLI library written in Go.

The library is intended to be used for a whole application, meaning only one CLI is possible for the whole app.

![Demo GIF](https://raw.githubusercontent.com/alexrsagen/go-cli/master/demo.gif)

## Features
- Command autocompletion
- Bash-like command history
- Easy to use!

## Usage
There aren't many exported methods, so it should be fairly easy to use.

Get started by importing the library:
```go
import "github.com/alexrsagen/go-cli"
```

### List
```go
type List map[string]*Item

func (l List) AddItem(name string, i *Item)
func (l List) RemoveItem(name string)
```

This is a collection of items stored by name.

Example command list:
```go
var list cli.List
list = cli.List{
    "command": &cli.Item{
        Description: "Example command",
        Handler: func(args []string) {
            cli.Println("Example command ran!")
        },
    },
    "submenu": &cli.Item{
        Description: "A nested menu of commands",
        Handler: func(args []string) {
            cli.SetList(list["submenu"].List)
            cli.SetPrefix("my-cli(submenu)# ")
        },
        List: cli.List{
            "command": &cli.Item{
                Description: "Example command",
                Handler: func(args []string) {
                    cli.Println("Submenu example command ran!")
                },
            },
            "return": &cli.Item{
                Description: "Return from submenu context",
                Handler: func(args []string) {
                    cli.SetList(list)
                    cli.SetPrefix("my-cli# ")
                },
            },
        },
    },
    "another_submenu": &cli.Item{
        Description: "Another nested menu of commands",
        Handler: func(args []string) {
            cli.SetList(list["submenu"].List)
            cli.SetPrefix("my-cli(submenu)# ")
        },
        List: cli.List{
            "command": &cli.Item{
                Description: "Example command",
                Handler: func(args []string) {
                    cli.Println("Another submenu example command ran!")
                },
            },
            "return": &cli.Item{
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

### Handler
```go
type Handler func(args []string)
```
This defines the prototype for a function ran when executing an [Item](#item)

### Item
```go
type Item struct {
    Description string
    Arguments   []string
    Handler     Handler
    List        List
}
```

This is a structure for storing a single command item. You cannot store the name of a command inside itself. Use a [List](#list) to store commands by name.

An [Item](#item) containing other items may not have a handler set. **If you do this, it will result in a runtime panic.**

Example command item:
```go
var item *cli.Item
item = &cli.Item{
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
    cli.SetList(cli.List{
        "command": &cli.Item{
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

### SetPrefix
```go
func SetPrefix(s string)
```

This function sets the CLI input prefix string.

Example usage: see [Exec](#exec)

### SetList
```go
func SetList(l List)
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