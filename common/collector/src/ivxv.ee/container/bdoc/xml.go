package bdoc

import (
	"crypto/x509"
	"encoding/xml"
	"time"
)

// Subelements must be explicitly tagged so that their XMLName namespace is
// checked by xml.Unmarshal.

type xadesSignatures struct {
	raw      []byte
	fileName string

	XMLName xml.Name `xml:"http://uri.etsi.org/02918/v1.2.1# XAdESSignatures"`

	// Although the specification requires to support multiple signatures
	// per file, we only support one signature per file.
	Signature signature
}

// https://www.w3.org/TR/xmldsig-core/#sec-Signature
type signature struct {
	XMLName        xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
	ID             string   `xml:"Id,attr"`
	SignedInfo     signedInfo
	SignatureValue signatureValue
	KeyInfo        keyInfo
	Object         object
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignedInfo
type signedInfo struct {
	XMLName                xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# SignedInfo"`
	CanonicalizationMethod canonicalizationMethod
	SignatureMethod        signatureMethod
	Reference              []reference
}

// https://www.w3.org/TR/xmldsig-core/#sec-CanonicalizationMethod
type canonicalizationMethod struct {
	XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# CanonicalizationMethod"`
	Algorithm string   `xml:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignatureMethod
type signatureMethod struct {
	XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# SignatureMethod"`
	Algorithm string   `xml:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-Reference
type reference struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Reference"`
	ID      string   `xml:"Id,attr"`
	Type    string   `xml:",attr"`
	URI     string   `xml:",attr"`

	Transforms   []transform `xml:"Transforms>Transform"`
	DigestMethod digestMethod
	DigestValue  string
}

// https://www.w3.org/TR/xmldsig-core/#sec-Transforms
type transform struct {
	XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Transform"`
	Algorithm string   `xml:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-DigestMethod
type digestMethod struct {
	XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# DigestMethod"`
	Algorithm string   `xml:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignatureValue
type signatureValue struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# SignatureValue"`
	ID      string   `xml:"Id,attr"`
	Value   []byte   `xml:",chardata"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-KeyInfo
type keyInfo struct {
	XMLName  xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# KeyInfo"`
	X509Data x509Data
}

// https://www.w3.org/TR/xmldsig-core/#sec-X509Data
type x509Data struct {
	XMLName         xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# X509Data"`
	X509Certificate string   `xml:"http://www.w3.org/2000/09/xmldsig# X509Certificate"`
	parsedCert      *x509.Certificate
}

// https://www.w3.org/TR/xmldsig-core/#sec-Object
type object struct {
	XMLName              xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Object"`
	QualifyingProperties qualifyingProperties
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2 (page 22)
type qualifyingProperties struct {
	XMLName            xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# QualifyingProperties"`
	Target             string   `xml:",attr"`
	SignedProperties   signedProperties
	UnsignedProperties unsignedProperties
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.1 (page 22)
type signedProperties struct {
	XMLName                    xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SignedProperties"`
	ID                         string   `xml:"Id,attr"`
	SignedSignatureProperties  signedSignatureProperties
	SignedDataObjectProperties signedDataObjectProperties
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.3 (page 23)
type signedSignatureProperties struct {
	XMLName xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SignedSignatureProperties"`

	SigningTime               time.Time
	SigningCertificate        signingCertificate
	SignaturePolicyIdentifier signaturePolicyIdentifier

	// We knowingly ignore the SignatureProductionPlace and SignerRole
	// fields which, while critical, bear no relevance to our use case.
	// We also have no means to validate their correctness.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type signingCertificate struct {
	XMLName xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SigningCertificate"`
	Cert    cert
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type cert struct {
	XMLName      xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# Cert"`
	CertDigest   certDigest
	IssuerSerial issuerSerial
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type certDigest struct {
	XMLName      xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# CertDigest"`
	DigestMethod digestMethod
	DigestValue  string
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type issuerSerial struct {
	XMLName          xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# IssuerSerial"`
	X509IssuerName   string
	X509SerialNumber string
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.3 (page 33)
type signaturePolicyIdentifier struct {
	XMLName           xml.Name          `xml:"http://uri.etsi.org/01903/v1.3.2# SignaturePolicyIdentifier"`
	SignaturePolicyID signaturePolicyID `xml:"SignaturePolicyId"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type signaturePolicyID struct {
	XMLName             xml.Name    `xml:"http://uri.etsi.org/01903/v1.3.2# SignaturePolicyId"`
	SigPolicyID         sigPolicyID `xml:"SigPolicyId"`
	SigPolicyHash       sigPolicyHash
	SigPolicyQualifiers sigPolicyQualifiers
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type sigPolicyID struct {
	XMLName    xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SigPolicyId"`
	Identifier identifier
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type identifier struct {
	XMLName   xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# Identifier"`
	Qualifier string   `xml:",attr"`
	Value     string   `xml:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type sigPolicyHash struct {
	XMLName      xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SigPolicyHash"`
	DigestMethod digestMethod
	DigestValue  string
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type sigPolicyQualifiers struct {
	XMLName            xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SigPolicyQualifiers"`
	SigPolicyQualifier sigPolicyQualifier
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2 (page 32)
type sigPolicyQualifier struct {
	XMLName xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SigPolicyQualifier"`
	SPURI   string   `xml:"http://uri.etsi.org/01903/v1.3.2# SPURI"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.4 (page 23)
type signedDataObjectProperties struct {
	XMLName          xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# SignedDataObjectProperties"`
	DataObjectFormat []dataObjectFormat
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.5 (page 36)
type dataObjectFormat struct {
	XMLName         xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# DataObjectFormat"`
	ObjectReference string   `xml:",attr"`
	MimeType        string
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.2 (page 22)
type unsignedProperties struct {
	XMLName                     xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# UnsignedProperties"`
	UnsignedSignatureProperties unsignedSignatureProperties
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.5 (page 24)
type unsignedSignatureProperties struct {
	XMLName           xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# UnsignedSignatureProperties"`
	CertificateValues certificateValues
	RevocationValues  revocationValues
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.1 (page 48)
type certificateValues struct {
	XMLName                     xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# CertificateValues"`
	EncapsulatedX509Certificate []encapsulatedX509Certificate
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.1 (page 48)
type encapsulatedX509Certificate struct {
	XMLName xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# EncapsulatedX509Certificate"`
	ID      string   `xml:"Id,attr"`
	Value   string   `xml:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.2 (page 49)
type revocationValues struct {
	XMLName    xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# RevocationValues"`
	OCSPValues ocspValues
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.2 (page 49)
type ocspValues struct {
	XMLName               xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# OCSPValues"`
	EncapsulatedOCSPValue encapsulatedOCSPValue
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.2 (page 49)
type encapsulatedOCSPValue struct {
	XMLName xml.Name `xml:"http://uri.etsi.org/01903/v1.3.2# EncapsulatedOCSPValue"`
	ID      string   `xml:"Id,attr"`
	Value   string   `xml:",chardata"`
}
