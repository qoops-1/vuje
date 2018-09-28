package main

import (
	"bytes"
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/bitly/go-simplejson"

	runewidth "github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

const (
	prompt      = ">>> "
	promptY     = 0
	completionY = promptY + 1
)

func drawString(offsetY int, contents string, fgColor termbox.Attribute, bgColor termbox.Attribute) {
	var cells []termbox.Cell
	for _, ch := range contents {
		cells = append(cells, termbox.Cell{Ch: ch, Fg: fgColor, Bg: bgColor})
	}
	drawLine(offsetY, cells)
}

func drawLine(offsetY int, cells []termbox.Cell) {
	for i, c := range cells {
		termbox.SetCell(i, offsetY, c.Ch, c.Fg, c.Bg)
	}
}

func (e *Explorer) drawCompletions() {
	if e.display.ActiveCompletion == -1 || len(e.completions) < 2 {
		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		var cells []termbox.Cell
		w, _ := termbox.Size()
		for i := 0; i < w; i++ {
			cells = append(cells, termbox.Cell{Ch: ' ', Fg: fg, Bg: bg})
		}
		drawLine(completionY, cells)
		return
	}
	var cells []termbox.Cell
	for i, key := range e.completions {
		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		if i == e.display.ActiveCompletion {
			fg = termbox.ColorDefault | termbox.AttrBold
		}
		for _, char := range key {
			cells = append(cells, termbox.Cell{Ch: char, Fg: fg, Bg: bg})
		}
		cells = append(cells, termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault})
	}
	drawLine(completionY, cells)
}

func (e *Explorer) drawQueryLine() {
	termbox.SetCursor(e.query.QueryPos+runewidth.StringWidth(prompt), promptY)
	var lastToken Token = Key("")
	if len(e.query.Parsed) != 0 {
		lastToken = e.query.Parsed[len(e.query.Parsed)-1]
	}
	firstCompl := bestCompletion(lastToken, e.completions)
	promptLen := utf8.RuneCountInString(prompt)
	queryLen := utf8.RuneCountInString(e.query.Raw())
	for x, symbol := range prompt + e.query.Raw() + firstCompl {
		textColor := termbox.ColorDefault
		if x >= promptLen && x < promptLen+queryLen {
			textColor = termbox.ColorBlue
		} else if x > promptLen+queryLen {
			textColor = termbox.ColorGreen
		}
		termbox.SetCell(x, promptY, symbol, textColor, termbox.ColorDefault)
	}
}

func (e *Explorer) drawContents(clear bool) {
	if clear {
		_, h := termbox.Size()
		for y := completionY + 1; y < h; y++ {
			clearLine(y)
		}
	}
	if e.display.Doc == nil {
		for x, ch := range "--- no results ---" {
			termbox.SetCell(x, completionY+1, ch, termbox.ColorRed, termbox.ColorDefault)
		}
		return
	}
	if e.display.OnlyKeys {
		displayKeys(e.display.Doc)
		return
	}
	json, err := e.display.Doc.EncodePretty()
	if err != nil {
		termboxFatalln(err)
	}
	e.display.DocHeight = bytes.Count(json, []byte("\n")) + 1
	JSONcells := *colorizeJSON(json)
	for i, line := range JSONcells[e.display.DocOffsetY:] {
		drawLine(completionY+1+i, line)
	}
}

func displayKeys(doc *simplejson.Json) {
	obj, err := doc.Map()
	if err != nil {
		idxs, err := doc.Array()
		if err != nil {
			drawString(completionY+1, "--- not an object or array ---", termbox.ColorRed, termbox.ColorDefault)
			return
		}
		drawString(completionY+1, fmt.Sprintf("%d..%d", 0, len(idxs)-1), termbox.ColorDefault, termbox.ColorDefault)
		return
	}
	var keys []string
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		drawString(completionY+1+i, k, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func clearLine(y int) {
	maxX, _ := termbox.Size()
	for i := 0; i < maxX; i++ {
		termbox.SetCell(i, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}
