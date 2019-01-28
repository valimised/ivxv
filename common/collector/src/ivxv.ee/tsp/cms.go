package tsp

import (
	"bytes"
	"crypto"
	"crypto/sha1" // nolint: gosec, Required by RFC 2634.
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"

	"ivxv.ee/cryptoutil"

	// Import all hash functions supported by this package.
	_ "crypto/sha256"
	_ "crypto/sha512"
)

const (
	// OIDs of supported digest and signature algorithms and signed
	// attributes. Use strings instead of asn1.ObjectIdentifier, since they
	// will be used as keys in maps.
	idSHA256 = "2.16.840.1.101.3.4.2.1"
	idSHA384 = "2.16.840.1.101.3.4.2.2"
	idSHA512 = "2.16.840.1.101.3.4.2.3"

	idRSA   = "1.2.840.113549.1.1.1"
	idECDSA = "1.2.840.10045.4.2" // XXX: Is this the correct OID for ECDSA?

	// https://tools.ietf.org/html/rfc5652#section-11
	idContentType   = "1.2.840.113549.1.9.3"
	idMessageDigest = "1.2.840.113549.1.9.4"
	idSigningTime   = "1.2.840.113549.1.9.5"

	// https://tools.ietf.org/html/rfc2634#section-5.4
	idSigningCert = "1.2.840.113549.1.9.16.2.12"
)

var (
	digestAlgs = map[string]crypto.Hash{
		idSHA256: crypto.SHA256,
		idSHA384: crypto.SHA384,
		idSHA512: crypto.SHA512,
	}

	// TODO: We should have unit tests that cover all of these cases to
	//       ensure that we are using correct OIDs.
	signatureAlgs = map[string]map[string]x509.SignatureAlgorithm{
		idRSA: {
			idSHA256: x509.SHA256WithRSA,
			idSHA384: x509.SHA384WithRSA,
			idSHA512: x509.SHA512WithRSA,
		},
		idECDSA: {
			idSHA256: x509.ECDSAWithSHA256,
			idSHA384: x509.ECDSAWithSHA384,
			idSHA512: x509.ECDSAWithSHA512,
		},
	}

	// https://tools.ietf.org/html/rfc5652#section-5.1
	idSignedData = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
)

func (c *Client) checkSignedData(token timeStampToken, gen time.Time) error {
	if !token.ContentType.Equal(idSignedData) {
		return UnexpectedTSTContentType{ContentType: token.ContentType}
	}
	sData := token.Content

	// https://tools.ietf.org/html/rfc5652#page-10
	// Since eContentType is other than id-data the version is 3.
	if sData.Version != 3 {
		return UnexpectedSignedDataVersionError{
			Version: sData.Version,
		}
	}

	// Ignore the contents of SignedData.DigestAlgorithms: the point of it
	// is to allow us to compute the digests of EncapsulatedContentInfo in
	// the same pass as we are ASN.1 decoding SignedData. However, we are
	// not performing stream processing, so we have to make multiple passes
	// anyway.

	// We expect only one signature and one signer
	if len(sData.SignerInfos) != 1 {
		return NotASingleSignerError{Count: len(sData.SignerInfos)}
	}
	sInfo := sData.SignerInfos[0]

	// Find the signer's certificate from our trusted pool.
	cert, err := findCertificate(sInfo, c.signers)
	if err != nil {
		return UntrustedSigningCertificateError{Err: err}
	}

	// Require that the certificate be included in the response.
	var included bool
	for _, scert := range sData.Certificates {
		if bytes.Equal(cert.Raw, scert.RawContent) {
			included = true
			break
		}
	}
	if !included {
		return MissingSignerCertificateError{Signer: cert.Subject.CommonName}
	}

	if err := c.checkSignedAttributes(sInfo, sData.EncapContentInfo, gen, cert); err != nil {
		return SignedAttributeCheckError{Err: err}
	}

	if err := checkSignature(sInfo, cert); err != nil {
		return CheckSignatureError{Err: err}
	}

	return nil
}

func findCertificate(sInfo signerInfo, certs []*x509.Certificate) (c *x509.Certificate, err error) {
	// https://tools.ietf.org/html/rfc5652#page-14
	// SignerInfo.Version depends on the choice of SignerIdentifier
	switch sInfo.Version {
	case 1:
		issuer := sInfo.IssuerAndSerialNumber.Issuer
		serial := sInfo.IssuerAndSerialNumber.SerialNumber
		if len(issuer) == 0 {
			return nil, Version1MissingIASNError{}
		}
		for _, c := range certs {
			if hasIssuerSerial(c, issuer, serial) {
				return c, nil
			}
		}
		return nil, IASNCertificateNotFoundError{
			Issuer: issuer,
			Serial: serial,
		}
	case 3:
		if len(sInfo.SubjectKeyIdentifier) == 0 {
			return nil, Version3MissingSKIError{}
		}
		for _, c := range certs {
			if bytes.Equal(sInfo.SubjectKeyIdentifier, c.SubjectKeyId) {
				return c, nil
			}
		}
		return nil, SKICertificateNotFoundError{SKI: sInfo.SubjectKeyIdentifier}
	default:
		return nil, SignerInfoVersionError{Version: sInfo.Version}
	}
}

func hasIssuerSerial(cert *x509.Certificate, issuer pkix.RDNSequence, serial *big.Int) bool {
	// Add all parsed non-standard names to serialized RDN sequence.
	cert.Issuer.ExtraNames = cert.Issuer.Names
	return cert.SerialNumber.Cmp(serial) == 0 &&
		cryptoutil.RDNSequenceEqual(cert.Issuer.ToRDNSequence(), issuer)
}

func (c *Client) checkSignedAttributes(sInfo signerInfo, encap encapsulatedContentInfo,
	gen time.Time, signer *x509.Certificate) (err error) {

	attrMap := make(map[string]bool)
	for _, attr := range sInfo.SignedAttrs {
		// Check that there are no duplicate signed attributes
		attrID := attr.AttrType.String()
		if attrMap[attrID] {
			return DuplicateSignedAttrError{Attribute: attrID}
		}
		attrMap[attrID] = true

		// All of the attributes that we are checking here require that
		// the attribute value be a SET with a single entry. Check that
		// the value is a SET and pass all content bytes as a single
		// entry.
		//
		// Since unmarshaling of a raw value has succeeded, we can
		// assume that at least the tag and length bytes exist.
		if tag := attr.AttrValue.FullBytes[0]; tag != 49 {
			return AttributeValueNotASetError{Tag: tag}
		}
		value := attr.AttrValue.Bytes

		switch attrID {
		case idContentType:
			if err = checkContentType(value, encap.EContentType); err != nil {
				return CheckContentTypeError{Err: err}
			}
		case idMessageDigest:
			if err = checkMessageDigest(value,
				sInfo.DigestAlgorithm, encap.EContent); err != nil {

				return CheckMsgDigestError{Err: err}
			}
		case idSigningTime:
			if err = c.checkSigningTime(value, gen); err != nil {
				return CheckSigningTimeError{Err: err}
			}
		case idSigningCert:
			if err = checkSigningCert(value, signer); err != nil {
				return CheckSigningCertError{Err: err}
			}
		default:
			return UnknownAttributeError{Attr: attrID}
		}
	}
	if !attrMap[idContentType] {
		return NoSignedContentTypeError{}
	}
	if !attrMap[idMessageDigest] {
		return NoSignedMsgDigestError{}
	}
	if !attrMap[idSigningTime] {
		return NoSignedGenTimeError{}
	}
	if !attrMap[idSigningCert] {
		return NoSigningCertError{}
	}
	return
}

func checkContentType(value []byte, encapType asn1.ObjectIdentifier) (err error) {
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(value, &oid)
	if err != nil {
		return ContentTypeUnmarshalError{Err: err}
	}
	if len(rest) > 0 {
		return ContentTypeUnmarshalExcessBytesError{Bytes: rest}
	}
	if !oid.Equal(encapType) {
		return SignedAttrContentTypeMismatchError{
			SignedAttribute: oid,
			EContentType:    encapType,
		}
	}
	return
}

func checkMessageDigest(value []byte, alg pkix.AlgorithmIdentifier, encap []byte) (err error) {
	var digest []byte
	rest, err := asn1.Unmarshal(value, &digest)
	if err != nil {
		return MessageDigestUnmarshalError{Err: err}
	}
	if len(rest) > 0 {
		return MessageDigestUnmarshalExcessBytesError{Bytes: rest}
	}

	chash, ok := digestAlgs[alg.Algorithm.String()]
	if !ok {
		return UnsupportedDigestAlgorithm{Algorithm: alg.Algorithm}
	}
	hash := chash.New()
	hash.Write(encap)
	calculated := hash.Sum(nil)

	if !bytes.Equal(digest, calculated) {
		return SignedAttributeMsgDigestMismatchError{
			SignedAttribute: digest,
			CalculatedHash:  calculated,
		}
	}
	return
}

func (c *Client) checkSigningTime(value []byte, gen time.Time) (err error) {
	var t time.Time
	rest, err := asn1.Unmarshal(value, &t)
	if err != nil {
		return SigningTimeUnmarshalError{Err: err}
	}
	if len(rest) > 0 {
		return SigningTimeUnmarshalExcessBytesError{Bytes: rest}
	}

	diff := t.Sub(gen)
	if diff < 0 || diff > c.delay {
		return SignedAttrSigningTimeMismatch{
			SignedAttribute: t,
			GenTime:         gen,
		}
	}
	return
}

func checkSigningCert(value []byte, signer *x509.Certificate) (err error) {
	var signingCert signingCertificate
	rest, err := asn1.Unmarshal(value, &signingCert)
	if err != nil {
		return SigningCertUnmarshalError{Err: err}
	}
	if len(rest) > 0 {
		return SigningCertUnmarshalExcessBytes{Bytes: rest}
	}

	if len(signingCert.Certs) == 0 {
		return SigningCertAttrCertsMissing{}
	}

	// https://tools.ietf.org/html/rfc2634#section-5.4
	// "The first certificate identified in the sequence of certificate
	// identifiers MUST be the certificate used to verify the signature."
	// We ignore all other chain certificates, because we receive the trusted
	// certificate via the configuration.
	essCert := signingCert.Certs[0]

	// Check the signed attribute hash to the certificate hash
	certHash := sha1.Sum(signer.Raw) // nolint: gosec, SHA1 required by RFC 2634.
	if !bytes.Equal(certHash[:], essCert.CertHash) {
		return SingingCertAttrHashMismatch{
			SigningAttrHash: essCert.CertHash,
			CertificateHash: certHash,
		}
	}

	// Check the issuer and serial number, if present.
	issuer := essCert.IssuerAndSerialNumber.Issuer.DirectoryName
	serial := essCert.IssuerAndSerialNumber.SerialNumber
	if len(issuer) > 0 {
		if !hasIssuerSerial(signer, issuer, serial) {
			return SigningCertIASNMismatchError{
				SignerIssuer: signer.Issuer,
				SignerSerial: signer.SerialNumber,
				AttrIssuer:   issuer,
				AttrSerial:   serial,
			}
		}
	}
	return
}

func checkSignature(sInfo signerInfo, cert *x509.Certificate) (err error) {
	if len(sInfo.Signature) == 0 {
		return NoSignatureError{}
	}

	digMap, ok := signatureAlgs[sInfo.SignatureAlgorithm.Algorithm.String()]
	if !ok {
		return SigAlgorithmNotSupported{
			Algorithm: sInfo.SignatureAlgorithm.Algorithm,
		}
	}

	algo, ok := digMap[sInfo.DigestAlgorithm.Algorithm.String()]
	if !ok {
		return SigDigestAlgorithmNotSupported{
			Algorithm: sInfo.DigestAlgorithm.Algorithm,
		}
	}

	var content []byte
	if content, err = asn1.Marshal(sInfo.SignedAttrs); err != nil {
		return SignedAttrsMarshalError{Err: err}
	}

	// https://tools.ietf.org/html/rfc5652#section-5.4
	content[0] = 49 // SET OF

	if err = cert.CheckSignature(algo, content, sInfo.Signature); err != nil {
		return CertificateCheckSignatureError{Err: err}
	}

	return
}
