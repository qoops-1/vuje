package main

import (
	"fmt"
	"io/ioutil"

	termbox "github.com/nsf/termbox-go"
	"github.com/nwidger/jsoncolor"
)

func colorizeJSON(data []byte) *[][]termbox.Cell {
	formatter := jsoncolor.NewFormatter()
	result := &[][]termbox.Cell{[]termbox.Cell{}}
	regular := &termboxSprintfFuncer{
		fg:     termbox.ColorDefault,
		bg:     termbox.ColorDefault,
		output: result,
	}

	bold := &termboxSprintfFuncer{
		fg:     termbox.AttrBold,
		bg:     termbox.ColorDefault,
		output: result,
	}

	blueBold := &termboxSprintfFuncer{
		fg:     termbox.ColorBlue | termbox.AttrBold,
		bg:     termbox.ColorDefault,
		output: result,
	}

	green := &termboxSprintfFuncer{
		fg:     termbox.ColorGreen,
		bg:     termbox.ColorDefault,
		output: result,
	}

	blackBold := &termboxSprintfFuncer{
		fg:     termbox.ColorBlack | termbox.AttrBold,
		bg:     termbox.ColorDefault,
		output: result,
	}

	formatter.SpaceColor = regular
	formatter.CommaColor = bold
	formatter.ColonColor = bold
	formatter.ObjectColor = bold
	formatter.ArrayColor = bold
	formatter.FieldQuoteColor = blueBold
	formatter.FieldColor = blueBold
	formatter.StringQuoteColor = green
	formatter.StringColor = green
	formatter.TrueColor = regular
	formatter.FalseColor = regular
	formatter.NumberColor = regular
	formatter.NullColor = blackBold

	formatter.Format(ioutil.Discard, data)
	return result
}

type termboxSprintfFuncer struct {
	fg     termbox.Attribute
	bg     termbox.Attribute
	output *[][]termbox.Cell
}

func (tsf *termboxSprintfFuncer) SprintfFunc() func(format string, a ...interface{}) string {
	return func(format string, a ...interface{}) string {
		cells := tsf.output
		idx := len(*cells) - 1
		str := fmt.Sprintf(format, a...)
		for _, s := range str {
			if s == '\n' {
				(*cells) = append((*cells), []termbox.Cell{})
				idx++
				continue
			}
			(*cells)[idx] = append((*cells)[idx], termbox.Cell{
				Ch: s,
				Fg: tsf.fg,
				Bg: tsf.bg,
			})
		}
		return "dummy"
	}
}
