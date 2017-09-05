package bdoc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCanonicalizeUnit(t *testing.T) {
	nameSpaces := []xmlns{
		{name: "test", url: "uri"},
	}

	tests := []struct {
		tag      string
		data     string
		expected string
	}{
		{
			"foo",
			"<foo\t\n\r\n/>",
			`<foo xmlns:test="uri"></foo>`,
		}, {
			"foo",
			"<foo\t\n\r\n></foo>",
			`<foo xmlns:test="uri"></foo>`,
		}, {
			"foo",
			"<foo attr1=\"test\"\t\nattr2=\"test\" />",
			`<foo xmlns:test="uri" attr1="test" attr2="test"></foo>`,
		}, {
			"foo",
			"<foo attr1=\"test\"\t\nattr2=\"test\" ></foo>",
			`<foo xmlns:test="uri" attr1="test" attr2="test"></foo>`,
		}, {
			"foo:bar",
			`<foo:bar attr="te st"  attr2="test2"  />`,
			`<foo:bar xmlns:test="uri" attr="te st" attr2="test2"></foo:bar>`,
		}, {
			"foo:bar",
			`<foo:bar attr="te st"  attr2="test2"  ></foo:bar>`,
			`<foo:bar xmlns:test="uri" attr="te st" attr2="test2"></foo:bar>`,
		}, {
			"foo",
			`<foo attr="test"></foo><bar/>`,
			`<foo xmlns:test="uri" attr="test"></foo>`,
		},
	}

	for i, test := range tests {
		nr := strconv.Itoa(i)
		t.Run("UnitTest#"+nr, func(t *testing.T) {
			result, err := canonicalize(test.data, test.tag, nameSpaces)
			if err != nil {
				t.Fatalf("Failed to canonicalize string: %v", err)
			}
			if string(result) != test.expected {
				t.Errorf("Canonicalization mismatch:\nExpected: %s\nGot: %s\n",
					test.expected, string(result))
			}
		})
	}
}

func TestCanonicalizeFull(t *testing.T) {
	nameSpaces := []xmlns{
		{name: "asic", url: "http://uri.etsi.org/02918/v1.2.1#"},
		{name: "ds", url: "http://www.w3.org/2000/09/xmldsig#"},
		{name: "xades", url: "http://uri.etsi.org/01903/v1.3.2#"},
	}

	tests := []string{"EID", "MID", "Compact"}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			fileName := fmt.Sprintf("%s%s%s", "signatures0", test, ".xml")
			xml, err := ioutil.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			result, err := canonicalize(string(xml), "ds:SignedInfo", nameSpaces)
			if err != nil {
				t.Fatalf("Failed to canonicalize ds:SignedInfo: %v", err)
			}
			fileName = fmt.Sprintf("%s%s", "canonicalSignedInfo", test)
			expected, err := ioutil.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatalf("Failed to read expected canonicalised ds:SignedInfo: %v", err)
			}
			if !bytes.Equal(result, expected) {
				t.Error("Failed to canonicalize SignedInfo")
			}
		})
	}
}
