package bdoc

import (
	"bytes"
	"encoding/xml"
	"io"
	"reflect"
	"testing"

	"ivxv.ee/errors"
)

func TestParserToken(t *testing.T) {
	// encoding/xml.Decoder.RawToken handles most things for us: we only
	// need to test namespace translations and start/end tag matching.
	tests := []struct {
		name   string
		input  string
		stream []xml.Token
		err    error // Error that comes after stream.
	}{
		// errors
		{"empty", "", nil, io.EOF},
		{"raw token error", "<foo", nil, new(RawTokenError)},
		{"unclosed element", "<foo>", []xml.Token{
			startElement{name: xml.Name{Local: "foo"}},
		}, new(UnexpectedParseEOFError)},
		{"unexpected close tag", "</foo>", nil, new(UnexpectedEndElementError)},
		{"mismatched element", "<foo></bar>", []xml.Token{
			startElement{name: xml.Name{Local: "foo"}},
		}, new(MismatchingTagsError)},
		{"mismatched element different prefix", `<foo></a:foo>`, []xml.Token{
			startElement{name: xml.Name{Local: "foo"}},
		}, new(MismatchingTagsError)},
		{"empty namespace prefix", `<foo xmlns:="ns"/>`, nil, new(EmptyNamespacePrefixError)},
		{"redeclare xmlns prefix", `<foo xmlns:xmlns="ns"/>`, nil, new(RedeclareXMLNSPrefixError)},
		{"undeclaring namespace", `<foo xmlns=""/>`, nil, new(UndeclaringNamespaceError)},
		{"duplicate attribute", `<foo bar="bar" bar="baz"/>`, nil, new(DuplicateAttributeError)},
		{"duplicate attribute different prefix",
			`<foo xmlns:a="ns" xmlns:b="ns" a:bar="bar" b:bar="baz"/>`,
			nil, new(DuplicateAttributeError)},
		{"undeclared prefix", "<foo:bar/>", nil, new(UndeclaredNamespacePrefixError)},
		{"empty attribute", `<foo bar=""/>`, nil, new(EmptyAttributeError)},

		// translations
		{"default", `<foo xmlns="namespace"/>`, []xml.Token{
			startElement{
				name: xml.Name{Space: "namespace", Local: "foo"},
				attr: []attr{
					{Attr: xml.Attr{
						Name:  xml.Name{Local: "xmlns"},
						Value: "namespace",
					}},
				},
			},
			xml.EndElement{
				Name: xml.Name{Space: "namespace", Local: "foo"},
			},
		}, io.EOF},

		{"named", `<foo:bar xmlns:foo="namespace"/>`, []xml.Token{
			startElement{
				nsprefix: "foo",
				name:     xml.Name{Space: "namespace", Local: "bar"},
				attr: []attr{
					{
						nsprefix: "xmlns",
						Attr: xml.Attr{
							Name:  xml.Name{Space: "xmlns", Local: "foo"},
							Value: "namespace",
						},
					},
				},
			},
			xml.EndElement{
				Name: xml.Name{Space: "namespace", Local: "bar"},
			},
		}, io.EOF},

		{"multiple",
			`<foo:bar xmlns:foo="namespace1" xmlns:baz="namespace2" baz:quux="attr"/>`,
			[]xml.Token{
				startElement{
					nsprefix: "foo",
					name:     xml.Name{Space: "namespace1", Local: "bar"},
					attr: []attr{
						{
							nsprefix: "xmlns",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "xmlns",
									Local: "foo",
								},
								Value: "namespace1",
							},
						},
						{
							nsprefix: "xmlns",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "xmlns",
									Local: "baz",
								},
								Value: "namespace2",
							},
						},
						{
							nsprefix: "baz",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "namespace2",
									Local: "quux",
								},
								Value: "attr",
							},
						},
					},
				},
				xml.EndElement{
					Name: xml.Name{Space: "namespace1", Local: "bar"},
				},
			}, io.EOF},

		{"nested", `<foo xmlns:bar="namespace1">
				<bar:baz xmlns:bar="namespace2"/>
				<bar:quux/>
			</foo>`,
			[]xml.Token{
				startElement{
					name: xml.Name{Local: "foo"},
					attr: []attr{
						{
							nsprefix: "xmlns",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "xmlns",
									Local: "bar",
								},
								Value: "namespace1",
							},
						},
					},
				},
				xml.CharData("\n\t\t\t\t"),
				startElement{
					nsprefix: "bar",
					name:     xml.Name{Space: "namespace2", Local: "baz"},
					attr: []attr{
						{
							nsprefix: "xmlns",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "xmlns",
									Local: "bar",
								},
								Value: "namespace2",
							},
						},
					},
				},
				xml.EndElement{
					Name: xml.Name{Space: "namespace2", Local: "baz"},
				},
				xml.CharData("\n\t\t\t\t"),
				startElement{
					nsprefix: "bar",
					name:     xml.Name{Space: "namespace1", Local: "quux"},
				},
				xml.EndElement{
					Name: xml.Name{Space: "namespace1", Local: "quux"},
				},
				xml.CharData("\n\t\t\t"),
				xml.EndElement{
					Name: xml.Name{Local: "foo"},
				},
			}, io.EOF},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := newParser(bytes.NewReader([]byte(test.input)))
			for _, expected := range test.stream {
				token, err := p.token()
				if err != nil {
					t.Fatal("unexpected error:", err)
				}

				// We cannot use != since startElement contains
				// a slice which is not comparable.
				if !reflect.DeepEqual(token, expected) {
					t.Errorf("unexpected token: %#v; want %#v",
						token, expected)
				}
			}

			_, err := p.token()
			if err != test.err && errors.CausedBy(err, test.err) == nil {
				t.Errorf("unexpected error: %+v; want %+v",
					err, test.err)
			}
		})
	}
}

func TestParseXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		panic    interface{}
		err      error
	}{
		// programmer errors (panic)
		{"nil", "", nil, "cannot process nil as XML root (must be non-nil pointer)", nil},
		{"non-struct", "", int(1), "cannot process int as XML element (must be struct)", nil},
		{"empty struct", "", struct{}{}, "struct {} must have at least one field", nil},
		{
			"first field", "", struct{ first int }{1},
			"struct { first int } must have XMLElement of type c14n as first field",
			nil,
		},
		{
			"untagged", "", struct{ XMLElement c14n }{},
			`XMLElement c14n "xmlx" tag "" must contain namespace and name`, nil,
		},
		{
			"no name", "", struct {
				XMLElement c14n `xmlx:""`
			}{}, `XMLElement c14n "xmlx" tag "" must contain namespace and name`, nil,
		},
		{
			"no namespace", "", struct {
				XMLElement c14n `xmlx:"foo"`
			}{}, `XMLElement c14n "xmlx" tag "foo" must contain namespace and name`, nil,
		},
		{
			"malformed", "", struct {
				XMLElement c14n `xmlx:"foo bar baz"`
			}{}, `malformed "xmlx" tag "foo bar baz"`, nil,
		},
		{
			"unsupported", "", struct {
				XMLElement c14n `xmlx:"foo bar,unsupported"`
			}{}, `unsupported "xmlx" flag "unsupported"`, nil,
		},
		{
			"unallowed flag", "", struct {
				XMLElement c14n `xmlx:"foo bar,attr"`
			}{}, `XMLElement bdoc.c14n tagged with unallowed "xmlx" flag "attr"`, nil,
		},
		{
			"attr chardata", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Attribute  string `xmlx:",attr,chardata"`
			}{}, `Attribute string tagged with unallowed "xmlx" flag "chardata"`, nil,
		},
		{
			"non-string attr", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Attribute  int  `xmlx:",attr"`
			}{}, "Attribute int tagged as XML attribute (must be string)", nil,
		},
		{
			"chardata name", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Value      string `xmlx:"foo baz,chardata"`
			}{}, `XML character data "xmlx" tag must not contain name`, nil,
		},
		{
			"chardata optional", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Value      string `xmlx:",chardata,optional"`
			}{}, `Value string tagged with unallowed "xmlx" flag "optional"`, nil,
		},
		{
			"non-string chardata", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Value      int  `xmlx:",chardata"`
			}{}, "Value int tagged as XML character data (must be string)", nil,
		},
		{
			"chardata subelements", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Value      string `xmlx:",chardata"`
				Subelement struct{}
			}{}, "struct { " +
				`XMLElement bdoc.c14n "xmlx:\"foo bar\""; ` +
				`Value string "xmlx:\",chardata\""; ` +
				`Subelement struct {} ` +
				"} must not have sub-elements along with character data", nil,
		},
		{
			"subelement name", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n     `xmlx:"foo bar"`
				Subelement struct{} `xmlx:"foo sub"`
			}{}, `XML sub-element "xmlx" tag must not contain name`, nil,
		},
		{
			"non-struct subelement", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Value      string
			}{}, "Value string tagged as XML sub-element (must be struct or []struct)", nil,
		},

		// errors
		{
			"EOF", "", struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLStartTokenError),
		},
		{
			"non-element", "character data", struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLNotStartElementError),
		},
		{
			"wrong name", `<baz xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLUnexpectedElementError),
		},
		{
			"missing attribute", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Attribute  string `xmlx:",attr"`
			}{}, nil, new(ParseXMLMissingAttributeError),
		},
		{
			"non-unique attribute", `<bar xmlns="foo">
						<first id="foo"/>
						<second id="foo"/>
					</bar>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				First      struct {
					XMLElement c14n   `xmlx:"foo first"`
					ID         string `xmlx:"id,attr,unique"`
				}
				Second struct {
					XMLElement c14n   `xmlx:"foo second"`
					ID         string `xmlx:"id,attr,unique"`
				}
			}{}, nil, new(ParseXMLNonUniqueAttributeError),
		},
		{
			"extra attribute", `<bar xmlns="foo" baz="baz"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLExtraAttributeError),
		},
		{
			"slice at least one", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Slice      []struct {
					XMLElement c14n `xmlx:"foo slice"`
				}
			}{}, nil, new(ParseXMLSubSliceError),
		},
		{
			"element trailing element", `<bar xmlns="foo"><baz/></bar>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLElementTrailingElementError),
		},
		{
			"element trailing token", `<bar xmlns="foo">trailing</bar>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLElementTrailingTokenError),
		},
		{
			"trailing token", `<bar xmlns="foo"/><baz/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{}, nil, new(ParseXMLTrailingTokenError),
		},

		// parsing
		{
			"simple", `<bar xmlns="foo"/>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{c14n{whitespace: []string{""}}}, nil, nil,
		},
		{
			"procinst and whitespace", xml.Header + " \t<bar xmlns=\"foo\"/>\r\n", struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{c14n{whitespace: []string{""}}}, nil, nil,
		},
		{
			"internal whitespace", "<bar xmlns=\"foo\">\n</bar>", struct {
				XMLElement c14n `xmlx:"foo bar"`
			}{c14n{whitespace: []string{"\n"}}}, nil, nil,
		},
		{
			"c14nroot", `<foo xmlns="default" xmlns:ns="prefixed"/>`, struct {
				XMLElement c14n `xmlx:"default foo,c14nroot"`
			}{c14n{
				ns: map[string]string{
					"":   "default",
					"ns": "prefixed",
				},
				whitespace: []string{""},
			}}, nil, nil,
		},
		{
			"attributes", `<foo ns:full="bar" name="baz" Field="quux"
					xmlns="default" xmlns:ns="namespace"/>`, struct {
				XMLElement c14n   `xmlx:"default foo"`
				Full       string `xmlx:"namespace full,attr"`
				Name       string `xmlx:"name,attr"`
				Field      string `xmlx:",attr"`
				Optional   string `xmlx:",attr,optional"`
			}{
				XMLElement: c14n{whitespace: []string{""}},
				Full:       "bar",
				Name:       "baz",
				Field:      "quux",
			}, nil, nil,
		},
		{
			"character data", `<bar xmlns="foo">character data</bar>`, struct {
				XMLElement c14n   `xmlx:"foo bar"`
				Value      string `xmlx:",chardata"`
			}{
				Value: "character data",
			}, nil, nil,
		},
		{
			"sub-elements", `<bar xmlns="foo">
						<struct/>
						<slice>first</slice>
						<slice>second</slice>
					</bar>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Struct     struct {
					XMLElement c14n `xmlx:"foo struct"`
				}
				Slice []struct {
					XMLElement c14n   `xmlx:"foo slice"`
					Value      string `xmlx:",chardata"`
				}
			}{
				XMLElement: c14n{whitespace: []string{
					"\n\t\t\t\t\t\t",
					"\n\t\t\t\t\t\t",
					"\n\t\t\t\t\t\t",
					"\n\t\t\t\t\t",
				}},
				Struct: struct {
					XMLElement c14n `xmlx:"foo struct"`
				}{c14n{
					start: startElement{
						name: xml.Name{
							Space: "foo",
							Local: "struct",
						},
					},
					whitespace: []string{""},
				}},
				Slice: []struct {
					XMLElement c14n   `xmlx:"foo slice"`
					Value      string `xmlx:",chardata"`
				}{
					{
						XMLElement: c14n{
							start: startElement{
								name: xml.Name{
									Space: "foo",
									Local: "slice",
								},
							},
						},
						Value: "first",
					},
					{
						XMLElement: c14n{
							start: startElement{
								name: xml.Name{
									Space: "foo",
									Local: "slice",
								},
							},
						},
						Value: "second",
					},
				},
			}, nil, nil,
		},
		{
			"optional", `<bar xmlns="foo"><after-optional/></bar>`, struct {
				XMLElement c14n `xmlx:"foo bar"`
				Struct     struct {
					XMLElement c14n `xmlx:"foo struct"`
				} `xmlx:",optional"`
				Slice []struct {
					XMLElement c14n   `xmlx:"foo slice"`
					Value      string `xmlx:",chardata"`
				} `xmlx:",optional"`
				AfterOptional struct {
					XMLElement c14n `xmlx:"foo after-optional"`
				}
			}{
				XMLElement: c14n{whitespace: []string{"", ""}},
				AfterOptional: struct {
					XMLElement c14n `xmlx:"foo after-optional"`
				}{c14n{
					start: startElement{
						name: xml.Name{
							Space: "foo",
							Local: "after-optional",
						},
					},
					whitespace: []string{""},
				}},
			}, nil, nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var v reflect.Value
			var pv interface{}
			if test.expected != nil { // Allows for testing with test.expected == nil.
				v = reflect.New(reflect.TypeOf(test.expected))
				pv = v.Interface()
			}

			defer func() {
				r := recover()
				if r != test.panic {
					t.Errorf("unexpected panic: %#v; want %#v",
						r, test.panic)
				}
			}()

			err := parseXML([]byte(test.input), pv)
			if err != nil || test.err != nil {
				if err != test.err && errors.CausedBy(err, test.err) == nil {
					t.Errorf("unexpected error: %+v; want %T",
						err, test.err)
				}
				return
			}

			// Comparing c14n.start bloats the test table a lot and
			// does not give much extra value, since startElement
			// is already tested in TestParserToken: skip it.
			v.Elem().Field(0).Addr().Interface().(*c14n).start = startElement{}
			parsed := v.Elem().Interface()
			if !reflect.DeepEqual(parsed, test.expected) {
				t.Errorf("unexpected result: %#v; want %#v",
					parsed, test.expected)
			}
		})
	}
}
