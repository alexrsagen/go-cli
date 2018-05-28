package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

// ErrNotRunning is returned when a CLI is not running
var ErrNotRunning = errors.New("a CLI is not running")

type pos struct {
	x, y int
}

var closed = true
var prefix = "# "
var curPos, termSize pos
var list CommandList

func clearArea(startPos, endPos pos) {
	if endPos.y-startPos.y < 0 {
		return
	}
	if endPos.x < 0 {
		return
	}

	for y := startPos.y; y <= endPos.y; y++ {
		if y == startPos.y {
			if endPos.y-y > 0 {
				for x := startPos.x; x <= termSize.x; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			} else {
				for x := startPos.x; x <= endPos.x; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			}
		} else {
			if endPos.y-y > 0 {
				for x := 0; x <= termSize.x; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			} else {
				for x := 0; x <= endPos.x; x++ {
					termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
			}
		}
	}
}

func drawText(cursor int, line string) {
	i := 0

	// Draw line contents
	for _, r := range line {
		// Set cursor position
		if i == cursor {
			termbox.SetCursor(curPos.x, curPos.y)
		}

		// Set cell contents
		switch r {
		case '\r':
			curPos.x = 0
			continue
		case '\n':
			curPos.x = 0
			curPos.y++
			continue
		default:
			termbox.SetCell(curPos.x, curPos.y, r, termbox.ColorWhite, termbox.ColorDefault)
		}

		// Move cell
		curPos.x++
		if curPos.x >= termSize.x {
			curPos.x = 0
			curPos.y++
		}
		// XXX: handle curPos.y >= termSize.y

		// Increment cell counter
		i++
	}
	if i == cursor {
		termbox.SetCursor(curPos.x, curPos.y)
	}

	// Flush contents to terminal
	termbox.Flush()
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

	var newArgs []string
	var arg *string

	var inQuote, isEscaped bool
	for i := range args {
		if args[i] == "" {
			continue
		}
		if inQuote || isEscaped {
			*arg += " "
			isEscaped = false
		} else {
			newArgs = append(newArgs, "")
			arg = &newArgs[len(newArgs)-1]
		}
		for _, r := range args[i] {
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
		default:
			*arg += string(r)
			isEscaped = false
		}
	}
	}

	return newArgs
}

// Printf outputs the formatted string to the active CLI
func Printf(format string, a ...interface{}) {
	if closed {
		fmt.Printf(format, a...)
	} else {
		drawText(-1, fmt.Sprintf(format, a...))
	}
}

// Println outputs the operands to the active CLI
func Println(a ...interface{}) {
	if closed {
		fmt.Println(a...)
	} else {
		drawText(-1, fmt.Sprintln(a...))
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
			maxNameLen := 0
			for _, name := range names {
				if len(name) > maxNameLen {
					maxNameLen = len(name)
				}
			}
			maxNameLen += 4

			for _, name := range names {
				if items[name].Handler != nil {
					Printf(strings.Repeat(" ", maxNameLen)+"%s\r%s\n", items[name].Description, name)
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
func SetList(l CommandList) {
	list = l
}

type inputEvent struct {
	Type   termbox.EventType
	Input  string
	Key    termbox.Key
	Cursor int
	Error  error
}

func getInput(startPos pos, cursor int, input string, mask rune) (ev inputEvent) {
	if closed {
		ev.Type = termbox.EventError
		ev.Error = ErrNotRunning
		return
	}

	ev.Input = input
	ev.Cursor = cursor

	switch tev := termbox.PollEvent(); tev.Type {
	case termbox.EventKey:
		ev.Type = termbox.EventKey
		ev.Key = tev.Key

		// Handle keypress
		switch tev.Key {
		case termbox.KeyTab:
			fallthrough
		case termbox.KeyEnd:
			// Move cursor pos to end
			ev.Cursor = utf8.RuneCountInString(ev.Input)
			// Redraw input area
			curPos = startPos
			if mask != 0 {
				drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
			} else {
				drawText(ev.Cursor, ev.Input)
			}

		case termbox.KeyHome:
			// Move cursor pos to start
			ev.Cursor = 0
			// Redraw input area
			curPos = startPos
			if mask != 0 {
				drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
			} else {
				drawText(ev.Cursor, ev.Input)
			}

		case termbox.KeyArrowLeft:
			// Move cursor pos back
			if ev.Cursor > 0 {
				ev.Cursor--
				// Redraw input area
				curPos = startPos
				if mask != 0 {
					drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
				} else {
					drawText(ev.Cursor, ev.Input)
				}
			}

		case termbox.KeyArrowRight:
			// Move cursor pos fwd
			if ev.Cursor < utf8.RuneCountInString(ev.Input) {
				ev.Cursor++
				// Redraw input area
				curPos = startPos
				if mask != 0 {
					drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
				} else {
					drawText(ev.Cursor, ev.Input)
				}
			}

		case termbox.KeyDelete:
			cells := utf8.RuneCountInString(ev.Input)
			if ev.Input != "" && ev.Cursor < cells {
				// Remove character at cursor pos
				pos := bytePos(ev.Cursor, ev.Input)
				ev.Input = ev.Input[:pos] + ev.Input[pos+1:]
				// Redraw input area
				clearArea(startPos, curPos)
				curPos = startPos
				if mask != 0 {
					drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
				} else {
					drawText(ev.Cursor, ev.Input)
				}
			}

		case termbox.KeyBackspace2:
			fallthrough
		case termbox.KeyBackspace:
			if ev.Input != "" && ev.Cursor > 0 {
				// Remove character before cursor pos
				pos := bytePos(ev.Cursor, ev.Input)
				ev.Input = ev.Input[:pos-1] + ev.Input[pos:]
				// Move cursor pos back
				ev.Cursor--
				// Redraw input area
				clearArea(startPos, curPos)
				curPos = startPos
				if mask != 0 {
					drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
				} else {
					drawText(ev.Cursor, ev.Input)
				}
			}

		case termbox.KeySpace:
			tev.Ch = ' '
			fallthrough
		case 0:
			// Insert character at cursor position in current history entry
			pos := bytePos(ev.Cursor, ev.Input)
			ev.Input = ev.Input[:pos] + string(tev.Ch) + ev.Input[pos:]
			// Move cursor pos fwd
			ev.Cursor++
			// Redraw input area
			curPos = startPos
			if mask != 0 {
				drawText(ev.Cursor, strings.Repeat(string(mask), utf8.RuneCountInString(ev.Input)))
			} else {
				drawText(ev.Cursor, ev.Input)
			}
		}

	case termbox.EventResize:
		// Store terminal size
		termSize.x = tev.Width
		termSize.y = tev.Height

	case termbox.EventError:
		// Return error
		ev.Type = termbox.EventError
		ev.Error = tev.Err
	}

	return
}

// Close signals for the CLI to exit on next event
func Close() {
	closed = true
}

// Run sets up a new CLI on the process tty
func Run() error {
	var log history
	var cursor int

	// Reset closed state
	closed = false

	// Initialize terminal
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	// Get initial terminal size
	termW, termH := termbox.Size()
	termSize.x = termW
	termSize.y = termH

	// Draw input area
	curPos = pos{0, 0}
	drawText(-1, prefix)
	startPos := curPos
	// Update cursor position
	drawText(cursor, "")

	for {
		switch ev := getInput(startPos, cursor, log.get(), 0); ev.Type {
		case termbox.EventKey:
			// Clear terminal if new log entry and character was entered
			if log.isLast() && log.get() == "" && ev.Key == 0 {
				clearArea(curPos, termSize)
				termbox.Flush()
			}

			cursor = ev.Cursor
			log.set(ev.Input)

			switch ev.Key {
			case termbox.KeyEnter:
				// Clear terminal
				termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
				curPos = pos{0, 1}

				// Attempt to execute command in current history entry
				if Exec(strings.Split(strings.Trim(log.get(), " "), " ")) {
					if closed {
						return nil
					}

					// If entry is not last, insert new history entry with edited contents and
					// restore any edits to original
					if !log.isLast() {
						log.revertAndAdd()
					}

					log.new()
					cursor = 0
				}

				// Redraw input area
				curPos = pos{0, 0}
				drawText(-1, prefix)
				startPos = curPos
				drawText(cursor, log.get())

			case termbox.KeyTab:
				// Clear terminal
				termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)

				// Autocomplete command in current history entry
				curPos.x = 0
				curPos.y++
				Exec(strings.Split(strings.Trim(log.get()+" ?", " "), " "))

				// Redraw input area
				curPos = pos{0, 0}
				drawText(-1, prefix)
				startPos = curPos
				drawText(cursor, log.get())

			case termbox.KeyArrowUp:
				// If history has a previous entry
				if log.prev() {
					// Clear terminal
					termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
					// Move cursor pos to end
					cursor = utf8.RuneCountInString(log.get())
					// Redraw input area
					curPos = pos{0, 0}
					drawText(-1, prefix)
					startPos = curPos
					drawText(cursor, log.get())
				}

			case termbox.KeyArrowDown:
				// If history has a next entry
				if log.next() {
					// Clear terminal
					termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
					// Move cursor pos to end
					cursor = utf8.RuneCountInString(log.get())
					// Redraw input area
					curPos = pos{0, 0}
					drawText(-1, prefix)
					startPos = curPos
					drawText(cursor, log.get())
				}
			}

		case termbox.EventError:
			return ev.Error
		}
	}
}
