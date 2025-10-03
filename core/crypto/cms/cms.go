package cms

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"tivi.io/core/crypto/util"
)

const (
	idSHA256 = "2.16.840.1.101.3.4.2.1"
	idSHA384 = "2.16.840.1.101.3.4.2.2"
	idSHA512 = "2.16.840.1.101.3.4.2.3"
)

var (
	idContentTypeData       = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	idContentTypeSignedData = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	digestAlgorithms        = map[string]crypto.Hash{
		idSHA256: crypto.SHA256,
		idSHA384: crypto.SHA384,
		idSHA512: crypto.SHA512,
	}
)

// VerifySignedData verifies the signed data, matching it to the signers' certificates provided.
// It also accepts a set of additional signed and unsigned attributes to be verified.
// Returns the data that was signed.
func VerifySignedData(derSignedData []byte, certs []*x509.Certificate, detachedData []byte, otherSignedAttrs, otherUnsignedAttrs map[string]AttrToVerify) error {
	signedData, err := unmarshalSignedData(derSignedData)
	if err != nil {
		return fmt.Errorf("unmarshal signed data: %w", err)
	}
	// There may be multiple independent signers, f.e., in CAdES
	// https://datatracker.ietf.org/doc/html/rfc5126#section-5.6
	if len(signedData.SignerInfos) == 0 {
		return fmt.Errorf("missing signers' info")
	}
	for _, signerInfo := range signedData.SignerInfos {
		// Get the local certificate of the signer
		cert, err := getCertificate(signerInfo, certs)
		if err != nil {
			return fmt.Errorf("find certificate: %w", err)
		}
		// VerifyAll if the same certificate comes in the response
		var certInResponse bool
		for _, sCert := range signedData.Certificates {
			if bytes.Equal(cert.Raw, sCert.RawContent) {
				certInResponse = true
				break
			}
		}
		if !certInResponse {
			return fmt.Errorf("missing signed certificate")
		}
		if err = verifySignedAttributes(signerInfo, signedData.EncapContentInfo, cert, detachedData, otherSignedAttrs); err != nil {
			return fmt.Errorf("verify signed attributes: %w", err)
		}
		if err = verifyUnsignedAttributes(signerInfo, otherUnsignedAttrs); err != nil {
			return fmt.Errorf("verify unsigned attributes: %w", err)
		}
		if err = verifySignature(signerInfo, cert); err != nil {
			return fmt.Errorf("verify signature: %w", err)
		}
	}
	return nil
}

// GetDataEContent unmarshals a DER encoded signedData container and
// returns the Encapsulated Content, if there is any attached data.
func GetDataEContent(derSignedData []byte) ([]byte, error) {
	signedData, err := unmarshalSignedData(derSignedData)
	if err != nil {
		return nil, fmt.Errorf("unmarshal signed data: %w", err)
	}
	eContent := signedData.EncapContentInfo.EContent
	if eContent == nil {
		return nil, fmt.Errorf("no attached data")
	}
	return eContent, nil
}

func unmarshalSignedData(derSignedData []byte) (SignedData, error) {
	var ctInfo ContentInfo
	rest, err := asn1.Unmarshal(derSignedData, &ctInfo)
	if err != nil {
		return SignedData{}, fmt.Errorf("unmarshal content info: %s", err)
	}
	if len(rest) > 0 {
		return SignedData{}, fmt.Errorf("unmarshal content info excess bytes: %d", len(rest))
	}
	if !ctInfo.ContentType.Equal(idContentTypeSignedData) {
		return SignedData{}, fmt.Errorf("bad content info content type: wanted '%s' got '%s'", idContentTypeSignedData, ctInfo.ContentType)
	}
	signedData := ctInfo.Content
	version := getVersion(signedData.EncapContentInfo)
	if signedData.Version != version {
		return SignedData{}, fmt.Errorf("bad signed data version: wanted '%d' got '%d'", version, signedData.Version)
	}
	return signedData, nil
}

// https://tools.ietf.org/html/rfc5652#page-10
// encapContentInfo eContentType is other than id-data, then version MUST be 3
// else version is 1
func getVersion(encap EncapsulatedContentInfo) int {
	if !encap.EContentType.Equal(idContentTypeData) {
		return 3
	}
	return 1
}

func getCertificate(signer SignerInfo, certs []*x509.Certificate) (*x509.Certificate, error) {
	// https://tools.ietf.org/html/rfc5652#page-14
	switch signer.Version {
	case 1:
		issuer := signer.IssuerAndSerialNumber.Issuer
		serial := signer.IssuerAndSerialNumber.SerialNumber
		if len(issuer) == 0 {
			return nil, fmt.Errorf("version 1 missing issuer")
		}
		for _, cert := range certs {
			if hasIssuerSerial(cert, issuer, serial) {
				return cert, nil
			}
		}
		return nil, fmt.Errorf("certificate not found: issuer '%s' serial '%s'", issuer, serial)
	case 3:
		if len(signer.SubjectKeyIdentifier) == 0 {
			return nil, fmt.Errorf("version 3 missing subject key identifier")
		}
		for _, cert := range certs {
			if bytes.Equal(signer.SubjectKeyIdentifier, cert.SubjectKeyId) {
				return cert, nil
			}
		}
		return nil, fmt.Errorf("certificate not found: ski '%s'", string(signer.SubjectKeyIdentifier))
	default:
		return nil, fmt.Errorf("bad signer info version: wanted '%d' or '%d' got '%d'", 1, 3, signer.Version)
	}
}

func hasIssuerSerial(cert *x509.Certificate, issuer pkix.RDNSequence, serial *big.Int) bool {
	// add all parsed non-standard names to serialized RDN sequence
	cert.Issuer.ExtraNames = cert.Issuer.Names
	return cert.SerialNumber.Cmp(serial) == 0 && util.IsRDNSequenceEqual(cert.Issuer.ToRDNSequence(), issuer)
}

const (
	// https://tools.ietf.org/html/rfc5652#section-11
	idContentType   = "1.2.840.113549.1.9.3"
	idMessageDigest = "1.2.840.113549.1.9.4"
	idSigningTime   = "1.2.840.113549.1.9.5"

	// https://tools.ietf.org/html/rfc2634#section-5.4
	idSigningCert = "1.2.840.113549.1.9.16.2.12"

	// https://tools.ietf.org/html/rfc5035#section-3
	idSigningCertV2 = "1.2.840.113549.1.9.16.2.47"

	// https://tools.ietf.org/html/rfc6211#section-2
	idCMSAlgorithmProtection = "1.2.840.113549.1.9.52"
)

func verifySignedAttributes(signer SignerInfo, encap EncapsulatedContentInfo, cert *x509.Certificate, detachedData []byte, other map[string]AttrToVerify) error {
	attributes := make(map[string]bool)
	for _, attr := range signer.SignedAttrs {
		attributeID := attr.AttrType.String()
		if attributes[attributeID] {
			return fmt.Errorf("found duplicate signed attribute: %s", attributeID)
		}
		attributes[attributeID] = true
		// Extract the ASN.1 universal tag
		tag := attr.AttrValue.FullBytes[0] & 31
		// All attribute values required to be a SET with a single entry
		if tag != asn1.TagSet {
			return fmt.Errorf("attribute value not a set: %b", tag)
		}
		value := attr.AttrValue.Bytes
		// 1. The mandatory CMS signed attributes are the content-type and message-digest.
		// 2. If some optional attribute is in the data, it is up to the user
		//    of the API to provide a callback. Nothing shall happen if the
		//    optional attribute is in the data, but the client chose not to
		//    verify it, f.e., as it can happen with the signingTime.
		// 3. If the user provides a callback for an attribute that is missing,
		//    there shall be an error - detected after the switch statement.
		switch attributeID {
		case other[attributeID].AttributeID:
			if err := other[attributeID].VerifyCallback(value); err != nil {
				return fmt.Errorf("verify other attribute with id '%s': %w", attributeID, err)
			}
		case idContentType:
			if err := verifyContentType(value, encap.EContentType); err != nil {
				return fmt.Errorf("verify content type: %w", err)
			}
		case idMessageDigest:
			data := encap.EContent
			if data == nil {
				if detachedData == nil {
					return fmt.Errorf("missing detached data to be verified")
				}
				data = detachedData
			} else if detachedData != nil {
				if len(data) != len(detachedData) {
					return fmt.Errorf("detached data length mismatches embedded data")
				}
				if !bytes.Equal(data, detachedData) {
					return fmt.Errorf("detached data content mismatches embedded data")
				}
			}
			if err := verifyMessageDigest(value, signer.DigestAlgorithm, data); err != nil {
				return fmt.Errorf("verify message digest: %w", err)
			}
		case idSigningCert, idSigningCertV2:
			if err := verifySigningCertificate(value, cert, attributeID == idSigningCertV2); err != nil {
				return fmt.Errorf("verify signing certificate: %w", err)
			}
		case idCMSAlgorithmProtection:
			if err := verifyCMSAlgorithmProtection(value, signer); err != nil {
				return fmt.Errorf("verify cms algorithm protection: %w", err)
			}
		}
	}
	if !attributes[idContentType] {
		return fmt.Errorf("missing signed content type")
	}
	if !attributes[idMessageDigest] {
		return fmt.Errorf("missing signed message digest")
	}
	for k := range other {
		if !attributes[k] {
			return fmt.Errorf("missing attribute with id '%s'", k)
		}
	}
	return nil
}

func verifyContentType(value []byte, encapType asn1.ObjectIdentifier) error {
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(value, &oid)
	if err != nil {
		return fmt.Errorf("unmarshal object identifier: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal object identifier excess bytes: %d", len(rest))
	}
	if !oid.Equal(encapType) {
		return fmt.Errorf("object identifier not equal to encap type: object identifier '%s' encap type '%s'", oid, encapType)
	}
	return nil
}

func verifyMessageDigest(value []byte, alg pkix.AlgorithmIdentifier, data []byte) error {
	var digest []byte
	rest, err := asn1.Unmarshal(value, &digest)
	if err != nil {
		return fmt.Errorf("unmarshal message digest: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal message digest excess bytes: %d", len(rest))
	}
	cHash, ok := digestAlgorithms[alg.Algorithm.String()]
	if !ok {
		return fmt.Errorf("unsupported digest algorithm: %s", alg.Algorithm)
	}
	hash := cHash.New()
	hash.Write(data)
	calculated := hash.Sum(nil)
	if !bytes.Equal(digest, calculated) {
		return fmt.Errorf("different hashed messages with algorithm: %s", alg.Algorithm)
	}
	return nil
}

func verifySigningCertificate(value []byte, cert *x509.Certificate, v2 bool) error {
	// SigningCertificateV2 structure is backwards-compatible to V1
	var signingCert SigningCertificateV2
	rest, err := asn1.Unmarshal(value, &signingCert)
	if err != nil {
		return fmt.Errorf("unmarshal signing certificate: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal signing certificate excess bytes: %d", len(rest))
	}
	if len(signingCert.Certs) == 0 {
		return fmt.Errorf("no signing certificate certs attribute")
	}
	// https://tools.ietf.org/html/rfc5035#section-3
	// "The first certificate identified in the sequence of certificate
	// identifiers MUST be the certificate used to verify the signature."
	essCert := signingCert.Certs[0]
	hashOID := essCert.HashAlgorithm.Algorithm
	var cHash crypto.Hash
	if v2 {
		// SigningCertificateV2 hash algorithm defaults to SHA-256
		// but can be explicitly provided
		cHash = crypto.SHA256
		if hashOID != nil {
			var ok bool
			if cHash, ok = digestAlgorithms[hashOID.String()]; !ok {
				return fmt.Errorf("unsupported digest algorithm: %s", hashOID.String())
			}
		}
	} else {
		// SigningCertificateV1 hash algorithm defaults to SHA-1
		// and no algorithm must be provided
		cHash = crypto.SHA1
		if hashOID != nil {
			return fmt.Errorf("(signing certificate V1) no algorithm should be provided: got '%s'", hashOID.String())
		}
	}
	hash := cHash.New()
	hash.Write(cert.Raw)
	certHash := hash.Sum(nil)
	if !bytes.Equal(certHash, essCert.CertHash) {
		return fmt.Errorf("different certificate hashes with algorithm: %s", hashOID.String())
	}
	issuer := essCert.IssuerAndSerialNumber.Issuer.DirectoryName
	serial := essCert.IssuerAndSerialNumber.SerialNumber
	// If present, verify the correct issuer and serial number
	if len(issuer) > 0 {
		if !hasIssuerSerial(cert, issuer, serial) {
			return fmt.Errorf("issuer and serial not found: issuer '%s' serial' '%s'", issuer, serial)
		}
	}
	return nil
}

func verifyCMSAlgorithmProtection(value []byte, signer SignerInfo) error {
	var protection CMSAlgorithmProtection
	rest, err := asn1.Unmarshal(value, &protection)
	if err != nil {
		return fmt.Errorf("unmarshal cms algorithm protection: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal cms algorithm protection excess bytes: %d", len(rest))
	}
	if !util.AlgorithmIdentifierCmp(signer.DigestAlgorithm, protection.DigestAlgorithm) {
		return fmt.Errorf("different digest algorithms: signer '%s' protection '%s'", signer.DigestAlgorithm.Algorithm, protection.DigestAlgorithm.Algorithm)
	}
	if !util.AlgorithmIdentifierCmp(signer.SignatureAlgorithm, protection.SignatureAlgorithm) {
		return fmt.Errorf("different signature algorithms: signer '%s' protection '%s'", signer.SignatureAlgorithm.Algorithm, protection.SignatureAlgorithm.Algorithm)
	}
	return nil
}

func verifyUnsignedAttributes(signer SignerInfo, unsignedAttrs map[string]AttrToVerify) error {
	attributes := make(map[string]bool)
	for _, attr := range signer.UnsignedAttrs {
		attributeID := attr.AttrType.String()
		if attributes[attributeID] {
			return fmt.Errorf("found duplicate unsigned attribute: %s", attributeID)
		}
		attributes[attributeID] = true
		// Extract the ASN.1 universal tag
		tag := attr.AttrValue.FullBytes[0] & 31
		// All attribute values required to be a SET with a single entry
		if tag != asn1.TagSet {
			return fmt.Errorf("attribute value not a set: %b", tag)
		}
		value := attr.AttrValue.Bytes
		if attr, ok := unsignedAttrs[attributeID]; ok {
			if err := attr.VerifyCallback(value); err != nil {
				return fmt.Errorf("verify other attribute with id '%s': %w", attributeID, err)
			}
		}
	}
	for k := range unsignedAttrs {
		if !attributes[k] {
			return fmt.Errorf("missing attribute with id '%s'", k)
		}
	}
	return nil
}

const (
	idRSAEncryption = "1.2.840.113549.1.1.1"

	idSHA256WithRSAPPS = "1.2.840.113549.1.1.10"

	idSHA256WithRSAEncryption = "1.2.840.113549.1.1.11"
	idSHA384WithRSAEncryption = "1.2.840.113549.1.1.12"
	idSHA512WithRSAEncryption = "1.2.840.113549.1.1.13"

	idECDSAWithSHA256 = "1.2.840.10045.4.3.2"
	idECDSAWithSHA384 = "1.2.840.10045.4.3.3"
	idECDSAWithSHA512 = "1.2.840.10045.4.3.4"
)

var (
	signatureAlgorithms = map[string]x509.SignatureAlgorithm{
		idRSAEncryption: x509.SHA256WithRSA,

		idSHA256WithRSAPPS: x509.SHA256WithRSAPSS,

		idSHA256WithRSAEncryption: x509.SHA256WithRSA,
		idSHA384WithRSAEncryption: x509.SHA384WithRSA,
		idSHA512WithRSAEncryption: x509.SHA512WithRSA,

		idECDSAWithSHA256: x509.ECDSAWithSHA256,
		idECDSAWithSHA384: x509.ECDSAWithSHA384,
		idECDSAWithSHA512: x509.ECDSAWithSHA512,
	}

	signatureDigestOIDs = map[string]string{
		idRSAEncryption: idSHA256,

		idSHA256WithRSAPPS: idSHA256,

		idSHA256WithRSAEncryption: idSHA256,
		idSHA384WithRSAEncryption: idSHA384,
		idSHA512WithRSAEncryption: idSHA512,

		idECDSAWithSHA256: idSHA256,
		idECDSAWithSHA384: idSHA384,
		idECDSAWithSHA512: idSHA512,
	}
)

func verifySignature(signer SignerInfo, cert *x509.Certificate) error {
	if len(signer.Signature) == 0 {
		return fmt.Errorf("no signature found")
	}
	signatureOID := signer.SignatureAlgorithm.Algorithm.String()
	algorithm, ok := signatureAlgorithms[signatureOID]
	if !ok {
		return fmt.Errorf("signature algorithm not supported: %s", signatureOID)
	}
	if signer.DigestAlgorithm.Algorithm.String() != signatureDigestOIDs[signatureOID] {
		return fmt.Errorf("different signature algorithms: digest '%s' signature '%s'", signer.DigestAlgorithm.Algorithm, signatureOID)
	}
	var err error
	var content []byte
	if content, err = asn1.Marshal(signer.SignedAttrs); err != nil {
		return fmt.Errorf("marshal signed attributes: %w", err)
	}
	// https://tools.ietf.org/html/rfc5652#section-5.4
	content[0] = 49
	if err = cert.CheckSignature(algorithm, content, signer.Signature); err != nil {
		return fmt.Errorf("check certificate signature: %w", err)
	}
	return nil
}
