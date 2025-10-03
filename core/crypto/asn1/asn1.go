package asn1

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
)

// identified is an ASN.1 OCTET STRING with an algorithm identifier that identifies data.
type identified struct {
	Algorithm pkix.AlgorithmIdentifier
	Data      asn1.RawValue
}

// Pack packs multiple byte slices into an ASN.1 structure, so that they can be
// unambiguously distinguished from each other and avoid message-extension
// attacks.
//
// Pack returns ASN.1 OCTET STRING.
func Pack(data ...[]byte) ([]byte, error) {
	return asn1.Marshal(data)
}

// AddIdentifier adds identifier to data, DER marshals it along with data and returns it.
func AddIdentifier(identifier asn1.ObjectIdentifier, data []byte) (der []byte, err error) {
	return asn1.Marshal(identified{
		Algorithm: pkix.AlgorithmIdentifier{Algorithm: identifier},
		Data:      asn1.RawValue{FullBytes: data},
	})
}

// ParseIdentifier returns identifier that was parsed from data and data without
// identifier.
func ParseIdentifier(data []byte) (asn1.ObjectIdentifier, []byte, error) {
	var parsed identified
	rest, err := asn1.Unmarshal(data, &parsed)
	if err != nil {
		return nil, nil, ParseIdentifierASN1UnmarshalError{Err: err}
	}

	if len(rest) > 0 {
		return nil, nil, ParseIdentifierASN1UnmarshalTrailingBytesError{}
	}

	if len(parsed.Algorithm.Parameters.FullBytes) > 0 {
		return nil, nil, ParseIdentifierASN1UnmarshalNoParamsExpectedError{}
	}

	return parsed.Algorithm.Algorithm, parsed.Data.FullBytes, nil
}

// Concat adds all data bytes into a ASN.1 SEQUENCE OF and returns it.
func Concat(data ...[]byte) (concatenated []byte, err error) {
	var sequenceOf []asn1.RawValue
	for _, item := range data {
		if len(item) == 0 {
			continue
		}

		if len(sequenceOf) > 0 && item[0] != sequenceOf[0].FullBytes[0] {
			return nil, ConcatTagMismatch{
				Tag:      fmt.Sprintf("0x%x", item[0]),
				Expected: fmt.Sprintf("0x%x", sequenceOf[0].FullBytes[0]),
			}
		}

		sequenceOf = append(sequenceOf, asn1.RawValue{FullBytes: item})
	}

	sequenceOfBytes, err := asn1.Marshal(sequenceOf)
	if err != nil {
		return nil, ConcatASN1MarshalError{Err: err}
	}

	return sequenceOfBytes, nil
}

// Split splits ASN.1 SEQUENCE OF into individual data packs.
func Split(concatenated []byte) (split [][]byte, err error) {
	// We can't use a slice of RawValues, so take the long way
	var raw asn1.RawValue
	rest, err := asn1.Unmarshal(concatenated, &raw)
	if err != nil {
		return nil, SplitASN1UnmarshalError{Err: err}
	}

	// Trailing bytes?
	if len(rest) > 0 {
		return nil, SplitASN1UnmarshalTrailingBytesError{Err: err}
	}

	sequenceOf := raw.Bytes

	for {
		sequenceOf, err = asn1.Unmarshal(sequenceOf, &raw)
		if err != nil {
			return nil, SplitItemError{
				Index: len(split),
				Err:   err,
			}
		}

		if len(split) > 0 && raw.FullBytes[0] != split[0][0] {
			return nil, SplitTagMismatchError{
				Tag:      fmt.Sprintf("0x%x", raw.FullBytes[0]),
				Expected: fmt.Sprintf("0x%x", split[0][0]),
			}
		}

		split = append(split, raw.FullBytes)

		if len(sequenceOf) == 0 {
			return
		}
	}
}

func HexN(bytes []byte, n int) string {
	if len(bytes) < n {
		n = len(bytes)
	}
	return fmt.Sprintf("%x", bytes[:n])
}
