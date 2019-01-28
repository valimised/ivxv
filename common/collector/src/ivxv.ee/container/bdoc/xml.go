package bdoc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
)

const xmlns = "xmlns"

// parser performs syntactic analysis on an XML document.
//
// We use encoding/xml.Decoder as the XML lexer, but it is too lenient for us
// to use as the parser. E.g., if an element occurs multiple times, then only
// the last element is reported. This is not acceptable for signature parsing.
// Additionally, we want to use the parsed tokens to reconstruct canonical XML,
// but the Decoder discards too much information for this to be possible.
//
// So we resort to parsing the raw stream of XML lexemes ourselves.
type parser struct {
	d     *xml.Decoder
	stack []element
	ns    map[string]string

	// Lookahead values for readToken and unreadToken.
	nextToken xml.Token
	nextError error

	// Uniqueness set for parsed attribute values.
	unique map[string]struct{}
}

func newParser(r io.Reader) *parser {
	return &parser{
		d:      xml.NewDecoder(r),
		ns:     make(map[string]string),
		unique: make(map[string]struct{}),
	}
}

// token returns the next lexeme for syntactic analysis. This method is largely
// a copy of encoding/xml.Decoder.Token with Strict = true, but it returns the
// local startElement instead of xml.StartElement and ignores comments.
func (p *parser) token() (xml.Token, error) {
	t, err := p.d.RawToken()
	if err != nil {
		if err == io.EOF {
			if len(p.stack) > 0 {
				err = UnexpectedParseEOFError{
					Unclosed: p.stack[len(p.stack)-1].name,
				}
			}
			return nil, err
		}
		return nil, RawTokenError{Err: err}
	}
	switch tt := t.(type) {
	case xml.StartElement:
		// Push element and any new namespaces on the stack. Replace
		// with translated startElement.
		if err = p.push(&tt); err != nil {
			return t, err
		}
		ts := startElement{nsprefix: tt.Name.Space, name: tt.Name}
		p.translate(&ts.name.Space, false)
		for _, tta := range tt.Attr {
			a := attr{nsprefix: tta.Name.Space, Attr: tta}
			p.translate(&a.Name.Space, true)
			if findAttr(ts.attr, a.Name) != nil {
				return t, DuplicateAttributeError{
					Element:   ts.name,
					Attribute: a.Name,
				}
			}
			if a.Value == "" { // Allowed by XML, but not us.
				return t, EmptyAttributeError{
					Element:   ts.name,
					Attribute: a.Name,
				}
			}
			ts.attr = append(ts.attr, a)
		}
		t = ts
	case xml.EndElement:
		// Translate the end element using the current namespace. Check
		// that the untranslated name matches the top of the stack and
		// undo any namespace bindings.
		name := tt.Name
		p.translate(&tt.Name.Space, false)
		t, err = tt, p.pop(name)
	case xml.Comment:
		// Skip any comments: they are irrelevant to parsing.
		return p.token()
	}
	return t, err
}

// attr and startElement are like their encoding/xml counterparts, but preserve
// original namespace prefix required for canonical re-encoding.
type attr struct {
	nsprefix string
	xml.Attr
}

func (a attr) namespace() bool {
	// Use nsprefix and not Name.Space which is translated and can cause
	// problems if "xmlns" is used as a namespace URI.
	return a.nsprefix == xmlns ||
		a.nsprefix == "" && a.Name.Local == xmlns
}

type startElement struct {
	nsprefix string
	name     xml.Name
	attr     []attr
}

func (p *parser) translate(space *string, attr bool) {
	// Skip if attribute without a namespace or declares a new namespace.
	if attr && (*space == "" || *space == xmlns) {
		return
	}
	if url, ok := p.ns[*space]; ok {
		*space = url
	}
}

func findAttr(attrs []attr, name xml.Name) *attr {
	for _, a := range attrs {
		// Only compare the name, because two attributes from the same
		// namespace but with different prefixes are still the same.
		if a.Name == name {
			return &a
		}
	}
	return nil
}

// element is an XML element that can define new namespaces. Used in the parser
// stack to keep track of nesting.
type element struct {
	name  xml.Name           // Untranslated.
	oldns map[string]*string // If nil, then the key was not defined before.
}

func (p *parser) push(start *xml.StartElement) error {
	e := element{name: start.Name, oldns: make(map[string]*string)}
	for _, a := range start.Attr {
		var ns string
		switch {
		case a.Name.Space == xmlns:
			switch a.Name.Local {
			case "":
				return EmptyNamespacePrefixError{
					Element: start.Name,
				}
			case xmlns:
				return RedeclareXMLNSPrefixError{
					Element: start.Name,
				}
			}
			ns = a.Name.Local
			fallthrough
		case a.Name.Space == "" && a.Name.Local == xmlns:
			if a.Value == "" {
				return UndeclaringNamespaceError{
					Element:   start.Name,
					Namespace: ns,
				}
			}
			if old, ok := p.ns[ns]; ok {
				e.oldns[ns] = &old
			} else {
				e.oldns[ns] = nil
			}
			p.ns[ns] = a.Value
		}
	}
	p.stack = append(p.stack, e)
	return nil
}

func (p *parser) pop(name xml.Name) error {
	if len(p.stack) == 0 {
		return UnexpectedEndElementError{Name: name}
	}
	e := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	if e.name != name {
		return MismatchingTagsError{Start: e.name, End: name}
	}
	for ns, old := range e.oldns {
		if old == nil {
			delete(p.ns, ns)
		} else {
			p.ns[ns] = *old
		}
	}
	return nil
}

// readToken and unreadToken provide lookahead for p.token.
func (p *parser) readToken() (xml.Token, error) {
	if p.nextToken != nil || p.nextError != nil {
		token, err := p.nextToken, p.nextError
		p.nextToken, p.nextError = nil, nil
		return token, err
	}
	return p.token()
}

func (p *parser) unreadToken(token xml.Token, err error) {
	if p.nextToken != nil || p.nextError != nil {
		panic("double unreadToken")
	}
	if token == nil && err == nil {
		panic("unreadToken(nil, nil)")
	}
	p.nextToken = token
	p.nextError = err
}

// readCharacterData merges and returns all character data until the next
// non-character data token. Note that since readCharacterData ends with an
// unreadToken the returned data cannot be unread.
func (p *parser) readCharacterData() (string, error) {
	buf := buffer()
	defer release(buf)
	for {
		token, err := p.readToken()
		if err != nil {
			return "", err
		}
		cdata, ok := token.(xml.CharData)
		if !ok {
			p.unreadToken(token, err)
			return buf.String(), nil
		}
		buf.Write(cdata)
	}
}

// readWhitespace merges and returns all whitespace-only character data until
// the next non-whitespace token or error. Note that since readWhitespace ends
// with an unreadToken the returned whitespace cannot be unread.
func (p *parser) readWhitespace() string {
	buf := buffer()
	defer release(buf)
	for {
		token, err := p.readToken()
		cdata, ok := token.(xml.CharData)
		index := bytes.IndexFunc(cdata, func(r rune) bool {
			// Limit our set of supported whitespace.
			return r != ' ' && r != '\t' && r != '\n'
		})
		if err != nil || !ok || index >= 0 {
			p.unreadToken(token, err)
			return buf.String()
		}
		buf.Write(cdata)
	}
}

// parse parses the token stream from p and stores the result in the struct
// reflected in v. It works much like a very stripped down version of
// encoding/xml.Decoder.Decode, only supporting structs and slices of structs.
//
// Because the XML tags used by this parser are incompatible with encoding/xml,
// the key for field tags is "xmlx" instead of "xml" to avoid confusion.
//
// Parsing into structs uses the following rules.
//
//   - Only XML elements can be parsed into structs.
//
//   - The first field of all structs must be XMLElement of type c14n with a
//     tag of the form "namespace name" specifying the name of the element
//     being parsed. This replaces the field named XMLName of type xml.Name for
//     structs used with encoding/xml. The tag serves the same purpose, but we
//     are not interested in storing the name (since we know it already) and
//     need canonicalization data instead: forcing both xml.Name and c14n
//     fields would be redundant so they are merged into one.
//
//   - If the tag specified in the last rule is suffixed with ",c14nroot", then
//     it marks the element as a canonicalization root and the ns field of
//     XMLElement will be filled. Structs without this tag suffix cannot be
//     used as canonical XML root-elements.
//
//   - For every exported field in the struct with a tag of the form
//     "namespace name,attr", "name,attr", or ",attr", the field must be of
//     type string and the XML element must contain an attribute with a
//     matching name and non-empty value, unless the tag is suffixed with
//     ",optional". If no namespace is given, then the attribute must not have
//     an explicit namespace prefix. If no name is given, then the name of the
//     field is used. The value of the attribute is stored in the field. If the
//     XML element has non-namespace declaring attributes which do not have
//     corresponding struct fields, then parse returns an error. This differs
//     from encoding/xml, which does not allow specifying the namespace, allows
//     types other than string, allows empty values, has no ",optional" tag
//     (since all attributes are optional), and ignores extra or missing
//     attributes.
//
//   - If the tag specified in the last rule is suffixed with ",unique", then
//     the value of the attribute must be unique compared to all other
//     attributes tagged with ",unique". This is intended to be used to ensure
//     uniqueness of identity attributes in XML elements.
//
//   - All attribute fields must come before any non-attribute fields
//     (excluding XMLElement). This differs from encoding/xml, which allows
//     attribute fields to be interleaved with others.
//
//   - If the struct contains an exported field with a ",chardata" tag, then
//     the field must be of type string and character data within the XML
//     element is stored in it. If such a field exists, then the XML element
//     must not contain any sub-elements. At most one field with this tag is
//     allowed. If no such field exists, then the element must not contain any
//     non-whitespace character data. This differs from encoding/xml, which
//     supports byte slice fields, allows sub-elements by accumulating all
//     character data between them into the field, ignores additional chardata
//     fields, and discards character data not stored anywhere.
//
//   - Every exported field in the struct without a ",attr" or ",chardata" tag
//     must be of type struct or slice of structs (note specifically that
//     pointers are not allowed). The XML element must contain sub-elements
//     matching those fields in the exact same order, unless a field has an
//     ",optional" tag in which case the corresponding sub-element can be
//     missing. If the XML element contains sub-elements which do not have
//     corresponding struct fields, then parse returns an error. This differs
//     from encoding/xml, which supports more types, allows non-matching
//     ordering of the fields and sub-elements, has no ",optional" tag (since
//     all fields are optional), and ignores extra or missing sub-elements.
//
//   - If the field being matched to an XML sub-element is of type struct, then
//     parsing is applied recursively to that struct.
//
//   - If the field being matched to an XML sub-element is of type slice of
//     structs, then one or more XML sub-elements are matched to the struct
//     type of the slice element, creating new entries in the slice.
//
// If v reflects a struct which does not satisfy the type rules set above, then
// parse will panic, as that means there was a programmer error.
//
// If optional is true and the first token is not a start element matching the
// XMLElement name in the struct, then v is not modified, the token is unread,
// and parse returns false, nil.
func (p *parser) parse(v reflect.Value, optional bool) (match bool, err error) {
	header := c14nHeader(v)
	name, c14nroot := xmlxElement(v.Type().Field(0)) // header ensured field 0 exists.

	token, err := p.readToken()
	if err != nil {
		return false, ParseXMLStartTokenError{Element: name, Err: err}
	}
	start, ok := token.(startElement)
	if !ok {
		if optional {
			p.unreadToken(token, err)
			return false, nil
		}
		return false, ParseXMLNotStartElementError{Element: name, Type: fmt.Sprintf("%T", token)}
	}

	if start.name != name {
		if optional {
			p.unreadToken(token, err)
			return false, nil
		}
		return false, ParseXMLUnexpectedElementError{Element: name, Name: start.name}
	}
	if c14nroot { // Copy current namespace bindings.
		header.ns = make(map[string]string)
		for prefix, uri := range p.ns {
			header.ns[prefix] = uri
		}
	}
	header.start = start

	attrn, err := p.parseAttributes(v, start)
	if err != nil {
		return true, err
	}

	if err = p.parseSubelements(v, header, attrn); err != nil {
		return true, err
	}

	// Start and end element matching already done by token(). Just
	// consume the end element, ensuring there are no extra tokens.
	if token, err = p.readToken(); err != nil {
		return true, ParseXMLEndTokenError{Element: name, Err: err}
	}
	if _, ok := token.(xml.EndElement); !ok {
		if more, ok := token.(startElement); ok {
			return true, ParseXMLElementTrailingElementError{
				Element:  name,
				Trailing: more.name,
			}
		}
		return true, ParseXMLElementTrailingTokenError{
			Element: name,
			Type:    fmt.Sprintf("%T", token),
		}
	}
	return true, nil
}

func (p *parser) parseAttributes(v reflect.Value, s startElement) (n int, err error) {
	var vattr []attr
	for i := 1; i < v.NumField(); i++ { // Skip first XMLElement c14n field.
		name, isattr, optional, unique := xmlxAttribute(v.Type().Field(i))
		if !isattr {
			break
		}
		vattr = append(vattr, attr{Attr: xml.Attr{Name: name}})

		found := findAttr(s.attr, name)
		if found == nil {
			if optional {
				continue
			}
			return 0, ParseXMLMissingAttributeError{Element: s.name, Attribute: name}
		}
		if unique {
			if _, ok := p.unique[found.Value]; ok {
				return 0, ParseXMLNonUniqueAttributeError{
					Element:   s.name,
					Attribute: name,
					Value:     found.Value,
				}
			}
			p.unique[found.Value] = struct{}{}
		}
		v.Field(i).SetString(found.Value)
	}
	for _, attr := range s.attr {
		if attr.namespace() {
			continue
		}
		if findAttr(vattr, attr.Name) == nil {
			return 0, ParseXMLExtraAttributeError{
				Element:   s.name,
				Attribute: attr.Name,
			}
		}
	}
	return len(vattr), nil
}

func (p *parser) parseSubelements(v reflect.Value, header *c14n, attrn int) error {
	t := v.Type()
	var subfields bool
	for i := 1 + attrn; i < v.NumField(); i++ { // Skip first field and attributes.
		field := v.Field(i)
		ftype := t.Field(i)
		if !subfields { // First field after attributes.
			subfields = true
			if xmlxCharacterData(ftype) {
				cdata, err := p.readCharacterData()
				if err != nil {
					return ParseXMLCharacterDataTokenError{
						Element: header.start.name,
						Err:     err,
					}
				}
				field.SetString(cdata)

				if v.NumField() > i+1 {
					panic(fmt.Sprint(t, " must not have ",
						`sub-elements along with character data`))
				}
				break
			}
			// Whitespace after start token if no character data.
			header.whitespace = append(header.whitespace, p.readWhitespace())
		}
		optional := xmlxSubelement(ftype)

		if ftype.Type.Kind() == reflect.Struct {
			match, err := p.parse(field, optional)
			if err != nil {
				return ParseXMLSubStructError{
					Element: header.start.name,
					Err:     err,
				}
			}
			if match {
				header.whitespace = append(header.whitespace, p.readWhitespace())
			}
			continue
		}

		// []struct
		for match := true; match; optional = true {
			elem := reflect.New(ftype.Type.Elem()).Elem()
			var err error
			if match, err = p.parse(elem, optional); err != nil {
				return ParseXMLSubSliceError{
					Element: header.start.name,
					Index:   field.Len(),
					Err:     err,
				}
			}
			if match {
				field.Set(reflect.Append(field, elem))
				header.whitespace = append(header.whitespace, p.readWhitespace())
			}
		}
	}
	if !subfields {
		// If there were no fields to parse, then we must still be
		// prepared for whitespace between the start and end token.
		// Must be done after the loop to not consume whitespace that
		// is part of stored character data.
		header.whitespace = append(header.whitespace, p.readWhitespace())
	}
	return nil
}

const xmlx = "xmlx"

const (
	xmlxC14NRoot = 1 << iota
	xmlxAttr
	xmlxOptional
	xmlxUnique
	xmlxCharData
)

func xmlxElement(f reflect.StructField) (name xml.Name, c14nroot bool) {
	name, flags := xmlxTag(f, xmlxC14NRoot)
	if name.Space == "" || name.Local == "" {
		panic(fmt.Sprintf(`XMLElement c14n "xmlx" tag %q must contain namespace and name`,
			f.Tag.Get(xmlx)))
	}
	return name, flags&xmlxC14NRoot > 0
}

func xmlxAttribute(f reflect.StructField) (name xml.Name, attr, optional, unique bool) {
	lookahead := xmlxCharData // Additional flags allowed after attributes.
	name, flags := xmlxTag(f, xmlxAttr|xmlxOptional|xmlxUnique|lookahead)
	if name.Local == "" {
		name.Local = f.Name
	}
	if attr = flags&xmlxAttr > 0; attr {
		if flags&lookahead > 0 { // Not allowed for attr.
			panic(fmt.Sprint(f.Name, " ", f.Type,
				` tagged with unallowed "xmlx" flag "chardata"`))
		}
		if f.Type.Kind() != reflect.String {
			panic(fmt.Sprint(f.Name, " ", f.Type,
				" tagged as XML attribute (must be string)"))
		}
	}
	return name, attr, flags&xmlxOptional > 0, flags&xmlxUnique > 0
}

func xmlxCharacterData(f reflect.StructField) bool {
	lookahead := xmlxOptional // Additional flags allowed instead of chardata.
	name, flags := xmlxTag(f, xmlxCharData|lookahead)
	if flags&xmlxCharData == 0 {
		return false
	}
	if name.Space != "" || name.Local != "" {
		panic(`XML character data "xmlx" tag must not contain name`)
	}
	if flags&lookahead > 0 { // Not allowed for chardata.
		panic(fmt.Sprint(f.Name, " ", f.Type,
			` tagged with unallowed "xmlx" flag "optional"`))
	}
	if kind := f.Type.Kind(); kind != reflect.String {
		panic(fmt.Sprint(f.Name, " ", f.Type,
			" tagged as XML character data (must be string)"))
	}
	return true
}

func xmlxSubelement(f reflect.StructField) (optional bool) {
	name, flags := xmlxTag(f, xmlxOptional)
	if name.Space != "" || name.Local != "" {
		panic(`XML sub-element "xmlx" tag must not contain name`)
	}
	if kind := f.Type.Kind(); kind != reflect.Struct &&
		(kind != reflect.Slice || f.Type.Elem().Kind() != reflect.Struct) {

		panic(fmt.Sprint(f.Name, " ", f.Type,
			" tagged as XML sub-element (must be struct or []struct)"))
	}
	return flags&xmlxOptional > 0
}

// xmlxTag parses the "xmlx" tag of the struct field. Only flags enumerated in
// allowed can be used. Returns the name in the tag and any specified flags.
func xmlxTag(f reflect.StructField, allowed int) (name xml.Name, flags int) {
	tag, ok := f.Tag.Lookup(xmlx)
	if !ok {
		return
	}

	split := strings.Split(tag, " ")
	local := split[0]
	switch len(split) {
	case 1: // No namespace.
	case 2:
		name.Space = split[0]
		local = split[1]
	default:
		panic(fmt.Sprintf(`malformed "xmlx" tag %q`, tag))
	}

	split = strings.Split(local, ",")
	name.Local = split[0]
	for i := 1; i < len(split); i++ {
		var flag int
		switch split[i] {
		case "c14nroot":
			flag = xmlxC14NRoot
		case "attr":
			flag = xmlxAttr
		case "optional":
			flag = xmlxOptional
		case "unique":
			flag = xmlxUnique
		case "chardata":
			flag = xmlxCharData
		default:
			panic(fmt.Sprintf(`unsupported "xmlx" flag %q`, split[i]))
		}
		if allowed&flag == 0 {
			panic(fmt.Sprintf(`%s %s tagged with unallowed "xmlx" flag %q`,
				f.Name, f.Type, split[i]))
		}
		flags |= flag
	}
	return
}

func parseXML(data []byte, v interface{}) error {
	p := newParser(bytes.NewReader(data))

	// Discard the leading XML declaration and any surrounding whitespace.
	token, err := p.readToken()
	pi, ok := token.(xml.ProcInst)
	if err != nil || !ok || pi.Target != "xml" {
		// Not a <?xml ...?> token, parse it as the root element.
		p.unreadToken(token, err)
	}
	p.readWhitespace()

	// Parse the root element.
	if _, err = p.parse(valueOfPtr(v).Elem(), false); err != nil {
		return err
	}

	// Only allow whitespace until EOF.
	p.readWhitespace()
	switch token, err = p.readToken(); {
	case err == io.EOF:
		return nil
	case err == nil:
		return ParseXMLTrailingTokenError{Type: fmt.Sprintf("%T", token)}
	default:
		return ParseXMLTrailingError{Err: err}
	}
}

func valueOfPtr(v interface{}) reflect.Value {
	r := reflect.ValueOf(v)
	var t string
	switch {
	case v == nil:
		t = "nil"
	case r.Kind() != reflect.Ptr:
		t = r.Type().String()
	case r.IsNil():
		t = "(" + r.Type().String() + ")(nil)"
	}
	if t != "" {
		panic(fmt.Sprint("cannot process ", t, " as XML root (must be non-nil pointer)"))
	}
	return r
}
