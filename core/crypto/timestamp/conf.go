package timestamp

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"
)

// Conf contains the configurable options for the TS api.
type Conf struct {
	// URL is the address of the service to which the TS requests are sent.
	URL string

	// Signers are the x509 certificates used by the server to sign the token.
	Signers []*x509.Certificate

	// DelayTime is the maximum time that GenTime and SignTime can differ.
	// Defaults to 1 second.
	DelayTime uint64

	// Retries is the number of times a TS request is retried in case of failure.
	// Defaults to 2 retries.
	Retries uint64
}

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

// https://tools.ietf.org/html/rfc3161#section-2.4.2
type tsResponse struct {
	PKIStatus      pkiStatusInfo
	TimeStampToken timestampToken `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc3161#section-2.4.2
type pkiStatusInfo struct {
	Status       int
	StatusString []string       `asn1:"optional"`
	FailInfo     asn1.BitString `asn1:"optional"`
}

// https://tools.ietf.org/html/rfc3161#section-2.4.2
// TimestampToken ::= ContentInfo
//     -- contentType is id-signedData
//     -- content is SignedData
type timestampToken struct {
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

//https://tools.ietf.org/html/rfc5652#section-5.2
type encapsulatedContentInfo struct {
	RawContent   asn1.RawContent
	EContentType asn1.ObjectIdentifier
	EContent     []byte `asn1:"explicit,tag:0,optional"`
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

// https://tools.ietf.org/html/rfc5652#section-10.2.2
type certificateChoices struct {
	RawContent asn1.RawContent
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type signerInfo struct {
	Version int
	// Golang doesn't support ASN.1 CHOICE, so make 2 optional fields
	IssuerAndSerialNumber issuerAndSerialNumber `asn1:"optional"`
	SubjectKeyIdentifier  []byte                `asn1:"tag:0,optional"`
	DigestAlgorithm       pkix.AlgorithmIdentifier
	SignedAttrs           []attribute `asn1:"set,tag:0,optional"`
	SignatureAlgorithm    pkix.AlgorithmIdentifier
	Signature             []byte
	UnsignedAttrs         []attribute `asn1:"set,tag:1,optional"`
}

// https://tools.ietf.org/html/rfc5652#section-5.3
type attribute struct {
	AttrType  asn1.ObjectIdentifier
	AttrValue asn1.RawValue
}

// https://tools.ietf.org/html/rfc5652#section-10.2.4
type issuerAndSerialNumber struct {
	Issuer       pkix.RDNSequence
	SerialNumber *big.Int
}
