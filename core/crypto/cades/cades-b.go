package cades

import (
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"time"

	"tivi.io/core/crypto/cms"
)

const idSigningTime = "1.2.840.113549.1.9.5"

var signedAttrs = map[string]cms.AttrToVerify{
	idSigningTime: {
		AttributeID: idSigningTime,
		// The section 11.3 of the RFC 5652 - Cryptographic Message Syntax
		// https://www.rfc-editor.org/rfc/rfc5652#section-11.3
		// specifies the details of the Signing Time attribute, in terms of
		// its possible encoding formats or values (UTC or Generalized Time).
		// The golang library asn1.go already parses it according to the
		// detected type and related constraints. So, everything is already
		// handled in terms of necessary verifications, just by calling the
		// asn1.Unmarshal function.
		VerifyCallback: func(b []byte) error {
			var signingTime time.Time
			rest, err := asn1.Unmarshal(b, &signingTime)
			if err != nil {
				return fmt.Errorf("unmarshal signing time: %w", err)
			}
			if len(rest) > 0 {
				return fmt.Errorf("unmarshal signing time excess bytes: %d", len(rest))
			}
			return nil
		},
	},
}

// VerifySignedData verifies the signed data according to the CAdES-b level,
// matching it to the signers' certificates provided.
// Returns the data that was signed.
func CAdESBVerifySignedData(data []byte, certs []*x509.Certificate, detachedData []byte) error {
	err := cms.VerifySignedData(data, certs, detachedData, signedAttrs, map[string]cms.AttrToVerify{})
	if err != nil {
		return fmt.Errorf("cms verify signed data: %w", err)
	}
	return nil
}

func cadesBVerifySignedData(data []byte, certs []*x509.Certificate, detachedData []byte, otherSignedAttrs, otherUnsignedAttrs map[string]cms.AttrToVerify) error {
	for k, v := range otherSignedAttrs {
		signedAttrs[k] = v
	}
	err := cms.VerifySignedData(data, certs, detachedData, signedAttrs, otherUnsignedAttrs)
	if err != nil {
		return fmt.Errorf("cms verify signed data: %w", err)
	}
	return nil
}
