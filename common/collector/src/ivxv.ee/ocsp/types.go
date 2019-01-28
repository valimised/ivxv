package ocsp // import "ivxv.ee/ocsp"

import (
	"bytes"
	"crypto/sha1" // nolint: gosec, See certIDHash.
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"

	"ivxv.ee/cryptoutil"
)

var (
	// Hash function used to calculate CertID fields.
	//
	// We would like to use a stronger hash function here, but at the
	// moment of writing, the SK OCSP responders used by Estonian IVXV
	// deployments do not support anything else (and give very unexpected
	// responses when used).
	certIDHash = sha1.Sum

	// OID for certIDHash.
	certIDHashOID = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
)

// REQUEST

// https://tools.ietf.org/html/rfc6960#section-4.1.1
type ocspRequest struct {
	TBSRequest tbsRequest
}

// https://tools.ietf.org/html/rfc6960#section-4.1.1
type tbsRequest struct {
	Version           int `asn1:"explicit,tag:0,optional,default:0"`
	RequestList       []request
	RequestExtensions []pkix.Extension `asn1:"explicit,tag:2,optional"`
}

// https://tools.ietf.org/html/rfc6960#section-4.1.1
type request struct {
	ReqCert certID
}

// https://tools.ietf.org/html/rfc6960#section-4.1.1
type certID struct {
	HashAlgorithm  pkix.AlgorithmIdentifier
	IssuerNameHash []byte
	IssuerKeyHash  []byte
	SerialNumber   *big.Int
}

func newCertID(cert *x509.Certificate) (id *certID, err error) {
	// Since certIDHash is SHA-1 then skip calculating the IssuerKeyHash
	// and simply use AuthorityKeyId.
	if len(cert.AuthorityKeyId) == 0 {
		return nil, AuthKeyIDMissingError{}
	}

	nameHash := certIDHash(cert.RawIssuer)
	return &certID{
		HashAlgorithm: pkix.AlgorithmIdentifier{
			Algorithm: certIDHashOID,
		},
		IssuerNameHash: nameHash[:],
		IssuerKeyHash:  cert.AuthorityKeyId,
		SerialNumber:   cert.SerialNumber,
	}, nil
}

func (c *certID) equal(other *certID) bool {
	return cryptoutil.AlgorithmIdentifierCmp(c.HashAlgorithm, other.HashAlgorithm) &&
		bytes.Equal(c.IssuerNameHash, other.IssuerNameHash) &&
		bytes.Equal(c.IssuerKeyHash, other.IssuerKeyHash) &&
		c.SerialNumber.Cmp(other.SerialNumber) == 0
}

// RESPONSE

// https://tools.ietf.org/html/rfc6960#section-4.2.1
type ocspResponse struct {
	ResponseStatus asn1.Enumerated
	ResponseBytes  responseBytes `asn1:"explicit,tag:0,optional"`
}

// https://tools.ietf.org/html/rfc6960#section-4.2.1
const ocspResponseStatusSuccessful = 0

// https://tools.ietf.org/html/rfc6960#section-4.2.1
var ocspResponseStatus = map[int]string{
	0: "successful",
	1: "malformedRequest",
	2: "internalError",
	3: "tryLater",
	// 4: unused,
	5: "sigRequired",
	6: "unauthorized",
}

// https://tools.ietf.org/html/rfc6960#section-4.2.1
type responseBytes struct {
	ResponseType asn1.ObjectIdentifier
	Response     []byte
}

// https://tools.ietf.org/html/rfc6960#page-15
type basicOCSPResponse struct {
	Raw                asn1.RawContent
	TBSResponseData    responseData
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Signature          asn1.BitString
	Certs              []asn1.RawValue `asn1:"explicit,tag:0,optional"`
}

// https://tools.ietf.org/html/rfc6960#page-15
type responseData struct {
	Raw     asn1.RawContent
	Version int `asn1:"explicit,tag:0,optional,default:0"`

	// The asn1 module does not support choices, so use 2 optional fields
	// instead of ResponderID.
	ResponderIDByName pkix.RDNSequence `asn1:"explicit,tag:1,optional"`
	ResponderIDByKey  []byte           `asn1:"explicit,tag:2,optional"`

	ProducedAt         time.Time
	Responses          []singleResponse
	ResponseExtensions []pkix.Extension `asn1:"explicit,tag:1,optional"`
}

// https://tools.ietf.org/html/rfc6960#page-15
type singleResponse struct { // nolint: aligncheck, maligned, preserve RFC ordering.
	CertID certID

	// The asn1 module does not support choices, so use 3 optional fields
	// instead of CertStatus.
	CertStatusGood    asn1.Flag   `asn1:"tag:0,optional"`
	CertStatusRevoked revokedInfo `asn1:"tag:1,optional"`
	CertStatusUnknown asn1.Flag   `asn1:"tag:2,optional"`

	ThisUpdate       time.Time
	NextUpdate       time.Time        `asn1:"explicit,tag:0,optional"`
	SingleExtensions []pkix.Extension `asn1:"explicit,tag:1,optional"`
}

// https://tools.ietf.org/html/rfc6960#page-15
type revokedInfo struct {
	RevocationTime   time.Time
	RevocationReason asn1.Enumerated `asn1:"explicit,tag:0,optional"`
}

// The following constants define the many possible reasons a certificate may
// have been revoked. These constants can be used to check the revocation reason
// returned inside the CertStatus structure
// https://tools.ietf.org/html/rfc5280#section-5.3.1
const (
	ReasonUnspecified = iota
	ReasonKeyCompromise
	ReasonCACompromise
	ReasonAffiliationChanged
	ReasonSuperseded
	ReasonCessationOfOperation
	ReasonCertificateHold
	_
	ReasonRemoveFromCRL
	ReasonPrivilegeWithdrawn
	ReasonAACompromise
)
