package yaml

import (
	"strings"
	"testing"
)

func testLex(t *testing.T, yaml string, lexemes []lexeme) {
	done := make(chan struct{})
	defer close(done)

	var i int
	c, errc := lex(done, strings.NewReader(yaml))
	for l := range c {
		if i >= len(lexemes) {
			t.Errorf("yaml %q, unexpected lexeme: %v", yaml, l)
			continue
		}
		x := lexemes[i]
		if l.content != x.content || l.line != x.line ||
			l.column != x.column || l.first != x.first {
			t.Errorf("yaml %q, unexpected lexeme value: %+v; want %+v", yaml, l, x)
		}
		i++
	}
	if err := <-errc; err != nil {
		t.Fatal("lexical analysis error:", err)
	}
	if i != len(lexemes) {
		t.Errorf("yaml %q, %d missing lexeme(s)", yaml, len(lexemes)-i)
	}
}

func TestLex(t *testing.T) {
	t.Run("comment", func(t *testing.T) {
		testLex(t, "#comment", nil)
		testLex(t, "    # comment\n", nil)
		testLex(t, "content # comment\n", []lexeme{{"content \n", 1, 1, true}})
	})

	t.Run("content", func(t *testing.T) {
		testLex(t, "content", []lexeme{{"content", 1, 1, true}})
		testLex(t, "content\n", []lexeme{{"content\n", 1, 1, true}})
		testLex(t, "    content", []lexeme{{"content", 1, 5, true}})
		testLex(t, "content    ", []lexeme{{"content    ", 1, 1, true}})
		testLex(t, "hello\nworld", []lexeme{{"hello\n", 1, 1, true}, {"world", 2, 1, true}})
		testLex(t, "url#anchor", []lexeme{{"url#anchor", 1, 1, true}})
		testLex(t, "-123", []lexeme{{"-123", 1, 1, true}})
		testLex(t, "-#", []lexeme{{"-#", 1, 1, true}})
		testLex(t, "foo - bar", []lexeme{{"foo - bar", 1, 1, true}})
		testLex(t, ":symbol", []lexeme{{":symbol", 1, 1, true}})
		testLex(t, ":#", []lexeme{{":#", 1, 1, true}})
		testLex(t, "http://example.org", []lexeme{{"http://example.org", 1, 1, true}})
		testLex(t, "a |", []lexeme{{"a |", 1, 1, true}})
		testLex(t, "| b", []lexeme{{"| b", 1, 1, true}})
		testLex(t, "|#", []lexeme{{"|#", 1, 1, true}})
		testLex(t, "a | b", []lexeme{{"a | b", 1, 1, true}})
		testLex(t, "!foo bar", []lexeme{{"!foo bar", 1, 1, true}})
	})

	t.Run("literal", func(t *testing.T) {
		testLex(t, "    |", []lexeme{{"|", 1, 5, true}})
		testLex(t, "|    ", []lexeme{{"|", 1, 1, true}})
		testLex(t, "| # comment", []lexeme{{"|", 1, 1, true}})
	})

	t.Run("mapping", func(t *testing.T) {
		testLex(t, "key: value", []lexeme{
			{"key", 1, 1, true}, {":", 1, 4, false}, {"value", 1, 6, false}})
		testLex(t, "key:\n value", []lexeme{
			{"key", 1, 1, true}, {":", 1, 4, false}, {"value", 2, 2, true}})
	})

	t.Run("sequence", func(t *testing.T) {
		testLex(t, "- sequence", []lexeme{{"-", 1, 1, true}, {"sequence", 1, 3, false}})
		testLex(t, "-\n sequence", []lexeme{{"-", 1, 1, true}, {"sequence", 2, 2, true}})
	})

	t.Run("reference", func(t *testing.T) {
		testLex(t, "!container file.txt", []lexeme{
			{"!container", 1, 1, true}, {"file.txt", 1, 12, false}})
	})

	t.Run("secondary", func(t *testing.T) {
		testLex(t, "!!str 123", []lexeme{
			{"!!str", 1, 1, true}, {"123", 1, 7, false}})
	})

	t.Run("carriage return", func(t *testing.T) {
		testLex(t, "content\r", []lexeme{{"content\n", 1, 1, true}})
		testLex(t, "content\r\n", []lexeme{{"content\n", 1, 1, true}})
		testLex(t, "hello\rworld", []lexeme{
			{"hello\n", 1, 1, true}, {"world", 2, 1, true}})
		testLex(t, "hello\r\nworld", []lexeme{
			{"hello\n", 1, 1, true}, {"world", 2, 1, true}})
	})

	t.Run("bom", func(t *testing.T) {
		testLex(t, "\xefcontent", []lexeme{{"\ufffdcontent", 1, 1, true}})
		testLex(t, "\xef\xbbcontent", []lexeme{{"\ufffd\ufffdcontent", 1, 1, true}})
		testLex(t, "\xef\xbb\xbfcontent", []lexeme{{"content", 1, 1, true}})
	})
}
