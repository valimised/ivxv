package cryptoutil

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
)

func TestRDNSequenceEqual(t *testing.T) {
	tests := []struct {
		name  string
		a, b  pkix.RDNSequence
		equal bool
	}{
		{"nil", nil, nil, true},
		{"empty", pkix.RDNSequence{}, pkix.RDNSequence{}, true},
		{"empty nil", pkix.RDNSequence{}, nil, true},
		{"non-empty nil", pkix.RDNSequence{{}}, nil, false},

		{"equal", pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, true},

		{"types", pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2},
				Value: "foobar",
			}},
		}, pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2, 3},
				Value: "foobar",
			}},
		}, false},

		{"values", pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2, 3},
				Value: "foobar",
			}},
		}, pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2, 3},
				Value: "not foobar",
			}},
		}, false},

		{"not comparable", pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2, 3},
				Value: []byte{0xde, 0xad, 0xbe, 0xef},
			}},
		}, pkix.RDNSequence{
			{{
				Type:  asn1.ObjectIdentifier{1, 2, 3},
				Value: []byte{0xde, 0xad, 0xbe, 0xef},
			}},
		}, false},

		{"reordered sequence", pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}, {{
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}}}, pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}}, {{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, false},

		{"reordered set", pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}, {
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}}}, pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}, {
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, true},

		{"missing RDN", pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}, {{
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}}}, pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, false},

		{"missing attribute", pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}, {
			Type:  asn1.ObjectIdentifier{4, 5, 6},
			Value: 123,
		}}}, pkix.RDNSequence{{{
			Type:  asn1.ObjectIdentifier{1, 2, 3},
			Value: "foobar",
		}}}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if RDNSequenceEqual(test.a, test.b) != test.equal {
				t.Errorf("assertion failed")
			}
		})
	}
}

func TestEncodeDecodeRDNSequence(t *testing.T) {
	tests := []struct {
		name    string
		decoded pkix.RDNSequence
		encoded string
	}{
		{"nil", nil, ""},
		{"empty", pkix.RDNSequence{}, ""},
		{"single", pkix.Name{CommonName: "foobar"}.ToRDNSequence(), "CN=foobar"},

		{"multiple", pkix.Name{
			CommonName:   "foobar",
			SerialNumber: "1234",
		}.ToRDNSequence(), "serialNumber=1234,CN=foobar"},

		{"multi-value", pkix.RDNSequence{{
			{Type: asn1.ObjectIdentifier{2, 5, 4, 42}, Value: "foo"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 4}, Value: "bar"},
		}}, "GN=foo+SN=bar"},

		{"escape", pkix.Name{CommonName: "#\"+,;<=>\\\x00 "}.ToRDNSequence(),
			`CN=\#\"\+\,\;\<\=\>\\\00\ `},

		{"spaces", pkix.Name{CommonName: "   "}.ToRDNSequence(), `CN=\  \ `},

		{"unknown", pkix.RDNSequence{{
			{Type: asn1.ObjectIdentifier{1, 2, 3}, Value: "foobar"},
		}}, "1.2.3=#1306666f6f626172"},

		{"non-string", pkix.RDNSequence{{
			{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: int64(1)},
		}}, "CN=#020101"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := EncodeRDNSequence(test.decoded)
			if err != nil {
				t.Fatal("failed to encode DN:", err)
			}
			if encoded != test.encoded {
				t.Errorf("unexpected encoding: got %q, want %q", encoded, test.encoded)
			}

			decoded, err := DecodeRDNSequence(test.encoded)
			if err != nil {
				t.Fatal("failed to decode DN:", err)
			}
			if !RDNSequenceEqual(decoded, test.decoded) {
				t.Errorf("unexpected decoding: got %v, want %v", decoded, test.decoded)
			}
		})
	}
}
