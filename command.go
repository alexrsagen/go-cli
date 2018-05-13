package cli

import (
	"errors"
	"strings"
)

// ErrInvalidPath is returned when a CommandList path is not found
var ErrInvalidPath = errors.New("invalid path")

// CommandHandler defines the function ran when executing a Command
type CommandHandler func(args []string)

// Command is a structure for storing a single command item.
// You cannot store the name of a command inside itself.
// Use a CommandList to store commands by name.
//
// A Command containing other commands may not have a handler set.
type Command struct {
	Description string
	Arguments   []string
	Handler     CommandHandler
	List        CommandList
}

// CommandList is a collection of commands stored by name
type CommandList map[string]*Command

func (l CommandList) resolvePath(path []string) (possibilities CommandList, args []string, list bool) {
	if path == nil || len(path) == 0 || len(path) == 1 && path[0] == "" {
		return
	}

	argsIndex := 0
	possibilities = CommandList{}
	prefix := ""
	curList := &l
	var curCmd *Command

	for i := 0; i < len(path); i++ {
		if i == len(path)-1 && path[i] == "?" {
			list = true
			break
		}
		if len(*curList) == 0 {
			break
		}
		if (*curList)[path[i]] != nil {
			// Exact match
			curCmd = (*curList)[path[i]]
			curList = &curCmd.List
			argsIndex++

			possibilities = CommandList{}
			prefix = strings.Join(path[:i+1], " ")
			possibilities[prefix] = curCmd
		} else {
			// Search
			possibilities = CommandList{}
			for name, item := range *curList {
				if strings.HasPrefix(name, path[i]) {
					possibilities[strings.TrimLeft(prefix+" "+name, " ")] = item
				}
			}
			if len(possibilities) == 1 {
				// Single match
				for name, item := range possibilities {
					curCmd = item
					curList = &curCmd.List
					argsIndex++

					prefix = strings.TrimLeft(prefix+" "+name, " ")
					break
				}
			} else {
				list = true
				args = path[argsIndex:]
				return
			}
		}
	}

	if *curList != nil && len(*curList) > 0 {
		if curCmd != nil && len(curCmd.Arguments) > 0 {
			panic("parent item cannot have arguments")
		}

		if list || curCmd != nil && curCmd.Handler == nil {
			for name, item := range *curList {
				possibilities[strings.TrimLeft(prefix+" "+name, " ")] = item
			}
		}
	}

	args = path[argsIndex:]
	if len(args) > 0 && args[len(args)-1] == "?" {
		list = true
	}
	return
}
