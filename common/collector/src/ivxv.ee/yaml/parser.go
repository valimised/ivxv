package yaml

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// parser is the syntactic analysis state.
type parser struct {
	c      <-chan lexeme     // Channel of lexemes from lexer.
	cont   map[string][]byte // Map of container data.
	crumbs []string          // Breadcrumbs to the current location.
	buf    bytes.Buffer      // Reusable buffer.
}

// Parse parses the root node of the YAML-serialized data. It is an error for
// input to contain more than one root.
//
// If data contains container references, then the data in container is used.
func Parse(data io.Reader, container map[string][]byte) (root Node, err error) {
	return parse(data, container, "")
}

func parse(data io.Reader, container map[string][]byte, path string) (root Node, err error) {
	done := make(chan struct{})
	defer func() {
		close(done)
		err = stopped(err, recover())
	}()

	// Start lexical analysis and create parser.
	c, errc := lex(done, data)
	p := parser{c: c, cont: container}

	// If path is not empty, then use as initial breadcrumb.
	if len(path) > 0 {
		p.push(path)
	}

	// Perform syntactic analysis.
	root, next := p.node(1)

	// Check for trailing data before reading the error, because if the
	// lexer is still waiting to send something and we attempt to read from
	// errc we will get a deadlock. Closing done would also not help,
	// because then we would hide trailing data errors with lexical anaysis
	// cancelled errors.
	if len(next.content) > 0 {
		p.error(next.line, next.column, TrailingDataError{Trailing: next.content})
	}
	err = <-errc
	return
}

// node parses a single node from the lexemes in p.c. It stops at and returns
// the first lexeme that does not belong to the node's block (i.e., is at a
// lesser indentation level than depth).
func (p *parser) node(depth int) (node Node, next lexeme) {
	next, ok := <-p.c
	if !ok || next.column < depth {
		return
	}
	switch next.content {
	case "-":
		return p.sequence(next, false)
	case ":":
		p.error(next.line, next.column, NodeEmptyMappingKeyError{})
		return // Never reached.
	case "|":
		// As a special case, a root level literal indicator increases
		// minimal depth by one.
		if depth == 1 {
			depth++
		}
		return p.literal(depth)
	case "!container":
		return p.reference(depth, next)
	default:
		return p.scalarOrMapping(depth, next)
	}
}

// sequence parses a sequence. l is the first sequence indicator lexeme. The
// method returns the next lexeme that is not part of the sequence.
//
// If mappingDepth is true, then the next lexeme after the sequence can be at
// the same depth as the first indicator: this is used in cases, where the
// sequence indicators are at the same depth as the parent mapping's keys. If
// mappingDepth is false, then it is an error for next to be at the same depth
// as the first sequence indicator.
func (p *parser) sequence(l lexeme, mappingDepth bool) (s Sequence, next lexeme) {
	s.path = p.path()
	depth := l.column
	for {
		// Read the next element of the sequence.
		p.push(fmt.Sprint(len(s.elements)))
		var element Node
		element, next = p.node(depth + 1)
		if element == nil {
			return // p.c was closed.
		}
		s.elements = append(s.elements, element)
		p.pop()

		// Check if the block ended, otherwise we must have a new
		// sequence indicator at the same depth.
		switch {
		case next.column < depth:
			return
		case mappingDepth && next.column == depth && next.content != "-":
			return
		case next.column != depth, next.content != "-":
			p.error(next.line, next.column,
				SequenceIndicatorError{Lexeme: next.content})
		}
	}
}

// literal parses a literal style scalar block with minimal depth. It returns
// the next lexeme that is not part of the scalar.
func (p *parser) literal(depth int) (s Scalar, next lexeme) {
	s.path = p.path()
	p.buf.Reset()
	var endofblock bool
	var written int // How many runes have been written on the current line.
	var pline int   // Previous line.
	for next = range p.c {
		// If next is shallower than depth, then the block ended.
		if next.column < depth {
			endofblock = true
			break
		}

		// The first lexeme sets the previous line and block depth
		// (which is already checked to be greater than or equal to the
		// minimal depth).
		if pline == 0 {
			pline = next.line
			depth = next.column
		}

		// If line has increased, then recreate the intermediate empty
		// lines. Usually the previous line already had a line feed at
		// its end, but special characters (e.g., mapping indicators)
		// do not have them, so check for it.
		var lf int
		if strings.HasSuffix(p.buf.String(), "\n") {
			lf++
		}
		for n := next.line - pline - lf; n > 0; n-- {
			p.buf.WriteByte('\n')
		}

		// If this is the first lexeme on the line, then reset written.
		if next.first {
			written = 0
		}

		// Recreate leading spaces.
		for n := next.column - written - depth; n > 0; n-- {
			p.buf.WriteByte(' ')
			written++
		}

		p.buf.WriteString(next.content)
		written += utf8.RuneCountInString(next.content)
		pline = next.line
	}
	// p.c was closed, next should be empty.
	if !endofblock {
		next = lexeme{}
	}
	s.value = p.buf.String()
	return
}

// reference parses a referenced value. depths is the block depth and l is the
// reference tag read by the caller. The method reads content from the external
// source and returns that as the node.
func (p *parser) reference(depth int, l lexeme) (node Node, next lexeme) {
	// The tag has to be followed by a scalar containing the reference. By
	// passing an empty scalar, we are avoiding scalarOrMapping returning a
	// mapping. However, if there is nothing to read, then it will not
	// return a nil node, but a Scalar with the content of the lexeme,
	// i.e., "". If there is something to read, then it will be prefixed by
	// a space meant to separate the empty lexeme and the first read one.
	refnode, next := p.scalarOrMapping(depth, lexeme{})
	ref := strings.TrimSpace(refnode.(Scalar).String())
	if ref == "" {
		p.error(l.line, l.column, MissingReferenceError{})
	}

	// Get the referenced value.
	data, ok := p.cont[ref]
	if !ok {
		p.error(l.line, l.column, InvalidContainerError{Ref: ref})
	}

	// If the ref ends in ".yaml", then parse data as YAML.
	if strings.HasSuffix(ref, ".yaml") {
		node, err := parse(bytes.NewReader(data), p.cont, p.path())
		if err != nil {
			p.error(l.line, l.column,
				ParseReferencedYAMLError{Reference: ref, Err: err})
		}
		return node, next
	}

	// Otherwise just return data as a scalar.
	return Scalar{path: p.path(), value: string(bytes.TrimSpace(data))}, next
}

// scalarOrMapping parses a scalar, which might later turn out to actually be a
// mapping key. depth is the block depth and l is the first lexeme of the
// scalar that was already read by the caller. The method returns the next
// lexeme that is not part of the scalar or mapping.
func (p *parser) scalarOrMapping(depth int, l lexeme) (node Node, next lexeme) {
	next, ok := <-p.c

	if strings.HasPrefix(l.content, "!!") {
		// Ignore secondary tag handles, which are (most likely) used
		// for types: all our types depend on the structures being
		// applied to, not what the tag specifies.
		l = next
		next, ok = <-p.c
	}

	// Use the value first and only then check if the channel was closed.
	// This is okay because a lexeme zero value will never satisfy the
	// condition. We do this to avoid writing to the buffer if it is never
	// used.
	if next.line == l.line && next.content == ":" {
		// Found a mapping indicator on the same line: l is a key.
		return p.mapping(l)
	}

	p.buf.Reset()
	p.buf.WriteString(strings.TrimSpace(l.content))

	for ok {
		// If next is shallower than depth, then the block ended.
		if next.column < depth {
			break
		}

		// If the first lexeme was not a mapping key, than any
		// following ones cannot be either.
		if next.content == ":" {
			p.error(next.line, next.column, UnexpectedMappingError{})
		}

		p.buf.WriteByte(' ')
		p.buf.WriteString(strings.TrimSpace(next.content))

		next, ok = <-p.c
	}
	return Scalar{path: p.path(), value: p.buf.String()}, next
}

// mapping parses a mapping. key is the first mapping key lexeme. The method
// returns the next lexeme that is not part of the mapping.
func (p *parser) mapping(key lexeme) (m Mapping, next lexeme) {
	m.path = p.path()
	m.pairs = make(map[string]Node)
	depth := key.column
	for {
		// Remove spaces between key and mapping indicator.
		key.content = strings.TrimSpace(key.content)

		// Keys must be unique. Unlike standard YAML, we require the
		// keys to be case-insensitively unique.
		for k := range m.pairs {
			if strings.EqualFold(k, key.content) {
				p.error(key.line, key.column,
					NonUniqueKeyError{Key: key.content})
			}
		}

		// Read the value mapped to the key.
		p.push(key.content)
		var element Node
		element, next = p.node(depth + 1)

		// As a special case, mapping may contain sequences at the same
		// depth as their keys, because the sequence indicator is
		// considered as part of indentation and therefore the sequence
		// block is actually deeper than depth.
		//
		// Allows for:
		//
		//     key:
		//     - first
		//     - second
		//     another key:
		//
		if element == nil && next.column == depth && next.content == "-" {
			element, next = p.sequence(next, true)
		}

		m.pairs[key.content] = element // element may be nil.
		p.pop()

		// Check if the block ended, otherwise we must have a new key
		// at the same depth and a mapping indicator.
		switch {
		case next.column < depth:
			return
		case next.column != depth:
			p.error(next.line, next.column,
				MappingKeyDepthError{Lexeme: next.content})
		case next.content == ":":
			p.error(next.line, next.column, EmptyMappingKeyError{})
		}
		key = next // The lexeme returned by p.node is the next key.

		// Read the mapping indicator and check that it is on the same
		// line as the key.
		var ok bool
		next, ok = <-p.c
		switch {
		case !ok:
			p.error(key.line, key.column, EOFMissingMappingIndicatorError{})
		case next.content != ":":
			p.error(next.line, next.column,
				MappingIndicatorError{Lexeme: next.content})
		case next.line != key.line:
			p.error(next.line, next.column,
				MappingIndicatorLineError{Want: key.line})
		}
	}
}

// error stops the syntactic analysis with the provided error, line, and column
// number.
func (p *parser) error(line, column int, err error) {
	stop(SyntacticAnalysisError{Line: line, Column: column, Err: err})
}

// push pushes an element to the path of the parser.
func (p *parser) push(s string) {
	p.crumbs = append(p.crumbs, s)
}

// pop pops the latest element from the path of the parser.
func (p *parser) pop() {
	p.crumbs = p.crumbs[:len(p.crumbs)-1]
}

// path returns the path to the current location of the parser.
func (p *parser) path() string {
	return strings.Join(p.crumbs, ".")
}
