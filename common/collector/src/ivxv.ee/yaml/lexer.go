package yaml

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// lexeme is a syntactic unit with metadata about where it occurred.
type lexeme struct {
	content string
	line    int
	column  int
	first   bool // Is this the first lexeme on the line? Simplifies parsing.
}

// lexer is the lexical analysis state.
type lexer struct {
	done <-chan struct{} // Cancellation channel.
	c    chan lexeme     // Output channel.

	b      *bufio.Reader // Buffered input.
	line   int           // Current line.
	column int           // Current column.

	// Keep the booleans in the struct together to minimize padding.
	pending lexeme // Pending lexeme.
	first   bool   // Are we reading the first lexeme on the line?

	buf bytes.Buffer // Reusable buffer.
}

// pend pends the given lexeme for a send, but does not send it yet. Any
// following calls to pend, while the lexeme is not sent, will append content,
// not modifying metadata. Calling lexer.send will force a flush of the pending
// lexeme.
func (l *lexer) pend(content string, line, column int) {
	if len(l.pending.content) == 0 {
		l.pending.line = line
		l.pending.column = column
		l.pending.first = l.first
		l.first = false
	}
	l.pending.content += content
}

var canceled = Canceled{}

// send sends the given lexeme, also flushing any pending sends. Empty content
// will not be sent, only triggering a send of pending content.
func (l *lexer) send(content string, line, column int) {
	if len(l.pending.content) > 0 {
		select {
		case l.c <- l.pending:
		case <-l.done:
			l.error(canceled)
		}
		l.pending.content = ""
		// No need to reset metadata, they will be set on the first
		// call to lexer.pend.
	}
	if len(content) == 0 {
		return
	}
	select {
	case l.c <- lexeme{content, line, column, l.first}:
	case <-l.done:
		l.error(canceled)
	}
	l.first = false
}

// lex starts a new goroutine that does lexical analysis on the input and
// writes lexemes into c. If done is closed, then that cancels the analysis.
// The error return value is written into err.
//
// Callers should close done to exit the goroutine when lexing is cancelled
// without reading all lexemes, so that it does not hang trying to send to c.
// Consumers should read from c until it is closed and then check if the error
// received from err is nil.
func lex(done <-chan struct{}, r io.Reader) (c <-chan lexeme, err <-chan error) {
	l := lexer{
		done:  done,
		c:     make(chan lexeme),
		b:     bufio.NewReader(r),
		line:  1,
		first: true,
	}
	errc := make(chan error, 1)
	go func() {
		defer func() {
			close(l.c)
			errc <- stopped(nil, recover()) // Buffered write, does not block.
		}()

		l.discardBOM()
		for {
			// Skip any spaces between lexemes. We will get
			// indentation information from the column location.
			l.while(" ")

			r, eof := l.rune()
			if eof {
				l.send("", 0, 0) // Flush pending content.
				return
			}
			switch r {
			case '#':
				// Comment: skip until end of line.
				l.until("\n")
				continue

			case '\n':
				// If we have pending content, then append the
				// line feed to it (so the last literal block
				// of a document knows if it ends with a line
				// feed or not). Otherwise skip line feeds, we
				// will get line information from line
				// location.
				if len(l.pending.content) > 0 {
					l.pend("\n", 0, 0)
				}
				l.send("", 0, 0) // Flush pending content.
				l.line++
				l.column = 0
				l.first = true
				continue

			case '-', ':':
				// If followed by a space or line feed, then
				// this is a sequence or mapping indicator.
				next, eof := l.peek()
				if eof || next == ' ' || next == '\n' {
					l.send(string(r), l.line, l.column)
					continue
				}
				// Not an indicator, the '-' or ':' is part of
				// content.
				l.pend(string(r), l.line, l.column)

			case '|':
				// If followed by nothing but spaces or a
				// space-separated comment until end of line,
				// then this is a literal style indicator.
				pos := l.column
				spaces, next, eof := l.while(" ")
				if eof || next == '\n' || (len(spaces) > 0 && next == '#') {
					l.send("|", l.line, pos)
					continue
				}
				// Not an indicator, the '|' and spaces are
				// part of content.
				l.pend("|"+spaces, l.line, pos)

			case '!':
				// If "!container" or a secondary tag followed
				// by a space or line feed, then send this as a
				// tag.
				pos := l.column
				tag, _, _ := l.until(" \n")
				if tag == "container" || strings.HasPrefix(tag, "!") {
					l.send("!"+tag, l.line, pos)
					continue
				}
				// Not a tag, the '!' and read runes are part
				// of content.
				l.pend("!"+tag, l.line, pos)

			default:
				// Start of content, pend the read rune.
				l.pend(string(r), l.line, l.column)
			}

			// By default, unless a case ended with "continue",
			// expect content to follow. Read until the next
			// possible mapping marker, comment, or line feed -- if
			// reading content, then a sequence or literal style
			// indicator or tag cannot occur on the same line
			// unless preceded by a mapping indicator.
			//
			// There is no need to track positional information, as
			// all cases have content already pending and we will
			// use that position.
			content, next, _ := l.until(":#\n")
			l.pend(content, 0, 0)
			for next == '#' && !strings.HasSuffix(l.pending.content, " ") {
				// We hit a regular '#' inside content: remove
				// it from the buffer, pend it, and repeat
				// reading content.
				l.rune()
				l.pend("#", 0, 0)
				content, next, _ = l.until(":#\n")
				l.pend(content, 0, 0)
			}
		}
	}()
	return l.c, errc
}

// discardBOM discards an optional UTF-8 byte order mark.
func (l *lexer) discardBOM() {
	switch bom, err := l.b.Peek(3); err {
	case nil:
		if bytes.Equal(bom, []byte{0xef, 0xbb, 0xbf}) {
			//nolint:errcheck // l.b.Discard(3) is guaranteed to
			// work after l.b.Peek(3), since l.b.Buffered() >= 3.
			l.b.Discard(3)
		}
	case io.EOF:
	default:
		l.error(LexPeekBOMError{Err: err})
	}
}

// rune reads the next rune from input.
func (l *lexer) rune() (r rune, eof bool) {
	l.column++

	// Peek next two bytes to detect \r\n. In that case we discard the \r
	// and only read \n to normalize the line ending. Must happen before
	// ReadRune in order for UnreadRune to work.
	switch rn, err := l.b.Peek(2); err {
	case nil:
		if rn[0] == '\r' && rn[1] == '\n' {
			//nolint:errcheck // l.b.Discard(1) is guaranteed to
			// work after l.b.Peek(2), since l.b.Buffered() >= 1.
			l.b.Discard(1)
		}
	case io.EOF:
	default:
		// Do not rely on the same error happening again on ReadRune.
		// It might have been a temporary failure, which results in
		// ReadRune succeeding and bypassing normalization.
		l.error(LexPeekError{Err: err})
	}

	r, _, err := l.b.ReadRune()
	eof = err == io.EOF
	if err != nil && !eof {
		l.error(LexReadError{Err: err})
	}
	if r == '\r' { // Also normalize a lone \r to \n.
		r = '\n'
	}
	return
}

// unread puts the last read rune back into the buffer.
func (l *lexer) unread() {
	err := l.b.UnreadRune()
	if err != nil {
		l.error(LexUnreadError{Err: err})
	}
	l.column--
}

// peek reads the next rune and immediately unreads it.
func (l *lexer) peek() (r rune, eof bool) {
	// Use rune and unread instead of utf8.DecodeRune(l.b.Peek(4)) to get
	// line-ending normalization.
	r, eof = l.rune()
	if !eof {
		l.unread()
	}
	return
}

// while reads all following runes from input until one that is not present in
// set is found.
func (l *lexer) while(set string) (read string, next rune, eof bool) {
	return l.whileEq(set, true)
}

// until reads all following runes from input until one that is present in set
// is found.
func (l *lexer) until(set string) (read string, next rune, eof bool) {
	return l.whileEq(set, false)
}

// whileEq reads all following runes from input while they either are in set or
// they are not, determined by eq.
func (l *lexer) whileEq(set string, eq bool) (read string, next rune, eof bool) {
	l.buf.Reset()
	r, eof := l.rune()
	for !eof && strings.ContainsRune(set, r) == eq {
		l.buf.WriteRune(r)
		r, eof = l.rune()
	}
	if !eof {
		// Put the triggering rune back.
		l.unread()
	}
	return l.buf.String(), r, eof
}

// error stops the lexical analysis with the provided error, wrapping it with
// the current line and column.
func (l *lexer) error(err error) {
	stop(LexicalAnalysisError{Line: l.line, Column: l.column, Err: err})
}
