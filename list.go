package cli

import (
	"errors"
	"strings"
)

// ErrInvalidPath is returned when a List path is not found
var ErrInvalidPath = errors.New("invalid path")

// Handler defines the function ran when executing an Item
type Handler func(args []string)

// Item is a structure for storing a single command item.
// You cannot store the name of a command inside itself.
// Use a List to store commands by name.
//
// An Item containing other items may not have a handler set.
type Item struct {
	Description string
	Arguments   []string
	Handler     Handler
	List        List
}

// List is a collection of items stored by name
type List map[string]*Item

// AddItem adds an Item to a List
func (l List) AddItem(name string, i *Item) {
	l[name] = i
}

// RemoveItem removes any Item with the given name from a List
func (l List) RemoveItem(name string) {
	delete(l, name)
}

func (l List) resolvePath(path []string) (possibilities List, args []string, list bool) {
	if path == nil || len(path) == 0 || len(path) == 1 && path[0] == "" {
		return
	}

	argsIndex := 0
	possibilities = List{}
	prefix := ""
	curList := &l
	var curItem *Item

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
			curItem = (*curList)[path[i]]
			curList = &curItem.List
			argsIndex++

			possibilities = List{}
			prefix = strings.Join(path[:i+1], " ")
			possibilities[prefix] = curItem
		} else {
			// Search
			possibilities = List{}
			for name, item := range *curList {
				if strings.HasPrefix(name, path[i]) {
					possibilities[strings.TrimLeft(prefix+" "+name, " ")] = item
				}
			}
			if len(possibilities) == 1 {
				// Single match
				for name, item := range possibilities {
					curItem = item
					curList = &curItem.List
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
		if curItem != nil && len(curItem.Arguments) > 0 {
			panic("parent item cannot have arguments")
		}

		if list || curItem != nil && curItem.Handler == nil {
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
