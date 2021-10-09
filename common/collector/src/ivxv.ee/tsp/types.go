package tsp

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"
)

// REQUEST

// https://tools.ietf.org/html/rfc3161#section-2.4.1
// Leave out optional fields we don't use
type tsRequest struct {
	Version        int `asn1:"default:1"`
	MessageImprint messageImprint
	Nonce          *big.Int `asn1:"optional"`
	CertReq        bool
}

// https://tools.ietf.org/html/rfc3161#section-2.4.1
type messageImprint struct {
	HashAlgorithm pkix.AlgorithmIdentifier
	HashedMessage []byte
}

// RESPONSE

// https://tools.ietf.org/html/rfc3161#section-2.4.2
type tsResponse struct {
	Status         pkiStatusInfo
	TimeStampToken timeStampToken `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc3161#section-2.4.2
type pkiStatusInfo struct {
	Status       int
	StatusString []string       `asn1:"optional"`
	FailInfo     asn1.BitString `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc3161#section-2.4.2
// TimeStampToken ::= ContentInfo
//     -- contentType is id-signedData
//     -- content is SignedData
type timeStampToken struct {
	Raw         asn1.RawContent
	ContentType asn1.ObjectIdentifier
	Content     signedData `asn1:"explicit,tag:0"`
}

// https://tools.ietf.org/html/rfc5652#section-5.1
type signedData struct {
	Version          int
	DigestAlgorithms []pkix.AlgorithmIdentifier `asn1:"set"`
	EncapContentInfo encapsulatedContentInfo
	Certificates     []certificateChoices   `asn1:"set,tag:0,optional"`
	Crls             []pkix.CertificateList `asn1:"set,tag:1,optional"`
	SignerInfos      []signerInfo           `asn1:"set"`
}

// https://tools.ietf.org/html/rfc5652#section-10.2.2
type certificateChoices struct {
	RawContent asn1.RawContent
}

//https://tools.ietf.org/html/rfc5652#section-5.2
type encapsulatedContentInfo struct {
	RawContent   asn1.RawContent
	EContentType asn1.ObjectIdentifier
	EContent     []byte `asn1:"explicit,tag:0,optional"`
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type signerInfo struct {
	Version int

	// Golang doesn't support ASN.1 CHOICE, so make 2 optional fields
	IssuerAndSerialNumber issuerAndSerialNumber `asn1:"optional"`
	SubjectKeyIdentifier  []byte                `asn1:"tag:0,optional"`

	DigestAlgorithm    pkix.AlgorithmIdentifier
	SignedAttrs        []signedAttribute `asn1:"set,tag:0,optional"`
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Signature          []byte
	UnsignedAttrs      []pkix.AttributeTypeAndValue `asn1:"set,tag:1,optional"`
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type signedAttribute struct {
	AttrType  asn1.ObjectIdentifier
	AttrValue asn1.RawValue
}

// https://tools.ietf.org/html/rfc5652#section-10.2.4
type issuerAndSerialNumber struct {
	Issuer       pkix.RDNSequence
	SerialNumber *big.Int
}

// https://tools.ietf.org/html/rfc3161#page-8
type tstInfo struct {
	Version        int
	Policy         asn1.ObjectIdentifier
	MessageImprint messageImprint
	SerialNumber   *big.Int
	GenTime        time.Time
	Accuracy       accuracy         `asn1:"optional"`
	Ordering       bool             `asn1:"optional"`
	Nonce          *big.Int         `asn1:"optional"`
	Extensions     []pkix.Extension `asn1:"tag:1,optional"`
}

// https://tools.ietf.org/html/rfc3161#page-10
type accuracy struct {
	Seconds int `asn1:"optional"`
	Millis  int `asn1:"tag:0,optional"`
	Micros  int `asn1:"tag:1,optional"`
}

// https://tools.ietf.org/html/rfc5035#section-3
type signingCertificateV2 struct {
	Certs    []essCertIDv2
	Policies asn1.RawValue `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc5035#section-4
type essCertIDv2 struct {
	HashAlgorithm         pkix.AlgorithmIdentifier `asn1:"optional"`
	CertHash              []byte
	IssuerAndSerialNumber issuerSerial `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc5035#section-4
type issuerSerial struct {
	// Although the RFC says SEQUENCE of generalName we will assume 1 name
	Issuer       generalName
	SerialNumber *big.Int
}

// https://tools.ietf.org/html/rfc5280#page-38
type generalName struct {
	DirectoryName pkix.RDNSequence `asn1:"explicit,tag:4"`
}

// https://tools.ietf.org/html/rfc6211#section-2
type cmsAlgorithmProtection struct {
	DigestAlgorithm    pkix.AlgorithmIdentifier
	SignatureAlgorithm pkix.AlgorithmIdentifier `asn1:"tag:1"`
}
