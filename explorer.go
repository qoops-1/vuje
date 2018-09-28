package main

import (
	"io"
	"log"
	"unicode/utf8"

	"github.com/bitly/go-simplejson"

	termbox "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

type Explorer struct {
	doc         *simplejson.Json
	display     *Display
	query       *Query
	completions []string
}

func NewExplorer(document io.Reader, sep rune) *Explorer {
	jsonDoc, err := simplejson.NewFromReader(document)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "cant parse json"))
	}
	q := &Query{
		QueryPos: 0,
		Sep:      sep,
		raw:      "",
	}
	return &Explorer{
		doc:   jsonDoc,
		query: q,
		display: &Display{
			DocOffsetY:       0,
			ActiveCompletion: -1,
			OnlyKeys:         false,
		},
		completions: []string{},
	}
}

// ExecuteQuery processes the query in non-interactive mode
func (e *Explorer) ExecuteQuery(query string) *simplejson.Json {
	e.query.SetRaw(query)
	toks := e.query.Parsed
	currentDoc, full := traverse(e.doc, toks)
	if !full {
		log.Fatalf("Bad query: %s", toks[len(toks)-1])
	}
	return currentDoc
}

// Run runs explorer in interactive mode
func (e *Explorer) Run() *simplejson.Json {
	err := termbox.Init()
	defer termbox.Close()
	if err != nil {
		termboxFatalf("failed to initialize termbox: %s", err.Error())
	}
	e.display.Doc = e.doc
	e.query.SetRaw("")
	e.syncWithQuery()
	termbox.Flush()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 0:
				e.symbolInput(ev.Ch)

			case termbox.KeyCtrlV:
				e.nextScreen()

			case termbox.KeyCtrlO:
				e.previousScreen()

			case termbox.KeyCtrlT:
				e.scrollToTop()

			case termbox.KeyCtrlR:
				e.scrollToBottom()

			case termbox.KeyBackspace, termbox.KeyBackspace2:
				e.display.ActiveCompletion = -1
				e.query.DeleteCurrentChar()
				e.syncWithQuery()

			case termbox.KeyTab:
				e.tabComplete()

			case termbox.KeyCtrlP, termbox.KeyArrowUp:
				e.scrollUp()

			case termbox.KeyCtrlN, termbox.KeyArrowDown:
				e.scrollDown()

			case termbox.KeyCtrlF, termbox.KeyArrowRight:
				e.cursorForwards()

			case termbox.KeyCtrlB, termbox.KeyArrowLeft:
				e.cursorBackwards()

			case termbox.KeyCtrlK:
				e.display.ActiveCompletion = -1
				e.deleteAfterCursor()

			case termbox.KeyCtrlG, termbox.KeyEsc:
				e.display.ActiveCompletion = -1
				e.drawCompletions()

			case termbox.KeyCtrlU:
				e.display.ActiveCompletion = -1
				e.deleteBeforeCursor()

			case termbox.KeyCtrlL:
				e.display.ActiveCompletion = -1
				e.display.OnlyKeys = !e.display.OnlyKeys
				e.drawContents(true)

			case termbox.KeyEnter:
				res := e.processEnter()
				if res != nil {
					return res
				}

			case termbox.KeyCtrlC:
				termboxFatalln("stopped with Ctrl+C")

			}
		default:
			e.fullRedraw()
		}
		termbox.Flush()
	}
}

func (e *Explorer) symbolInput(ch rune) {
	e.display.ActiveCompletion = -1
	e.query.InsertChar(ch)
	e.syncWithQuery()
}

func (e *Explorer) tabComplete() {
	if len(e.completions) == 0 {
		return
	}
	if len(e.completions) == 1 {
		e.query.CompleteWith(e.completions[0])
		e.syncWithQuery()
		return
	}
	e.display.ActiveCompletion++
	if e.display.ActiveCompletion >= len(e.completions) {
		e.display.ActiveCompletion = 0
	}
	e.drawCompletions()
}

func (e *Explorer) deleteAfterCursor() {
	beforeCursor := e.query.Raw()[:e.query.QueryPos]
	e.query.SetRaw(beforeCursor)
	e.syncWithQuery()
}

func (e *Explorer) deleteBeforeCursor() {
	afterCursor := e.query.Raw()[e.query.QueryPos:]
	e.query.SetRaw(afterCursor)
	e.query.QueryPos = 0
	e.syncWithQuery()
}

func (e *Explorer) cursorBackwards() {
	if e.query.QueryPos == 0 {
		return
	}
	e.query.QueryPos--
	termbox.SetCursor(utf8.RuneCountInString(prompt)+e.query.QueryPos, promptY)
}

func (e *Explorer) cursorForwards() {
	if e.query.QueryPos == utf8.RuneCountInString(e.query.Raw()) {
		return
	}
	e.query.QueryPos++
	termbox.SetCursor(utf8.RuneCountInString(prompt)+e.query.QueryPos, promptY)
}

func (e *Explorer) processEnter() *simplejson.Json {
	if e.display.ActiveCompletion == -1 {
		return e.display.Doc
	}
	e.query.CompleteWith(e.completions[e.display.ActiveCompletion])
	e.syncWithQuery()
	e.display.ActiveCompletion = -1
	return nil
}

func (e *Explorer) scrollToTop() {
	e.display.DocOffsetY = 0
	e.drawContents(true)
}

func (e *Explorer) scrollToBottom() {
	_, windowHeight := termbox.Size()
	e.display.MoveWindow(e.display.DocHeight, windowHeight)
	e.drawContents(true)
}

func (e *Explorer) nextScreen() {
	_, windowHeight := termbox.Size()
	contentsHeight := windowHeight - (completionY + 1)
	e.display.MoveWindow(contentsHeight, windowHeight)
	e.drawContents(true)
}

func (e *Explorer) previousScreen() {
	_, windowHeight := termbox.Size()
	contentsHeight := windowHeight - (completionY + 1)
	e.display.MoveWindow(-contentsHeight, windowHeight)
	e.drawContents(true)
}

func (e *Explorer) scrollUp() {
	if e.display.DocOffsetY == 0 {
		return
	}
	e.display.DocOffsetY--
	e.drawContents(true)
}

func (e *Explorer) scrollDown() {
	_, windowHeight := termbox.Size()
	e.display.MoveWindow(1, windowHeight)
	e.drawContents(true)
}

func (e *Explorer) syncWithQuery() {
	var full bool
	e.display.Doc, full = traverse(e.doc, e.query.Parsed)
	e.completions = nil
	if !full {
		e.completions = completionsFor(e.display.Doc, e.query.Parsed)
		if e.completions == nil {
			e.display.Doc = nil
		}
	}
	e.fullRedraw()
}

func (e *Explorer) fullRedraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	e.drawQueryLine()
	e.drawCompletions()
	e.drawContents(false)
}

func traverse(doc *simplejson.Json, path []Token) (*simplejson.Json, bool) {
	jdoc := doc
	for step, key := range path {
		switch t := key.(type) {
		case Index:
			i := int(t)
			newDoc, ok := checkGetIndex(jdoc, i)
			if !ok {
				return nil, false
			}
			jdoc = newDoc
		case Key:
			k := string(t)
			newDoc, ok := jdoc.CheckGet(k)
			if !ok {
				if step == len(path)-1 {
					return jdoc, false
				}
				return nil, false
			}
			jdoc = newDoc
		case ErrIndex:
			return jdoc, false
		case ErrKey:
			return nil, false
		}
	}
	return jdoc, true
}

func checkGetIndex(json *simplejson.Json, i int) (*simplejson.Json, bool) {
	if arr, err := json.Array(); err != nil {
		return nil, false
	} else if len(arr) <= i || i < 0 {
		return nil, false
	}
	return json.GetIndex(i), true
}
