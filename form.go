package cli

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/alexrsagen/termbox-go"
)

type drawableForm interface {
	drawForm()
}

// Field is a structure containing a single form field
type Field struct {
	DisplayName, Input string
	Mask               rune
	Format             *regexp.Regexp
	pos                pos
}

func (f *Field) drawField(maxDNameLen int) {
	if len(f.DisplayName) > 0 {
		Printf("%s:%s    ", f.DisplayName, strings.Repeat(" ", maxDNameLen-len(f.DisplayName)))
		f.pos = curPos
		Println(f.Input)
	}
}

func (f *Field) getInput(cursor int) inputEvent {
	ev := getInput(f.pos, cursor, f.Input, f.Mask)
	switch ev.Type {
	case termbox.EventKey:
		f.Input = ev.Input
	}
	return ev
}

// FieldList is a collection of fields
type FieldList []*Field

func (fl FieldList) getInputs(form drawableForm) {
	if len(fl) == 0 {
		return
	}

	var curField int
	cursor := utf8.RuneCountInString(fl[curField].Input)

	// Update cursor position
	curPos = fl[curField].pos
	drawText(cursor, fl[curField].Input)

	for {
		initPos := curPos

		// Get input
		switch ev := fl[curField].getInput(cursor); ev.Type {
		case termbox.EventKey:
			cursor = ev.Cursor

			if curPos.y != initPos.y {
				// Redraw form
				termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
				form.drawForm()

				// Update cursor position
				curPos = fl[curField].pos
				drawText(cursor, fl[curField].Input)
			}

			switch ev.Key {
			case termbox.KeyEnter:
				// Submit form if on last field
				if curField == len(fl)-1 {
					return
				}
				fallthrough
			case termbox.KeyTab:
				fallthrough
			case termbox.KeyArrowDown:
				if curField < len(fl)-1 {
					curField++
					cursor = utf8.RuneCountInString(fl[curField].Input)

					// Update cursor position
					curPos = fl[curField].pos
					drawText(cursor, fl[curField].Input)
				}
			case termbox.KeyArrowUp:
				if curField > 0 {
					curField--
					cursor = utf8.RuneCountInString(fl[curField].Input)

					// Update cursor position
					curPos = fl[curField].pos
					drawText(cursor, fl[curField].Input)
				}
			}
		case termbox.EventResize:
			// Redraw form
			termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
			form.drawForm()

			// Update cursor position
			curPos = fl[curField].pos
			drawText(cursor, fl[curField].Input)
		}
	}
}

func (fl FieldList) drawForm() {
	maxDNameLen := 0
	for _, f := range fl {
		if len(f.DisplayName) > maxDNameLen {
			maxDNameLen = len(f.DisplayName)
		}
	}
	for _, f := range fl {
		// Render input name
		f.drawField(maxDNameLen)
	}
}

// Form renders a series of input fields to be filled before returning
func (fl FieldList) Form() {
	// Draw form
	fl.drawForm()

	// Get form input
	fl.getInputs(fl)

	// Clear terminal
	termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
	curPos = pos{0, 1}

	// TODO: Validate form input
}

// FieldCategory is a FieldList with a title
type FieldCategory struct {
	DisplayName string
	Fields      FieldList
}

// FieldCategoryList is a collection of field categories
type FieldCategoryList []*FieldCategory

func (fcl FieldCategoryList) drawForm() {
	curPos = pos{0, 0}

	for _, fc := range fcl {
		// Render category title
		Println(fc.DisplayName)

		// Render category form
		fc.Fields.drawForm()
		Println()
	}
}

// Form renders a series of input fields to be filled before returning
func (fcl FieldCategoryList) Form() {
	// Draw form
	fcl.drawForm()

	// Get form input
	var fields FieldList
	for _, fc := range fcl {
		for _, f := range fc.Fields {
			fields = append(fields, f)
		}
	}
	fields.getInputs(fcl)

	// Clear terminal
	termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
	curPos = pos{0, 1}

	// TODO: Validate form input
}
