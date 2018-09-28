package main

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	esc           = '\\'
	asterisk      = Wildcard("*")
	arrayAsterisk = Wildcard("[*]")
)

type Query struct {
	QueryPos   int
	Sep        rune
	Parsed     []Token
	LastEscape bool
	raw        string
}

type Token interface{}

type Key string

type ErrKey string

type Index int

type ErrIndex string

type Wildcard string

func (q *Query) SetRaw(newRaw string) {
	q.raw = newRaw
	q.Parsed, q.LastEscape = parseQuery(q.raw, q.Sep)
}

func (q *Query) Raw() string {
	return q.raw
}

func (q *Query) InsertChar(ch rune) {
	rawSymbols := []rune(q.raw)
	before, after := rawSymbols[:q.QueryPos], rawSymbols[q.QueryPos:]
	withCh := append(append(before, ch), after...)
	q.SetRaw(string(withCh))
	q.QueryPos++
}

func (q *Query) DeleteCurrentChar() {
	if q.QueryPos == 0 {
		return
	}
	rawSymbols := []rune(q.raw)
	before, after := rawSymbols[:q.QueryPos-1], rawSymbols[q.QueryPos:]
	withoutCh := append(before, after...)
	q.SetRaw(string(withoutCh))
	q.QueryPos--
}

func (q *Query) CompleteWith(compl string) {
	var complSuffix string
	switch t := q.Parsed[len(q.Parsed)-1].(type) {
	case Key:
		k := string(t)
		complSuffix = strings.TrimPrefix(compl, k)
	case ErrIndex:
		idx := string(t)
		complSuffix = strings.TrimPrefix(compl, idx)
	}
	if q.LastEscape && len(complSuffix) > 0 {
		complSuffix = complSuffix[:1] + q.Escape(complSuffix[1:])
	} else {
		complSuffix = q.Escape(complSuffix)
	}
	q.SetRaw(q.Raw() + complSuffix)
	q.QueryPos = utf8.RuneCountInString(q.Raw())
}

func (q Query) Escape(s string) string {
	escape := string(esc)
	unescaped := []string{escape, string(q.Sep), "["}
	for _, specialSymbol := range unescaped {
		s = strings.Replace(s, specialSymbol, escape+specialSymbol, -1)
	}
	return s
}

func parseQuery(rawQuery string, sep rune) (tokens []Token, inEscape bool) {
	query := []rune(rawQuery)
	inEscape = false
	var curToken Token = Key("")
	for i := 0; i < len(query); i++ {
		switch {
		case inEscape:
			inEscape = false
			fallthrough
		default:
			curKey, ok := curToken.(Key)
			if !ok {
				tokens = append(tokens, curToken, ErrKey(query[i:]))
				return
			}
			curKey += Key(query[i])
			curToken = curKey
		case query[i] == esc:
			inEscape = true
		case query[i] == sep:
			tokens = append(tokens, curToken)
			curToken = Key("")
		case query[i] == '[':
			tokens = append(tokens, curToken)
			contents, size := parseBrackets(string(query[i:]))
			if size == -1 {
				tokens = append(tokens, contents)
				return
			}
			curToken = contents
			i += size - 1
		}
	}
	tokens = append(tokens, curToken)
	return
}

// parseBrackets returns a parsed token and its length. If length == -1, the token parsing failed
func parseBrackets(q string) (Token, int) {
	if strings.HasPrefix(q, string(arrayAsterisk)) {
		return arrayAsterisk, len(string(arrayAsterisk))
	}
	idxRegex := regexp.MustCompile(`^\[([1-9]\d*|0)\]`)
	match := idxRegex.FindStringSubmatch(q)
	if match == nil {
		return ErrIndex(q), -1
	}
	idx, err := strconv.Atoi(match[1])
	if err != nil {
		termboxFatalf("got \"%s\" instead of number in index", match[0])
	}
	return Index(idx), len(match[0])
}
