package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	sep := '.'
	tbl := []struct {
		inQuery       string
		outLastEscape bool
		outTokens     []Token
	}{
		{
			inQuery:       "",
			outLastEscape: false,
			outTokens:     []Token{Key("")},
		},
		{
			inQuery:       "key",
			outLastEscape: false,
			outTokens:     []Token{Key("key")},
		},
		{
			inQuery:       "key1.key2",
			outLastEscape: false,
			outTokens:     []Token{Key("key1"), Key("key2")},
		},
		{
			inQuery:       "key[5].key2",
			outLastEscape: false,
			outTokens:     []Token{Key("key"), Index(5), Key("key2")},
		},
		{
			inQuery:       "key[1][10]",
			outLastEscape: false,
			outTokens:     []Token{Key("key"), Index(1), Index(10)},
		},
		{
			inQuery:       "key[1",
			outLastEscape: false,
			outTokens:     []Token{Key("key"), ErrIndex("[1")},
		},
		{
			inQuery:       "key.",
			outLastEscape: false,
			outTokens:     []Token{Key("key"), Key("")},
		},
		{
			inQuery:       `ke\.y`,
			outLastEscape: false,
			outTokens:     []Token{Key("ke.y")},
		},
		{
			inQuery:       `ke\.y.ke\.y2`,
			outLastEscape: false,
			outTokens:     []Token{Key("ke.y"), Key("ke.y2")},
		},
		{
			inQuery:       `ke[y1.key2`,
			outLastEscape: false,
			outTokens:     []Token{Key("ke"), ErrIndex("[y1.key2")},
		},
		{
			inQuery:       `key\`,
			outLastEscape: true,
			outTokens:     []Token{Key("key")},
		},
	}

	for _, tt := range tbl {
		tokens, lastEscape := parseQuery(tt.inQuery, sep)
		assert.Equal(t, tt.outTokens, tokens)
		assert.Equal(t, tt.outLastEscape, lastEscape)
	}
}

func TestQuery_Escape(t *testing.T) {
	query := Query{Sep: '.'}
	actual := query.Escape(`\[strange\.key[.`)
	assert.Equal(t, `\\\[strange\\\.key\[\.`, actual)
}

func TestQuery_CompleteWith(t *testing.T) {
	query := &Query{Sep: '.'}
	query.SetRaw(`key\`)
	query.CompleteWith(`\end.`)
	assert.Equal(t, `key\\end\.`, query.Raw())
	query.SetRaw(`key`)
	query.CompleteWith(`\end`)
	assert.Equal(t, `key\\end`, query.Raw())
}
