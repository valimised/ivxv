package cms

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
)

// https://www.rfc-editor.org/rfc/rfc5652#section-3
// ContentInfo
//     -- contentType is id-signedData
//     -- content is SignedData
type ContentInfo struct {
	Raw         asn1.RawContent
	ContentType asn1.ObjectIdentifier
	Content     SignedData `asn1:"explicit,tag:0"`
}

// https://tools.ietf.org/html/rfc5652#section-5.1
type SignedData struct {
	Version          int
	DigestAlgorithms []pkix.AlgorithmIdentifier `asn1:"set"`
	EncapContentInfo EncapsulatedContentInfo
	Certificates     []CertificateChoices   `asn1:"set,tag:0,optional"`
	Crls             []pkix.CertificateList `asn1:"set,tag:1,optional"`
	SignerInfos      []SignerInfo           `asn1:"set"`
}

//https://tools.ietf.org/html/rfc5652#section-5.2
type EncapsulatedContentInfo struct {
	RawContent   asn1.RawContent
	EContentType asn1.ObjectIdentifier
	EContent     []byte `asn1:"explicit,tag:0,optional"`
}

// https://tools.ietf.org/html/rfc5652#section-10.2.2
type CertificateChoices struct {
	RawContent asn1.RawContent
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type SignerInfo struct {
	Version int
	// Golang doesn't support ASN.1 CHOICE, so make 2 optional fields
	IssuerAndSerialNumber IssuerAndSerialNumber `asn1:"optional"`
	SubjectKeyIdentifier  []byte                `asn1:"tag:0,optional"`
	DigestAlgorithm       pkix.AlgorithmIdentifier
	SignedAttrs           []Attribute `asn1:"set,tag:0,optional"`
	SignatureAlgorithm    pkix.AlgorithmIdentifier
	Signature             []byte
	UnsignedAttrs         []Attribute `asn1:"set,tag:1,optional"`
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type Attribute struct {
	AttrType  asn1.ObjectIdentifier
	AttrValue asn1.RawValue
}

// AttrToVerify represents an additional attribute
// that needs verification. It can also act as a replacement
// for the default verification callback of the given attribute, f.e.,
// when verifying the signing time of timestamp responses.
type AttrToVerify struct {
	// AttributeID is the string representation
	// of the object identifier of the attribute
	AttributeID string

	// VerifyCallback encapsulates the call to
	// a function that verifies the attribute value
	// when being parsed.
	VerifyCallback func([]byte) error
}

// https://tools.ietf.org/html/rfc5652#section-10.2.4
type IssuerAndSerialNumber struct {
	Issuer       pkix.RDNSequence
	SerialNumber *big.Int
}

// https://tools.ietf.org/html/rfc5035#section-3
type SigningCertificateV2 struct {
	Certs    []ESSCertIDv2
	Policies asn1.RawValue `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc5035#section-4
type ESSCertIDv2 struct {
	HashAlgorithm         pkix.AlgorithmIdentifier `asn1:"optional"`
	CertHash              []byte
	IssuerAndSerialNumber IssuerSerial `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc5035#section-4
type IssuerSerial struct {
	Issuer       GeneralName
	SerialNumber *big.Int
}

// https://tools.ietf.org/html/rfc5280#page-38
type GeneralName struct {
	DirectoryName pkix.RDNSequence `asn1:"explicit,tag:4"`
}

// https://tools.ietf.org/html/rfc6211#section-2
type CMSAlgorithmProtection struct {
	DigestAlgorithm    pkix.AlgorithmIdentifier
	SignatureAlgorithm pkix.AlgorithmIdentifier `asn1:"tag:1"`
}
