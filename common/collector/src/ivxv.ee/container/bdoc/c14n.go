package bdoc

import (
	"bytes"
	"fmt"
	"reflect"
)

// c14n contains metadata for a parsed XML element required to reconstruct
// canonical XML from it.
type c14n struct {
	// ns specifies the set of namespace bindings that apply to this
	// element. Since copying this to every element of an XML document
	// would be extremely wasteful, this field is non-nil only for elements
	// that are marked as canonicalization roots. See parse on how to mark
	// roots.
	ns    map[string]string
	start startElement // Start token of the XML element.

	// whitespace contains the whitespace character data between
	// sub-elements. If the XML element contains a character data field,
	// then this is nil. Otherwise it will always contain at least one
	// entry - the whitespace following the start token - and additional
	// entries for the whitespace following each sub-element.
	whitespace []string
}

// isPresent reports if an XML element has been parsed into c14n. Useful for
// checking the presence of optional elements.
func (c c14n) isPresent() bool {
	return len(c.start.name.Local) > 0
}

var c14nType = reflect.TypeOf(c14n{})

// c14nHeader returns a pointer to the c14n header in v.
func c14nHeader(v reflect.Value) *c14n {
	t := v.Type()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprint("cannot process ", t, " as XML element (must be struct)"))
	}
	if t.NumField() < 1 {
		panic(fmt.Sprint(t, " must have at least one field"))
	}
	f := t.Field(0)
	if f.Name != "XMLElement" || f.Type != c14nType {
		panic(fmt.Sprint(t, " must have XMLElement of type c14n as first field"))
	}
	return v.Field(0).Addr().Interface().(*c14n)
}

// writer formats a parsed XML structure as canonical XML as specified in
// https://www.w3.org/TR/xml-c14n11/.
type writer struct {
	// bytes.Buffer is used as the write destination instead of a generic
	// io.Writer because of its guaranteed writes. If a use case appears
	// which requires a generic destination, then writer can be rewritten.
	b  *bytes.Buffer
	ns mapstack
}

// write write the canonical form of the struct reflected in v into p.w. v must
// satisfy the requirements set in parser.parse or write will panic. Cannot use
// encoding/xml.Encoder.EncodeToken here because it writes non-canonical XML.
//
// If root is true, then write is processing the root element of the document
// and the namespace bindings in c14nHeader(v).ns are written.
func (w *writer) write(v reflect.Value, root bool) {
	header := c14nHeader(v)

	// Check that it is properly tagged.
	name, c14nroot := xmlxElement(v.Type().Field(0))
	if name != header.start.name {
		panic(fmt.Sprint(`"xmlx" tag name `, name,
			" does not match start element", header.start.name))
	}
	if root && !c14nroot {
		panic(fmt.Sprint("document root ", v.Type(),
			` "xmlx" tag does not contain "c14nroot" flag`))
	}

	w.ns.push()
	defer w.ns.pop()

	w.writeRaw("<")
	if header.start.nsprefix != "" {
		w.writeRaw(header.start.nsprefix, ":")
	}
	w.writeRaw(header.start.name.Local)
	w.writeAttributes(header, root)
	w.writeRaw(">")

	w.writeSubelements(v, header)

	w.writeRaw("</")
	if header.start.nsprefix != "" {
		w.writeRaw(header.start.nsprefix, ":")
	}
	w.writeRaw(header.start.name.Local, ">")
}

func (w *writer) writeAttributes(header *c14n, root bool) {
	if root {
		ns := make([]attr, 0, len(header.ns))
		for prefix, uri := range header.ns {
			var decl attr
			if prefix == "" {
				decl.Name.Local = xmlns
			} else {
				decl.nsprefix = xmlns
				decl.Name.Local = prefix
			}
			decl.Value = uri
			ns = append(ns, decl)
			w.ns.set(prefix, uri)
		}
		sortAttr(ns)
		for _, decl := range ns {
			w.writeAttribute(decl)
		}
	}
	sortAttr(header.start.attr) // Can do in-place, order only matters for c14n.
	for _, attr := range header.start.attr {
		if attr.namespace() {
			var prefix string
			if attr.nsprefix == xmlns {
				prefix = attr.Name.Local
			}
			if uri, ok := w.ns.get(prefix); ok && uri == attr.Value {
				continue // Skip superfluous declarations.
			}
			if root { // Should never reach if root.
				panic(fmt.Sprintf("XML namespace map does not "+
					"include %q declared by root element", prefix))
			}
			w.ns.set(prefix, attr.Value)
		}
		w.writeAttribute(attr)
	}
}

func (w *writer) writeAttribute(a attr) {
	w.writeRaw(" ")
	if a.nsprefix != "" {
		w.writeRaw(a.nsprefix, ":")
	}
	w.writeRaw(a.Name.Local, `="`)
	w.writeString(a.Value, true)
	w.writeRaw(`"`)
}

func (w *writer) writeSubelements(v reflect.Value, header *c14n) {
	t := v.Type()
	i := 1                        // Skip first field.
	for ; i < v.NumField(); i++ { // Find first non-attribute field.
		if _, attr, _, _ := xmlxAttribute(t.Field(i)); !attr {
			break
		}
	}
	var subfields bool
	for ws := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		ftype := t.Field(i)
		if !subfields { // First field after attributes.
			subfields = true
			if xmlxCharacterData(ftype) {
				w.writeString(field.String(), false)
				if v.NumField() > i+1 {
					panic(fmt.Sprint(t, " must not have ",
						`sub-elements along with character data`))
				}
				break
			}
			w.writeRaw(header.whitespace[ws])
			ws++
		}
		optional := xmlxSubelement(ftype)

		if field.Kind() == reflect.Struct {
			// Skip optional field if it is not present.
			if optional && !c14nHeader(field).isPresent() {
				continue
			}
			w.write(field, false)
			w.writeRaw(header.whitespace[ws])
			ws++
			continue
		}

		// []struct
		if optional && field.IsNil() {
			continue
		}
		for j := 0; j < field.Len(); j++ {
			w.write(field.Index(j), false)
			w.writeRaw(header.whitespace[ws])
			ws++
		}
	}
	if !subfields {
		// Even without subfields there can be internal whitespace.
		w.writeRaw(header.whitespace[0])
	}
}

var (
	escapeAttr = map[byte]string{
		'\t': "&#x9;",
		'\n': "&#xA;",
		'\r': "&#xD;",
		'"':  "&quot;",
		'&':  "&amp;",
		'<':  "&lt;",
	}
	escapeText = map[byte]string{
		'\r': "&#xD;",
		'&':  "&amp;",
		'<':  "&lt;",
		'>':  "&gt;",
	}
)

func (w *writer) writeString(s string, attr bool) {
	// Cannot use encoding/xml.EscapeText since it does not provide
	// canonical output and does not consider context for what to escape.
	esc := escapeText
	if attr {
		esc = escapeAttr
	}
	// Iterate over bytes instead of runes to avoid decode-encode overhead.
	var r int
	for i := 0; i < len(s); i++ {
		if e, ok := esc[s[i]]; ok {
			w.writeRaw(s[r:i], e)
			r = i + 1
		}
	}
	w.writeRaw(s[r:])
}

func (w *writer) writeRaw(str ...string) {
	// Avoid fmt.Fprint which needs to use intermediate buffers.
	for _, s := range str {
		w.b.WriteString(s)
	}
}

func sortAttr(attrs []attr) {
	for i := 1; i < len(attrs); i++ {
		insert := attrs[i]
		j := i
		for ; j > 0 && lessAttr(insert, attrs[j-1]); j-- {
			attrs[j] = attrs[j-1]
		}
		attrs[j] = insert
	}
}

// lessAttr return true if a should come before b in canonical order.
func lessAttr(a, b attr) bool {
	// Default namespace declarations come first.
	if a.nsprefix == "" && a.Name.Local == xmlns {
		return true
	}
	if b.nsprefix == "" && b.Name.Local == xmlns {
		return false
	}

	// Then come namespace declaration sorted by local name.
	if a.nsprefix == xmlns {
		return b.nsprefix != xmlns || a.Name.Local < b.Name.Local
	}
	if b.nsprefix == xmlns {
		return false
	}

	// Remaining attributes are first sorted by namespace URI (not the
	// namespace binding nsprefix!) and secondly by local name.
	return a.Name.Space < b.Name.Space || a.Name.Local < b.Name.Local
}

func writeXML(v interface{}, b *bytes.Buffer) {
	(&writer{b: b}).write(valueOfPtr(v).Elem(), true)
}
