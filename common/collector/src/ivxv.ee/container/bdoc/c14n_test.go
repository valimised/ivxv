package bdoc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestWriteXML(t *testing.T) {
	tests := []struct {
		name     string
		element  interface{}
		expected string
	}{
		{
			"simple", &struct {
				XMLElement c14n `xmlx:"default foo,c14nroot"`
			}{c14n{
				ns:         make(map[string]string),
				start:      startElement{name: xml.Name{Space: "default", Local: "foo"}},
				whitespace: []string{""},
			}}, "<foo></foo>",
		},
		{
			"prefix", &struct {
				XMLElement c14n `xmlx:"foo bar,c14nroot"`
			}{c14n{
				ns: make(map[string]string),
				start: startElement{
					nsprefix: "foo",
					name:     xml.Name{Space: "foo", Local: "bar"},
				},
				whitespace: []string{""},
			}}, "<foo:bar></foo:bar>",
		},
		{
			"internal whitespace", &struct {
				XMLElement c14n `xmlx:"foo bar,c14nroot"`
			}{c14n{
				ns: make(map[string]string),
				start: startElement{
					nsprefix: "foo",
					name:     xml.Name{Space: "foo", Local: "bar"},
				},
				whitespace: []string{"\n"},
			}}, "<foo:bar>\n</foo:bar>",
		},
		{
			"root namespace", &struct {
				XMLElement c14n `xmlx:"namespace bar,c14nroot"`
			}{c14n{
				ns: map[string]string{
					"foo": "namespace",
					"":    "default",
				},
				start: startElement{
					nsprefix: "foo",
					name:     xml.Name{Space: "namespace", Local: "bar"},
				},
				whitespace: []string{""},
			}}, `<foo:bar xmlns="default" xmlns:foo="namespace"></foo:bar>`,
		},
		{
			"attributes", &struct {
				XMLElement c14n `xmlx:"default foo,c14nroot"`
			}{c14n{
				ns: map[string]string{
					"":  "default",
					"a": "ns2",
					"b": "ns1",
				},
				start: startElement{
					name: xml.Name{Space: "default", Local: "foo"},
					attr: []attr{
						{
							nsprefix: "a",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "ns2",
									Local: "attr",
								},
								Value: "fifth\t\n\r\"&<>",
							},
						},
						{
							nsprefix: "b",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "ns1",
									Local: "attr2",
								},
								Value: "fourth",
							},
						},
						{
							nsprefix: "b",
							Attr: xml.Attr{
								Name: xml.Name{
									Space: "ns1",
									Local: "attr",
								},
								Value: "third",
							},
						},
						{
							Attr: xml.Attr{
								Name:  xml.Name{Local: "attr"},
								Value: "first",
							},
						},
						{
							Attr: xml.Attr{
								Name:  xml.Name{Local: "attr2"},
								Value: "second",
							},
						},
					},
				},
				whitespace: []string{""},
			}}, `<foo xmlns="default" xmlns:a="ns2" xmlns:b="ns1" ` +
				`attr="first" attr2="second" ` +
				`b:attr="third" b:attr2="fourth" ` +
				`a:attr="fifth&#x9;&#xA;&#xD;&quot;&amp;&lt;>"></foo>`,
		},
		{
			"character data", &struct {
				XMLElement c14n   `xmlx:"default foo,c14nroot"`
				Value      string `xmlx:",chardata"`
			}{
				XMLElement: c14n{
					ns: make(map[string]string),
					start: startElement{name: xml.Name{
						Space: "default",
						Local: "foo",
					}},
				},
				Value: " character data\t\n\r\"&<> ",
			}, "<foo> character data\t\n&#xD;\"&amp;&lt;&gt; </foo>",
		},
		{
			"sub-elements", &struct {
				XMLElement c14n `xmlx:"default foo,c14nroot"`
				Struct     struct {
					XMLElement c14n `xmlx:"default struct"`
				}
				Slice []struct {
					XMLElement c14n   `xmlx:"default slice"`
					Value      string `xmlx:",chardata"`
				}
				OptionalMissing struct {
					XMLElement c14n `xmlx:"default optional-m"`
				} `xmlx:",optional"`
				OptionalPresent struct {
					XMLElement c14n `xmlx:"default optional-p"`
				} `xmlx:",optional"`
			}{
				XMLElement: c14n{
					ns: make(map[string]string),
					start: startElement{name: xml.Name{
						Space: "default",
						Local: "foo",
					}},
					whitespace: []string{
						"\n\t\t\t\t",
						"\n\t\t\t\t",
						"\n\t\t\t\t",
						"\n\t\t\t\t",
						"\n\t\t\t",
					},
				},
				Struct: struct {
					XMLElement c14n `xmlx:"default struct"`
				}{c14n{
					ns: make(map[string]string),
					start: startElement{name: xml.Name{
						Space: "default",
						Local: "struct",
					}},
					whitespace: []string{""},
				}},
				Slice: []struct {
					XMLElement c14n   `xmlx:"default slice"`
					Value      string `xmlx:",chardata"`
				}{
					{
						XMLElement: c14n{
							ns: make(map[string]string),
							start: startElement{name: xml.Name{
								Space: "default",
								Local: "slice",
							}},
						},
						Value: "first",
					},
					{
						XMLElement: c14n{
							ns: make(map[string]string),
							start: startElement{name: xml.Name{
								Space: "default",
								Local: "slice",
							}},
						},
						Value: "second",
					},
				},
				OptionalPresent: struct {
					XMLElement c14n `xmlx:"default optional-p"`
				}{c14n{
					ns: make(map[string]string),
					start: startElement{name: xml.Name{
						Space: "default",
						Local: "optional-p",
					}},
					whitespace: []string{""},
				}},
			}, `<foo>
				<struct></struct>
				<slice>first</slice>
				<slice>second</slice>
				<optional-p></optional-p>
			</foo>`,
		},
		{
			"superfluous namespaces", &struct {
				XMLElement   c14n `xmlx:"default root,c14nroot"`
				OverwriteFoo struct {
					XMLElement c14n `xmlx:"default overwrite-foo"`
					RestoreFoo struct {
						XMLElement c14n `xmlx:"default restore-foo"`
					}
					OverwriteBar struct {
						XMLElement c14n `xmlx:"default overwrite-bar"`
					}
				}
			}{
				XMLElement: c14n{
					ns: map[string]string{
						"foo": "root-foo",
						"bar": "root-bar",
					},
					start: startElement{name: xml.Name{
						Space: "default",
						Local: "root",
					}},
					whitespace: []string{
						"\n\t\t\t\t",
						"\n\t\t\t",
					},
				},
				OverwriteFoo: struct {
					XMLElement c14n `xmlx:"default overwrite-foo"`
					RestoreFoo struct {
						XMLElement c14n `xmlx:"default restore-foo"`
					}
					OverwriteBar struct {
						XMLElement c14n `xmlx:"default overwrite-bar"`
					}
				}{
					XMLElement: c14n{
						start: startElement{
							name: xml.Name{
								Space: "default",
								Local: "overwrite-foo",
							},
							attr: []attr{
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "foo",
										},
										Value: "overwritten-foo",
									},
								},
							},
						},
						whitespace: []string{
							"\n\t\t\t\t\t",
							"\n\t\t\t\t\t",
							"\n\t\t\t\t",
						},
					},
					RestoreFoo: struct {
						XMLElement c14n `xmlx:"default restore-foo"`
					}{c14n{
						start: startElement{
							name: xml.Name{
								Space: "default",
								Local: "restore-foo",
							},
							attr: []attr{
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "foo",
										},
										Value: "root-foo",
									},
								},
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "baz",
										},
										Value: "new-baz",
									},
								},
							},
						},
						whitespace: []string{"\n\t\t\t\t\t"},
					}},
					OverwriteBar: struct {
						XMLElement c14n `xmlx:"default overwrite-bar"`
					}{c14n{
						ns: make(map[string]string),
						start: startElement{
							name: xml.Name{
								Space: "default",
								Local: "overwrite-bar",
							},
							attr: []attr{
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "foo",
										},
										Value: "overwritten-foo",
									},
								},
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "bar",
										},
										Value: "overwritten-bar",
									},
								},
								{
									nsprefix: xmlns,
									Attr: xml.Attr{
										Name: xml.Name{
											Local: "baz",
										},
										Value: "",
									},
								},
							},
						},
						whitespace: []string{"\n\t\t\t\t\t"},
					}},
				},
			}, `<root xmlns:bar="root-bar" xmlns:foo="root-foo">
				<overwrite-foo xmlns:foo="overwritten-foo">
					<restore-foo xmlns:baz="new-baz" xmlns:foo="root-foo">
					</restore-foo>
					<overwrite-bar xmlns:bar="overwritten-bar" xmlns:baz="">
					</overwrite-bar>
				</overwrite-foo>
			</root>`,
		},
	}

	var buf bytes.Buffer
	for _, test := range tests {
		buf.Reset()
		t.Run(test.name, func(t *testing.T) {
			writeXML(test.element, &buf)
			if xml := buf.String(); xml != test.expected {
				t.Errorf("unexpected result: %q; want %q", xml, test.expected)
			}
		})
	}
}

func TestWriteXMLSignedInfo(t *testing.T) {
	var buf bytes.Buffer
	for _, test := range []string{"EID", "MID", "Compact"} {
		buf.Reset()
		t.Run(test, func(t *testing.T) {
			prec14n := fmt.Sprint("signatures0", test, ".xml")
			xml, err := ioutil.ReadFile(filepath.Join("testdata", prec14n))
			if err != nil {
				t.Fatal(err)
			}
			postc14n := fmt.Sprint("canonicalSignedInfo", test)
			expected, err := ioutil.ReadFile(filepath.Join("testdata", postc14n))
			if err != nil {
				t.Fatal(err)
			}

			var signatures xadesSignatures
			if err := parseXML(xml, &signatures); err != nil {
				t.Fatal("unexpected error parsing signatures:", err)
			}
			writeXML(&signatures.Signature.SignedInfo, &buf)

			if !bytes.Equal(buf.Bytes(), expected) {
				t.Error("canonicalized ", prec14n, " does not match ", postc14n)
			}
		})
	}
}
