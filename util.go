package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Row is a JSON object.
type Row = map[string]interface{}

func trimSpace(b []byte) []byte {
	return bytes.TrimFunc(b, func(r rune) bool { return unicode.IsSpace(r) })
}

func bytesReader(b []byte) io.Reader {
	return bytes.NewReader(b)
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	return strings.ReplaceAll(s, "\n", " ")
}
