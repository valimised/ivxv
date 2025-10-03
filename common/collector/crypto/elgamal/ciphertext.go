package elgamal

import (
	"encoding/asn1"
	"strings"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	asn_11 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
)

// UnmarshalCiphertext ASN.1 unmarshalls IVXV specific ASN.1 ElGamal ciphertext as
//
//	ciphertext ::= SEQUENCE {
//		algorithm	AlgorithmIdentifier
//		data	IVXV ELGAMAL CIPHERTEXT
//	}
//
//	IVXV ELGAMAL CIPHERTEXT ::= SEQUENCE {
//		A	GROUP ELEMENT
//		B	GROUP ELEMENT
//	}
//
//	GROUP ELEMENT ::= CHOICE {
//		ModPElement	INTEGER
//		ECpElement	OCTET STRING
//		Edwards25519Element	OCTET STRING
//	}
//
// and then performs all ciphertext correctness checks as
// `processor squash --conf conf.bdoc --params params.bdoc` offline application
// would do, i.e.
//
// 1. ElGamal ciphertext is correct ASN.1
//
// 2. OID of ElGamal ciphertext is supported by the group implementation
//
// 3.1. For ModP group checks that:
//
// a) Both ElGamal ciphertext A and B values belong to the group
//
// b) Both ElGamal ciphertext A and B values > 0
//
// c) Both ElGamal ciphertext A and B values' group order > 0
//
// d) Both ElGamal ciphertext A and B values are quadratic residue modulo P
//
// 3.2. For EC group check that:
//
// a) Both ElGamal ciphertext A and B values belong to the group
//
// b) Both ElGamal ciphertext A and B values are valid EC points and are on the curve
func UnmarshalCiphertext(g group.Group, data asn_11.DER) (*elgamal.Ciphertext, error) {
	sequence := cryptobyte.String(data)
	var ciphertext cryptobyte.String
	var algorithm cryptobyte.String
	var oid asn1.ObjectIdentifier
	var ciphertextData cryptobyte.String
	var A, B cryptobyte.String

	if !sequence.ReadASN1(&ciphertext, asn_1.SEQUENCE) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNotSequenceError{
			Description: _CIP_SEQ,
		}
	}

	if !ciphertext.ReadASN1(&algorithm, asn_1.SEQUENCE) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNotAlgorithmIdentifierError{
			Description: _CIP_AID,
		}
	}

	if !algorithm.ReadASN1ObjectIdentifier(&oid) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNotAlgorithmError{
			Description: _CIP_AID_OID,
		}
	}

	// Trusted ElGamal oid
	elGamalOIDs := []asn1.ObjectIdentifier{
		crypto.ElGamalEncryptionOID(),
		ecElGamalEncryptionOID(),
	}

	var found bool
	for _, elGamalOID := range elGamalOIDs {
		if strings.Compare(oid.String(), elGamalOID.String()) == 0 {
			found = true
			break
		}
	}

	if !found {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNotElGamalOIDError{
			OID:         oid.String(),
			Description: _CIP_OID_MISMATCH,
		}
	}

	if !ciphertext.ReadASN1(&ciphertextData, asn_1.SEQUENCE) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierCiphertextDataError{
			Description: _CIP_DATA,
		}
	}

	if !ciphertextData.ReadAnyASN1Element(&A, nil) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNoEphemeralError{
			Description: _CIP_A,
		}
	}

	if !ciphertextData.ReadAnyASN1Element(&B, nil) {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierNoBlindedMsgError{
			Description: _CIP_B,
		}
	}

	// Is there left any unread bytes?
	if !sequence.Empty() || !ciphertext.Empty() || !algorithm.Empty() || !ciphertextData.Empty() {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierTrailingBytesError{
			Description: _CIP_TRAIL,
		}
	}

	// Prepare ElGamal ASN.1 ciphertext for crypto library supported format
	var builder cryptobyte.Builder
	builder.AddASN1(asn_1.SEQUENCE, func(builder *cryptobyte.Builder) {
		builder.AddBytes(A)
		builder.AddBytes(B)
	})

	der, err := builder.Bytes()
	if err != nil {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierASN1MarshalCiphertextError{Err: err,
			Description: _CIP_MARSHAL}
	}

	ct, err := elgamal.ASN1UnmarshalCiphertext(g, der)
	if err != nil {
		return nil, UnmarshalCiphertextElGamalASN1CiphertextVerifierError{Err: err,
			Description: _CIP_VERIFY,
		}
	}

	return ct, nil
}
