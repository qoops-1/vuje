package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

func completionsFor(doc *simplejson.Json, fullPath []Token) []string {
	if doc == nil {
		return []string{}
	}
	var keyword Token = Key("")
	if len(fullPath) > 0 {
		keyword = fullPath[len(fullPath)-1]
	}
	if idx, ok := keyword.(ErrIndex); ok {
		indexStart := regexp.MustCompile(`\d+`)
		idx := indexStart.FindString(string(idx))
		if !ok {
			return nil
		}
		return []string{fmt.Sprintf("[%s]", idx)}
	}
	userKey, ok := keyword.(Key)
	if !ok {
		return nil
	}
	var matches []string
	keys, err := doc.Map()
	if err != nil {
		return nil
	}
	for jsonKey := range keys {
		if strings.HasPrefix(jsonKey, string(userKey)) {
			matches = append(matches, jsonKey)
		}
	}
	sort.Strings(matches)
	return matches
}

func bestCompletion(query Token, compls []string) string {
	if len(compls) == 0 {
		return ""
	}
	switch t := query.(type) {
	case Key:
		return strings.TrimPrefix(compls[0], string(t))
	case ErrIndex:
		return strings.TrimPrefix(compls[0], string(t))
	}
	return ""
}
