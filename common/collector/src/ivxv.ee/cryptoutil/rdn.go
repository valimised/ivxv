package cryptoutil

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// RDNSequenceEqual checks if two relative distinguished name sequences contain
// the same names in the same order and if names contain the same attributes,
// ignoring order in case of multiple values.
//
// This does not enforce any restrictions on attributes: it only checks if the
// sequences match even if they contain multiple attributes of the same type,
// values not valid for a type, etc.
//
// Only attribute values which are comparable using the equality operator can
// be compared (https://golang.org/ref/spec#Comparison_operators). So if there
// is an attribute with matching types and values, but the values are not
// comparable, then RDNSequenceEqual will return false.
func RDNSequenceEqual(a, b pkix.RDNSequence) bool {
	if len(a) != len(b) {
		return false
	}

	for i, set := range a {
		if !rdnEqual(set, b[i]) {
			return false
		}
	}
	return true
}

func rdnEqual(a, b pkix.RelativeDistinguishedNameSET) bool {
	if len(a) != len(b) {
		return false
	}

	// Create a copy of b that we can manipulate.
	c := make(pkix.RelativeDistinguishedNameSET, len(b))
	copy(c, b)

	// Comparing non-comparable attribute values panics, so defer recover.
	// We can safely say that a and b are not equal if a panic was caused.
	defer func() { // defer recover() does not work.
		recover() //nolint:errcheck // Discard the panic.
	}()
next:
	for _, aatv := range a {
		for i, catv := range c {
			// FIXME: Comparison of different numeric types, e.g., int and uint.
			if aatv.Type.Equal(catv.Type) && aatv.Value == catv.Value {
				// Delete c[i] so we do not match it twice.
				// Minimize copying by not preserving order.
				c[i] = c[len(c)-1]
				c = c[:len(c)-1]
				continue next
			}
		}
		return false
	}
	return true
}

var (
	// shortnames is a map from short names of relative distinguished name
	// attribute types to their corresponding object identifiers.
	//
	// Keep these sorted by OID for easier tracking.
	shortnames = map[string]asn1.ObjectIdentifier{
		"emailAddress":           {1, 2, 840, 113549, 1, 9, 1},
		"CN":                     {2, 5, 4, 3},
		"SN":                     {2, 5, 4, 4},
		"serialNumber":           {2, 5, 4, 5},
		"C":                      {2, 5, 4, 6},
		"L":                      {2, 5, 4, 7},
		"ST":                     {2, 5, 4, 8},
		"O":                      {2, 5, 4, 10},
		"OU":                     {2, 5, 4, 11},
		"GN":                     {2, 5, 4, 42},
		"organizationIdentifier": {2, 5, 4, 97},
	}

	// shortToOID is a version of the shortnames map with lowercase keys
	// necessary for case-insensitive lookup. Filled from shortnames on
	// init.
	shortToOID = make(map[string]asn1.ObjectIdentifier)

	// oidToShort maps relative distinguished name attribute type object
	// identifiers (in string form) to short names. Filled from shortnames
	// on init.
	oidToShort = make(map[string]string)
)

func init() {
	for short, oid := range shortnames {
		shortToOID[strings.ToLower(short)] = oid
		oidToShort[oid.String()] = short
	}
}

// EncodeRDNSequence encodes a sequence of relative distinguished names (i.e.,
// a distinguished name) into a string according to RFC 4514 section 2.
func EncodeRDNSequence(dn pkix.RDNSequence) (string, error) {
	if len(dn) == 0 {
		return "", nil
	}
	var buf bytes.Buffer
	for i := len(dn) - 1; i >= 0; i-- {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		for i, atv := range dn[i] {
			if i > 0 {
				buf.WriteByte('+')
			}
			if err := encodeATV(&buf, atv); err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}

var bsEscapeRE = regexp.MustCompile(`(^#|^ |["+,;<=>\\]| $)`)

func encodeATV(buf *bytes.Buffer, atv pkix.AttributeTypeAndValue) error {
	at := atv.Type.String()
	short, haveShort := oidToShort[at]
	if haveShort {
		at = short
	}
	buf.WriteString(at)

	buf.WriteByte('=')

	s, ok := atv.Value.(string)
	if !haveShort || !ok {
		// Use hexadecimal encoding of the DER encoding of the value.
		buf.WriteByte('#')
		der, err := asn1.Marshal(atv.Value)
		if err != nil {
			return EncodeAttributeValueError{Type: at, Err: err}
		}
		buf.WriteString(hex.EncodeToString(der))
		return nil
	}
	s = bsEscapeRE.ReplaceAllString(s, `\$1`)
	s = strings.ReplaceAll(s, "\x00", "\\00")
	buf.WriteString(s)
	return nil
}

// DecodeRDNSequence decodes a string into a sequence of relative distinguished
// names (i.e., a distinguished name) according to RFC 4514 Section 3.
func DecodeRDNSequence(encoded string) (pkix.RDNSequence, error) {
	var dn pkix.RDNSequence
	for len(encoded) > 0 {
		if len(dn) > 0 {
			if !strings.HasPrefix(encoded, ",") {
				return nil, MissingRDNSeparatorError{
					Encoded: encoded,
				}
			}
			encoded = encoded[1:]
			if len(encoded) == 0 {
				return nil, EmptyRDNError{}
			}
		}

		var rdn pkix.RelativeDistinguishedNameSET
		for len(encoded) > 0 {
			if len(rdn) > 0 {
				if !strings.HasPrefix(encoded, "+") {
					break
				}
				encoded = encoded[1:]
				if len(encoded) == 0 {
					return nil, EmptyAttributeError{}
				}
			}

			atv, rest, err := decodeATV(encoded)
			if err != nil {
				return nil, err
			}
			rdn = append(rdn, atv)
			encoded = rest
		}

		// Insert the new RDN as first.
		dn = append(dn, nil)
		copy(dn[1:], dn)
		dn[0] = rdn
	}
	return dn, nil
}

func decodeATV(encoded string) (atv pkix.AttributeTypeAndValue, rest string, err error) {
	// Find the = that succeeds the type.
	eq := strings.IndexByte(encoded, '=')
	if eq < 0 {
		return atv, "", DecodeATVMissingEqualsError{Encoded: encoded}
	}

	// Get the type OID. Short names are case-insensitive
	atype := encoded[:eq]
	var ok bool
	if atv.Type, ok = shortToOID[strings.ToLower(atype)]; !ok {
		// Attempt to decode it as a raw OID.
		for _, s := range strings.Split(atype, ".") {
			var n int
			n, err = strconv.Atoi(s)
			if err != nil {
				return atv, "", UnknownAttributeTypeError{Type: atype}
			}
			atv.Type = append(atv.Type, n)
		}
		if len(atv.Type) < 2 {
			return atv, "", AttributeTypeOIDTooShortError{OID: atv.Type}
		}
	}

	// If the value starts with #, then it is a hexstring.
	encoded = encoded[eq+1:]
	if strings.HasPrefix(encoded, "#") {
		rest, err = decodeHexString(encoded[1:], &atv.Value)
		return
	}

	// A hexstring must be used if the attribute type was a
	// numericoid, but we allow it to be a string anyway. This
	// should be OK.

	// Otherwise it is an escaped UTF-8 string.
	atv.Value, rest, err = decodeEscaped(encoded)
	return
}

func decodeHexString(encoded string, value *interface{}) (rest string, err error) {
	// Find the comma or plus which terminates the hexstring.
	end := strings.IndexAny(encoded, ",+")
	if end < 0 {
		end = len(encoded) // Last value of the DN.
	}
	encoded, rest = encoded[:end], encoded[end:]

	// Decode the hexadecimal string.
	if len(encoded) == 0 {
		return "", EmptyAttributeValueHexStringError{}
	}
	der, err := hex.DecodeString(encoded)
	if err != nil {
		return "", AttributeValueHexStringError{
			Value: encoded,
			Err:   err,
		}
	}

	*value = new(interface{})
	trailing, err := asn1.Unmarshal(der, value)
	if err != nil {
		return "", InvalidAttributeValueDERError{
			Value: der,
			Err:   err,
		}
	}
	if len(trailing) > 0 {
		return "", AttributeValueDERTrailingDataError{
			Value:    der,
			Trailing: trailing,
		}
	}
	return
}

func decodeEscaped(encoded string) (value, rest string, err error) {
	var buf bytes.Buffer
	var i int
loop:
	for i < len(encoded) {
		r, size := utf8.DecodeRuneInString(encoded[i:])
		switch r {
		case utf8.RuneError:
			return "", "", AttributeValueInvalidUTF8Error{Encoded: encoded[i:]}

		case '\\': // Escaped byte.
			// Check for special escaped char first.
			if len(encoded[i+size:]) == 0 {
				return "", "", AttributeValueUnescapedTrailingSlashError{
					Encoded: encoded,
				}
			}
			if special := encoded[i+size]; strings.IndexByte(` "#+,;<=>\`, special) >= 0 {
				buf.WriteByte(special)
				i += size + 1
				break
			}

			// Otherwise it must be a hex pair.
			if len(encoded[i+size:]) < 2 {
				return "", "", AttributeValueBadEscapedCharError{
					Escaped: string(encoded[i+size]),
				}
			}
			b, err := hex.DecodeString(encoded[i+size : i+size+2])
			if err != nil {
				return "", "", AttributeValueHexPairError{
					Pair: encoded[i+size : i+size+2],
					Err:  err,
				}
			}
			buf.WriteByte(b[0])
			i += size + 2

		case '+', ',': // End of value.
			break loop

		case 0, '"', ';', '<', '>': // Not allowed.
			return "", "", AttributeValueUnescapedSpecialError{
				Special: string(r),
			}

		case ' ': // Not allowed in leading nor trailing character.
			if i == 0 {
				return "", "", AttributeValueUnescapedLeadingSpaceError{
					Encoded: encoded,
				}
			}
			if tail := encoded[i+size:]; len(tail) == 0 ||
				tail[0] == ',' || tail[0] == '+' {

				return "", "", AttributeValueUnescapedTrailingSpaceError{
					Encoded: encoded,
				}
			}
			fallthrough

		default:
			buf.WriteRune(r)
			i += size
		}
	}
	value = buf.String()
	rest = encoded[i:]
	return
}
