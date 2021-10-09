package bdoc

// http://www.etsi.org/deliver/etsi_ts/102900_102999/102918/01.02.01_60/ts_102918v010201p.pdf -
// annex A.5
type xadesSignatures struct {
	XMLElement c14n      `xmlx:"http://uri.etsi.org/02918/v1.2.1# XAdESSignatures"`
	Signature  signature // Restrict to exactly one Signature.
}

// https://www.w3.org/TR/xmldsig-core/#sec-Signature
type signature struct {
	XMLElement     c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# Signature"`
	ID             string `xmlx:"Id,attr,unique"`
	SignedInfo     signedInfo
	SignatureValue signatureValue
	KeyInfo        keyInfo // Restrict to exactly one entry: signer's X.509 certificate.
	Object         object  // Restrict to exactly one entry: XAdES QualifyingProperties.
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignedInfo
type signedInfo struct {
	XMLElement             c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# SignedInfo,c14nroot"`
	ID                     string `xmlx:"Id,attr,optional,unique"`
	CanonicalizationMethod canonicalizationMethod
	SignatureMethod        signatureMethod
	Reference              []reference
}

// https://www.w3.org/TR/xmldsig-core/#sec-CanonicalizationMethod
type canonicalizationMethod struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# CanonicalizationMethod"`
	Algorithm  string `xmlx:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignatureMethod
type signatureMethod struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# SignatureMethod"`
	Algorithm  string `xmlx:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-Reference
type reference struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# Reference"`
	ID         string `xmlx:"Id,attr,unique"`
	URI        string `xmlx:",attr"`
	Type       string `xmlx:",attr,optional"`

	Transforms   transforms `xmlx:",optional"`
	DigestMethod digestMethod
	DigestValue  digestValue
}

// https://www.w3.org/TR/xmldsig-core/#sec-Transforms
type transforms struct {
	XMLElement c14n `xmlx:"http://www.w3.org/2000/09/xmldsig# Transforms"`
	Transform  []transform
}

type transform struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# Transform"`
	Algorithm  string `xmlx:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-DigestMethod
type digestMethod struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# DigestMethod"`
	Algorithm  string `xmlx:",attr"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-DigestValue
type digestValue struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# DigestValue"`
	Value      string `xmlx:",chardata"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-SignatureValue
type signatureValue struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# SignatureValue,c14nroot"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	Value      string `xmlx:",chardata"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-KeyInfo
type keyInfo struct {
	XMLElement c14n     `xmlx:"http://www.w3.org/2000/09/xmldsig# KeyInfo"`
	ID         string   `xmlx:"Id,attr,optional,unique"`
	X509Data   x509Data // Restrict to exactly one entry: signer's X.509 certificate.
}

// https://www.w3.org/TR/xmldsig-core/#sec-X509Data
type x509Data struct {
	XMLElement      c14n            `xmlx:"http://www.w3.org/2000/09/xmldsig# X509Data"`
	X509Certificate x509Certificate // Restrict to exactly one entry: signer's X.509 certificate.
}

type x509Certificate struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# X509Certificate"`
	Value      string `xmlx:",chardata"`
}

// https://www.w3.org/TR/xmldsig-core/#sec-Object
type object struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# Object"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	// MimeType not allowed.
	// Encoding not allowed.

	QualifyingProperties qualifyingProperties // Restrict to exactly one entry: XAdES QualifyingProperties.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2
type qualifyingProperties struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# QualifyingProperties"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	Target     string `xmlx:",attr"`

	SignedProperties   signedProperties
	UnsignedProperties unsignedProperties `xmlx:",optional"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.1
type signedProperties struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignedProperties,c14nroot"`
	ID         string `xmlx:"Id,attr,unique"`

	SignedSignatureProperties  signedSignatureProperties
	SignedDataObjectProperties signedDataObjectProperties
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.3
type signedSignatureProperties struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignedSignatureProperties"`
	ID         string `xmlx:"Id,attr,optional,unique"`

	SigningTime               signingTime
	SigningCertificate        signingCertificate
	SignaturePolicyIdentifier signaturePolicyIdentifier `xmlx:",optional"`

	// SignatureProductionPlace and SignerRole are actually not allowed in
	// our use case, but empty elements are put here by DigiDocService: so
	// allow the elements, but no actual content.
	SignatureProductionPlace signatureProductionPlace `xmlx:",optional"`
	SignerRole               signerRole               `xmlx:",optional"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.1
type signingTime struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigningTime"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.2
type signingCertificate struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigningCertificate"`
	Cert       cert // Restrict to exactly one entry: signer's Cert.
}

type cert struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# Cert"`
	// URI not allowed.

	CertDigest   certDigest
	IssuerSerial issuerSerial
}

type certDigest struct {
	XMLElement   c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# CertDigest"`
	DigestMethod digestMethod
	DigestValue  digestValue
}

type issuerSerial struct {
	XMLElement       c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# IssuerSerial"`
	X509IssuerName   x509IssuerName
	X509SerialNumber x509SerialNumber
}

// https://www.w3.org/TR/xmldsig-core/#sec-X509Data
type x509IssuerName struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# X509IssuerName"`
	Value      string `xmlx:",chardata"`
}

type x509SerialNumber struct {
	XMLElement c14n   `xmlx:"http://www.w3.org/2000/09/xmldsig# X509SerialNumber"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.3
type signaturePolicyIdentifier struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignaturePolicyIdentifier"`

	SignaturePolicyID signaturePolicyID
	// SignaturePolicyImplied not allowed.
}

type signaturePolicyID struct {
	XMLElement  c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignaturePolicyId"`
	SigPolicyID sigPolicyID
	// Transforms not allowed.
	SigPolicyHash       sigPolicyHash
	SigPolicyQualifiers sigPolicyQualifiers
}

type sigPolicyID struct {
	XMLElement  c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigPolicyId"`
	Identifier  identifier
	Description description `xmlx:",optional"`
	// DocumentationReferences not allowed.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.1.2
type identifier struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# Identifier"`
	Qualifier  string `xmlx:",attr"`
	Value      string `xmlx:",chardata"`
}

type description struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# Description"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.3
type sigPolicyHash struct {
	XMLElement   c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigPolicyHash"`
	DigestMethod digestMethod
	DigestValue  digestValue
}

type sigPolicyQualifiers struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigPolicyQualifiers"`

	SigPolicyQualifier sigPolicyQualifier // Restrict to exactly one entry: BDOC URI.
}

type sigPolicyQualifier struct {
	XMLElement c14n  `xmlx:"http://uri.etsi.org/01903/v1.3.2# SigPolicyQualifier"`
	SPURI      spURI // Restrict to exactly one type: BDOC URI.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.3.1
type spURI struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SPURI"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.7
type signatureProductionPlace struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignatureProductionPlace"`

	City            city            `xmlx:",optional"`
	StateOrProvince stateOrProvince `xmlx:",optional"`
	PostalCode      postalCode      `xmlx:",optional"`
	CountryName     countryName     `xmlx:",optional"`
}

type city struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# City"`
	// No content allowed: see signedSignatureProperties.
}

type stateOrProvince struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# StateOrProvince"`
	// No content allowed: see signedSignatureProperties.
}

type postalCode struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# PostalCode"`
	// No content allowed: see signedSignatureProperties.
}

type countryName struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# CountryName"`
	// No content allowed: see signedSignatureProperties.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.8
type signerRole struct {
	XMLElement   c14n         `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignerRole"`
	ClaimedRoles claimedRoles `xmlx:",optional"`
	// CertifiedRoles not allowed.
}

type claimedRoles struct {
	XMLElement  c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# ClaimedRoles"`
	ClaimedRole []claimedRole
}

type claimedRole struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# ClaimedRole"`
	// No content allowed: see signedSignatureProperties.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.4
type signedDataObjectProperties struct {
	XMLElement       c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignedDataObjectProperties"`
	ID               string `xmlx:"Id,attr,optional,unique"`
	DataObjectFormat []dataObjectFormat
	// CommitmentTypeIndication not allowed.
	// AllDataObjectsTimeStamp not allowed.
	// IndividualDataObjectsTimeStamp not allowed.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.2.5
type dataObjectFormat struct {
	XMLElement      c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# DataObjectFormat"`
	ObjectReference string `xmlx:",attr"`

	// Description not allowed.
	// ObjectIdentifier not allowed.
	MimeType mimeType
	// Encoding not allowed.
}

type mimeType struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# MimeType"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.2
type unsignedProperties struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# UnsignedProperties"`
	ID         string `xmlx:"Id,attr,optional,unique"`

	UnsignedSignatureProperties unsignedSignatureProperties
	// UnsignedDataObjectProperties not allowed.
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 6.2.5
type unsignedSignatureProperties struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# UnsignedSignatureProperties"`
	ID         string `xmlx:"Id,attr,optional,unique"`

	// Restrict to an optional SignatureTimeStamp and mandatory
	// CertificateValues and RevocationValues.
	SignatureTimeStamp xadesTimeStamp `xmlx:",optional"`
	CertificateValues  certificateValues
	RevocationValues   revocationValues
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.1.4.3
type xadesTimeStamp struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# SignatureTimeStamp"`
	ID         string `xmlx:"Id,attr,unique"`

	CanonicalizationMethod canonicalizationMethod `xmlx:",optional"`
	EncapsulatedTimeStamp  encapsulatedTimeStamp  // Restrict to exactly one timestamp.
}

type encapsulatedTimeStamp struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# EncapsulatedTimeStamp"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	Value      string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.1
type certificateValues struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# CertificateValues"`
	ID         string `xmlx:"Id,attr,optional,unique"`

	EncapsulatedX509Certificate []encapsulatedX509Certificate
	// OtherCertificate not allowed.
}

type encapsulatedX509Certificate struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# EncapsulatedX509Certificate"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	// Encoding not allowed: always http://uri.etsi.org/01903/v1.2.2#DER.

	Value string `xmlx:",chardata"`
}

// http://uri.etsi.org/01903/v1.3.2/ts_101903v010302p.pdf - section 7.6.2
type revocationValues struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# RevocationValues"`
	ID         string `xmlx:"Id,attr,optional,unique"`

	// CRLValues not allowed.
	OCSPValues ocspValues
	// OtherValues not allowed.
}

type ocspValues struct {
	XMLElement c14n `xmlx:"http://uri.etsi.org/01903/v1.3.2# OCSPValues"`

	EncapsulatedOCSPValue encapsulatedOCSPValue // Restrict to exactly one entry: signer's OCSP response.
}

type encapsulatedOCSPValue struct {
	XMLElement c14n   `xmlx:"http://uri.etsi.org/01903/v1.3.2# EncapsulatedOCSPValue"`
	ID         string `xmlx:"Id,attr,optional,unique"`
	Value      string `xmlx:",chardata"`
}
