package elgamal

import "encoding/asn1"

var (
	ecElGamalEncryptionOID = func() asn1.ObjectIdentifier { return asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1} } //nolint:lll
)
