package ocsp

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"
)

// Conf contains the configurable options for the OCSP api.
type Conf struct {
	// URL is the address of the OCSP responder.
	URL string

	// Retries is the number of times a OCSP request is retried in case of failure.
	// Defaults to 2 retries.
	Retries uint64
}

// REQUEST

// https://tools.ietf.org/html/rfc6960#section-4.1.1
type ocspRequest struct {
	TBSRequest tbsRequest
}

type tbsRequest struct {
	Version           int `asn1:"explicit,tag:0,optional,default:0"`
	RequestList       []request
	RequestExtensions []pkix.Extension `asn1:"explicit,tag:2,optional"`
}

type request struct {
	ReqCert certID
}

type certID struct {
	HashAlgorithm  pkix.AlgorithmIdentifier
	IssuerNameHash []byte
	IssuerKeyHash  []byte
	SerialNumber   *big.Int
}

// RESPONSE

// CertStatus is the return value for OCSP commands containing all the
// relevant information about the OCSP response.
type CertStatus struct {
	ProducedAt       time.Time // The time this response was produced at
	Nonce            []byte    // The nonce used in the response
	Good             bool      // If the status of the requested certificate is good
	Unknown          bool      // If the status of the certificate is unknown
	RevocationReason int       // If the certificate is revoked, the reason of revocation.
}

// LiveCertStatus is an extension of CertStatus returned from live OCSP queries
// which also contains the raw basic response received from the server.
type LiveCertStatus struct {
	RawResponse []byte // The raw ASN.1 DER-encoded basic response from the server.
	CertStatus
}

// https://tools.ietf.org/html/rfc6960#section-4.2.1
type ocspResponse struct {
	ResponseStatus asn1.Enumerated
	ResponseBytes  responseBytes `asn1:"explicit,tag:0,optional"`
}

const ocspResponseStatusSuccessful = 0

var ocspResponseStatus = map[int]string{
	0: "successful",
	1: "malformedRequest",
	2: "internalError",
	3: "tryLater",
	// 4: unused,
	5: "sigRequired",
	6: "unauthorized",
}

type responseBytes struct {
	ResponseType asn1.ObjectIdentifier
	Response     []byte
}

type basicOCSPResponse struct {
	Raw                asn1.RawContent
	TBSResponseData    responseData
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Signature          asn1.BitString
	Certs              []asn1.RawValue `asn1:"explicit,tag:0,optional"`
}

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

type singleResponse struct {
	CertID certID

	// The asn1 module does not support choices, so use 3 optional fields
	// instead of CertStatus.
	CertStatusGood        asn1.Flag   `asn1:"explicit,tag:0,optional"`
	CertStatusRevokedTime time.Time   `asn1:"explicit,tag:1,optional"`
	CertStatusRevokedInfo revokedInfo `asn1:"explicit,tag:1,optional"`
	CertStatusUnknown     asn1.Flag   `asn1:"explicit,tag:2,optional"`

	ThisUpdate       time.Time
	NextUpdate       time.Time        `asn1:"explicit,tag:0,optional"`
	SingleExtensions []pkix.Extension `asn1:"explicit,tag:1,optional"`
}

// Alternative type for CertStatusRevoked in singleResponse
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
