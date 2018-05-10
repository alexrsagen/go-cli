package cli

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

var outX, outY, termW, termH int
var prefix = "# "
var list List

func clearArea(startX, startY, endX, endY int) {
	if endY-startY < 0 {
		return
	}
	if endX < 0 {
		return
	}

	for y := startY; y <= endY; y++ {
		if y == 0 {
			if endY-y > 0 {
				for x := startX; x <= termW; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			} else {
				for x := startX; x <= endX; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			}
		} else {
			if endY-y > 0 {
				for x := 0; x <= termW; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			} else {
				for x := 0; x <= endX; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			}
		}
	}
}

func drawText(x, y, cursor int, line string) (int, int) {
	startY := y
	i := 0

	// Draw line contents
	for _, r := range line {
		// Set cursor position
		if i == cursor {
			termbox.SetCursor(x, y)
		}

		// Set cell contents
		switch r {
		case '\r':
			x = 0
			continue
		case '\n':
			x = 0
			y++
			continue
		default:
			termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorDefault)
		}

		// Move cell
		x++
		if x >= termW {
			x = 0
			y++
		}
		// XXX: handle y >= termH

		// Increment cell counter
		i++
	}
	if i == cursor {
		termbox.SetCursor(x, y)
	}

	// Flush contents to terminal
	termbox.Flush()

	return x, y - startY
}

func bytePos(runePos int, s string) int {
	curRunePos := 0
	for curBytePos := range s {
		if curRunePos == runePos {
			return curBytePos
		}
		curRunePos++
	}
	if curRunePos == runePos {
		return len(s)
	}
	// panic because runePos should always be clamped
	panic("rune position outside of string range")
}

func parseArgs(args []string) []string {
	if args == nil || len(args) == 0 {
		return args
	}

	newArgs := make([]string, 1)
	arg := &newArgs[0]
	argStr := strings.Join(args, " ")

	var inQuote, isEscaped bool
	for _, r := range argStr {
		switch r {
		case '\\':
			if isEscaped {
				*arg += "\\"
				isEscaped = false
			} else {
				isEscaped = true
			}
		case '"':
			if isEscaped {
				*arg += "\""
				isEscaped = false
			} else {
				inQuote = !inQuote
			}
		case ' ':
			if inQuote || isEscaped {
				*arg += " "
				isEscaped = false
			} else {
				newArgs = append(newArgs, "")
				arg = &newArgs[len(newArgs)-1]
			}
		default:
			*arg += string(r)
			isEscaped = false
		}
	}

	return newArgs
}

// Printf outputs the formatted string to the active CLI
func Printf(format string, a ...interface{}) {
	if termbox.IsInit {
		var offY int
		outX, offY = drawText(outX, outY, -1, fmt.Sprintf(format, a...))
		outY += offY
	} else {
		fmt.Printf(format, a...)
	}
}

// Println outputs the operands to the active CLI
func Println(a ...interface{}) {
	if termbox.IsInit {
		var offY int
		outX, offY = drawText(outX, outY, -1, fmt.Sprintln(a...))
		outY += offY
	} else {
		fmt.Println(a...)
	}
}

// Exec attempts to execute a single command, and returns true if the command executed
func Exec(path []string) bool {
	items, args, showList := list.resolvePath(path)
	if items == nil {
		// Do nothing
	} else if len(items) == 0 {
		// Print command not found message
		Println("Command not found")
	} else {
		if len(items) == 1 && !showList {
			// Execute item handler
			for name, item := range items {
				if item.Handler != nil {
					args = parseArgs(args)
					if args == nil || len(args) != len(item.Arguments) {
						// Print usage message
						Printf("Usage: %s", name)
						for _, arg := range item.Arguments {
							Printf(" <%s>", arg)
						}
						Printf("\n")
					} else {
						item.Handler(args)
						return true
					}
				}
				break
			}
		} else {
			// Get item keys
			var names []string
			for name := range items {
				names = append(names, name)
			}

			// Sort item keys alphabetically
			sort.Strings(names)

			// List sorted items
			longestName := 0
			for _, name := range names {
				if len(name) > longestName {
					longestName = len(name)
				}
			}
			longestName += 4

			for _, name := range names {
				if items[name].Handler != nil {
					Printf(strings.Repeat(" ", longestName)+"%s\r%s\n", items[name].Description, name)
				}
			}
		}
	}

	return false
}

// SetPrefix sets the CLI input prefix string
func SetPrefix(s string) {
	prefix = s
}

// SetList sets the CLI command list
func SetList(l List) {
	list = l
}

// Run sets up a new CLI on the process tty
func Run() error {
	var log history
	var offX, offY, cursor int

	// Initialize terminal
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	// Get initial terminal size
	termW, termH = termbox.Size()

	// Redraw input area
	offX, offY = drawText(0, 0, -1, prefix)
	offX, offY = drawText(offX, offY, cursor, log.get())

	for {
		// Reset output position
		outX = 0
		outY = 1

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			// Handle keypress
			switch ev.Key {
			case termbox.KeyTab:
				// Clear terminal
				clearArea(0, 0, termW, termH)
				outX = 0
				outY = 1
				// Move cursor pos to end
				cursor = utf8.RuneCountInString(log.get())
				// Autocomplete command in current history entry
				Exec(strings.Split(strings.Trim(log.get()+" ?", " "), " "))
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyEnter:
				// Clear terminal
				clearArea(0, 0, termW, termH)
				outX = 0
				outY = 1
				// Attempt to execute command in current history entry
				if Exec(strings.Split(strings.Trim(log.get(), " "), " ")) {
					// If entry is not last, insert new history entry with edited contents and
					// restore any edits to original
					if !log.isLast() {
						log.revertAndAdd()
					}
					log.new()
					cursor = 0
				}
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyHome:
				// Move cursor pos to start
				cursor = 0
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyEnd:
				// Move cursor pos to end
				cursor = utf8.RuneCountInString(log.get())
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyArrowLeft:
				// Move cursor pos back
				if cursor > 0 {
					cursor--
				}
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyArrowRight:
				// Move cursor pos fwd
				if cursor < utf8.RuneCountInString(log.get()) {
					cursor++
				}
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())

			case termbox.KeyArrowUp:
				// If history has a previous entry
				if log.prev() {
					// Clear input area
					clearArea(0, 0, offX, offY)
					// Move cursor pos to end
					cursor = utf8.RuneCountInString(log.get())
					// Redraw input area
					offX, offY = drawText(0, 0, -1, prefix)
					offX, offY = drawText(offX, offY, cursor, log.get())
				}

			case termbox.KeyArrowDown:
				// If history has a next entry
				if log.next() {
					// Clear input area
					clearArea(0, 0, offX, offY)
					// Move cursor pos to end
					cursor = utf8.RuneCountInString(log.get())
					// Redraw input area
					offX, offY = drawText(0, 0, -1, prefix)
					offX, offY = drawText(offX, offY, cursor, log.get())
				}

			case termbox.KeyDelete:
				cells := utf8.RuneCountInString(log.get())
				if log.get() != "" && cursor < cells {
					// Remove character at cursor pos
					pos := bytePos(cursor, log.get())
					log.set(log.get()[:pos] + log.get()[pos+1:])
					// Maintain cursor pos
					// Clear input area
					clearArea(0, 0, offX, offY)
					// Redraw input area
					offX, offY = drawText(0, 0, -1, prefix)
					offX, offY = drawText(offX, offY, cursor, log.get())
				}

			case termbox.KeyBackspace2:
				fallthrough
			case termbox.KeyBackspace:
				if log.get() != "" && cursor > 0 {
					// Remove character before cursor pos
					pos := bytePos(cursor, log.get())
					log.set(log.get()[:pos-1] + log.get()[pos:])
					// Move cursor pos back
					cursor--
					// Clear input area
					clearArea(0, 0, offX, offY)
					// Redraw input area
					offX, offY = drawText(0, 0, -1, prefix)
					offX, offY = drawText(offX, offY, cursor, log.get())
				}

			case termbox.KeySpace:
				ev.Ch = ' '
				fallthrough
			case 0:
				// Insert character at cursor position in current history entry
				pos := bytePos(cursor, log.get())
				log.set(log.get()[:pos] + string(ev.Ch) + log.get()[pos:])
				// Move cursor pos fwd
				cursor++
				// Redraw input area
				offX, offY = drawText(0, 0, -1, prefix)
				offX, offY = drawText(offX, offY, cursor, log.get())
			}

		case termbox.EventResize:
			// Store terminal size
			termW = ev.Width
			termH = ev.Height

		case termbox.EventError:
			// Return error
			return ev.Err
		}
	}
}
