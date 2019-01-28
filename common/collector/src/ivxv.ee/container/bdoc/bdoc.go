/*
Package bdoc implements BDOC verification following the BDOC specification
version 2.1.2:2014.
http://www.id.ee/public/bdoc-spec212-est.pdf
http://www.id.ee/public/bdoc-spec212-eng.pdf
*/
package bdoc // import "ivxv.ee/container/bdoc"

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"io"
	"math/big"
	"net/url"
	"strings"
	"time"

	// Import supported hash functions for calculating TM nonce.
	_ "crypto/sha256"
	_ "crypto/sha512"

	"ivxv.ee/container"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
	"ivxv.ee/tsp"
	"ivxv.ee/yaml"
)

func init() {
	container.Register(container.BDOC, configure, unverifiedOpen, container.ASiCE)
}

// configure configures a new Opener and returns its Open function.
func configure(n yaml.Node) (container.OpenFunc, error) {
	var c Conf
	if err := yaml.Apply(n, &c); err != nil {
		return nil, ConfigurationYAMLError{Err: err}
	}

	o, err := New(&c)
	if err != nil {
		return nil, ConfigurationError{Err: err}
	}

	// Wrap o.Open in a short function to cast the returned value from
	// *BDOC to container.Container.
	return func(r io.Reader) (container.Container, error) {
		return o.Open(r)
	}, nil
}

func unverifiedOpen(encoded io.Reader) (container.Container, error) {
	// The bootstrap container and files in it should fit inside 10 MiB.
	const limit = 10 * 1024 * 1024
	files, err := openASiCE(encoded, limit, limit, false)
	if err != nil {
		return nil, UnverifiedOpenError{Err: err}
	}
	return &BDOC{files: files}, nil
}

// Profile identifies a BDOC signature profile.
type Profile string

// Enumeration of BDOC signature profiles.
const (
	BES Profile = "BES" // Base profile.
	TM          = "TM"  // Timemark profile.
	TS          = "TS"  // Timestamp profile.
)

// Conf contains the configurable options for the BDOC container opener. It
// only contains serialized values such that it can easily be unmarshaled from
// a file.
type Conf struct {
	// BDOCSize is the maximum accepted size of BDOC containers read from
	// streams into memory. Note that this limit does not apply to sources
	// which already support random access.
	BDOCSize int64

	// FileSize is the maximum accepted decompressed size of files in the container.
	FileSize int64

	// Roots contains the root certificates used for verification in PEM format
	Roots []string

	// Roots contains the intermediate certificates used for verification in PEM format
	Intermediates []string

	// Profile specifies the profile for opened BDOC containers.
	Profile Profile

	// OCSP is the configuration for the OCSP client used to check a
	// certificate's revocation status if Profile is TM or TS. Only offline
	// checks are performed.
	OCSP ocsp.Conf

	// TSP is the configuration for the TSP client used to check timestamps
	// if Profile is TS.
	TSP tsp.Conf
}

// Opener is a configured BDOC container opener.
type Opener struct {
	bdocSize int64
	fileSize int64
	rpool    *x509.CertPool
	ipool    *x509.CertPool
	ocsp     *ocsp.Client
	tsp      *tsp.Client
	profile  Profile
}

// New returns a new BDOC container opener.
func New(c *Conf) (*Opener, error) {
	o := &Opener{
		bdocSize: c.BDOCSize,
		fileSize: c.FileSize,
		profile:  c.Profile,
	}

	if o.bdocSize <= 0 {
		return nil, InvalidBDOCSizeError{Size: o.bdocSize}
	}

	if o.fileSize <= 0 {
		return nil, InvalidFileSizeError{Size: o.fileSize}
	}

	if len(c.Roots) == 0 {
		return nil, UnconfiguredRootsError{}
	}
	var err error
	if o.rpool, err = cryptoutil.PEMCertificatePool(c.Roots...); err != nil {
		return nil, RootsParseError{Err: err}
	}
	if o.ipool, err = cryptoutil.PEMCertificatePool(c.Intermediates...); err != nil {
		return nil, IntermediatesParseError{Err: err}
	}

	switch c.Profile {
	case TS: // Create TSP client for checking timestamps.
		if o.tsp, err = tsp.New(&c.TSP); err != nil {
			return nil, TSPClientError{Err: err}
		}
		fallthrough
	case TM: // Create OCSP client for checking revocation status.
		if o.ocsp, err = ocsp.New(&c.OCSP); err != nil {
			return nil, OCSPClientError{Err: err}
		}
	case BES: // No additional setup.
	default:
		return nil, UnsupportedProfileError{Profile: c.Profile}
	}
	return o, nil
}

// BDOC is a parsed BDOC container
type BDOC struct {
	signatures []container.Signature
	files      map[string]*asiceFile
	ocspValues map[string][]byte
	tspData    map[string]*bytes.Buffer
}

// Open opens and verifies a BDOC signature container.
func (o *Opener) Open(encoded io.Reader) (bdoc *BDOC, err error) {
	// Open the container and get all the files from it.
	files, err := openASiCE(encoded, o.bdocSize, o.fileSize, true)
	if err != nil {
		return nil, OpenBDOCContainerError{Err: err}
	}

	bdoc = &BDOC{
		files:      files,
		ocspValues: make(map[string][]byte),
		tspData:    make(map[string]*bytes.Buffer),
	}
	// Pass b as explicit argument so that it gets evaluated now and we can
	// later safely return nil. Do not do the same for err, because we want
	// to know its value on return, not now.
	defer func(bdoc *BDOC) {
		// Close all files on error.
		if err != nil {
			bdoc.Close() // nolint: errcheck, always returns nil.
		}
	}(bdoc)

	// Parse the signatures from the container.
	var signatures []*signature
	for key, file := range bdoc.files {
		if !file.signature() {
			continue
		}

		// Parse the XAdES signatures.
		var x xadesSignatures
		if err = parseXML(file.data.Bytes(), &x); err != nil {
			return nil, SignatureXMLError{Name: file.name, Err: err}
		}
		s := &x.Signature // We allow only one signature per file.
		signatures = append(signatures, s)

		// Remove the signature from BDOC data files and close it: we
		// do not need it anymore after it is parsed.
		delete(bdoc.files, key)
		file.close()

		// Parse signer certificate.
		var cert *x509.Certificate
		if cert, err = parseBase64Certificate(s.KeyInfo.
			X509Data.X509Certificate.Value); err != nil {

			return nil, SignerCertificateParseError{
				FileName:    file.name,
				SignatureID: s.ID,
				Err:         err,
			}
		}

		// Get signing time. If profile is TM, then this will get
		// replaced by the OCSP response time. If profile is TS, then
		// this will get replaced by the timestamp time.
		var signingTime time.Time
		signingTime, err = time.Parse(time.RFC3339, s.Object.
			QualifyingProperties.SignedProperties.
			SignedSignatureProperties.SigningTime.Value)
		if err != nil {
			return nil, SigningTimeParseError{Err: err}
		}

		// Append to BDOC signatures.
		bdoc.signatures = append(bdoc.signatures, container.Signature{
			ID:     s.ID,
			Signer: cert,

			// Issuer will be determined and inserted during
			// certificate checking.

			SigningTime: signingTime,
		})

		// Decode and store the ivxv.ee/q11n/ocsp.SignatureValuer value.
		bdoc.ocspValues[s.ID], err = b64d(s.SignatureValue.Value)
		if err != nil {
			return nil, SignatureValueDecodeError{Err: err}
		}

		// Canonicalize and store the ivxv.ee/q11n/tsp.TimestampDataer data.
		sigval := buffer()
		writeXML(&s.SignatureValue, sigval)
		bdoc.tspData[s.ID] = sigval
	}

	// Check the parsed signatures against the data files in the container.
	for i, s := range signatures {
		if err = o.check(s, &bdoc.signatures[i], bdoc.files); err != nil {
			return nil, CheckSignatureError{Signature: s.ID, Err: err}
		}
	}
	return bdoc, nil
}

// Signatures implements the container.Container interface.
func (b *BDOC) Signatures() []container.Signature {
	return b.signatures
}

// Data implements the container.Container interface.
func (b *BDOC) Data() map[string][]byte {
	files := make(map[string][]byte)
	for _, file := range b.files {
		files[file.name] = file.data.Bytes()
	}
	return files
}

// Close implements the container.Container interface.
func (b *BDOC) Close() error {
	for _, file := range b.files {
		file.close()
	}
	b.files = nil
	for _, buf := range b.tspData {
		release(buf)
	}
	b.tspData = nil
	return nil
}

// SignatureValue implements the ivxv.ee/q11n/ocsp.SignatureValuer interface.
// It returns the raw signature value for the requested ID.
func (b *BDOC) SignatureValue(id string) ([]byte, error) {
	value, ok := b.ocspValues[id]
	if !ok {
		return nil, SignatureValueNoSuchIDError{ID: id}
	}
	return value, nil
}

// TimestampData implements the ivxv.ee/q11n/tsp.TimestampDataer interface. It
// returns the canonical SignatureValue XML element for the requested ID.
func (b *BDOC) TimestampData(id string) ([]byte, error) {
	data, ok := b.tspData[id]
	if !ok {
		return nil, TimestampDataNoSuchIDError{ID: id}
	}
	return data.Bytes(), nil
}

const xmlc14n11 = "http://www.w3.org/2006/12/xml-c14n11"

var signAlgorithms = map[string]x509.SignatureAlgorithm{
	"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256":   x509.SHA256WithRSA,
	"http://www.w3.org/2001/04/xmldsig-more#rsa-sha384":   x509.SHA384WithRSA,
	"http://www.w3.org/2001/04/xmldsig-more#rsa-sha512":   x509.SHA512WithRSA,
	"http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256": x509.ECDSAWithSHA256,
	"http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha384": x509.ECDSAWithSHA384,
	"http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha512": x509.ECDSAWithSHA512,
}

func (o *Opener) check(s *signature, c *container.Signature, files map[string]*asiceFile) error {
	c14n := s.SignedInfo.CanonicalizationMethod.Algorithm
	if c14n != xmlc14n11 {
		return UnsupportedCanonicalizationAlgorithmError{Algorithm: c14n}
	}

	signing := s.SignedInfo.SignatureMethod.Algorithm
	x509sa, ok := signAlgorithms[signing]
	if !ok {
		return UnsupportedSigningAlgorithmError{Algorithm: signing}
	}

	var err error
	if err = checkReferences(s, files); err != nil {
		return err
	}

	// We use the declared signing time here to check the validity of the
	// certificate, because the certificate might have expired by now.
	// However this is not enough, because the declared time cannot be
	// trusted and the certificate might have been revoked before
	// expiration: additional OCSP checks are performed for these cases,
	// either here if the profile is TM or TS, or later during signature
	// qualification.
	if c.Issuer, err = o.checkKeyInfo(c.Signer, c.SigningTime); err != nil {
		return err
	}

	sigval, err := checkSignatureValue(s, c.Signer, x509sa)
	if err != nil {
		return err
	}

	// Check that QualifyingProperties targets this Signature.
	qprop := &s.Object.QualifyingProperties
	if qprop.Target != "#"+s.ID {
		return QualifyingPropertiesTargetError{
			SignatureID: s.ID,
			Target:      qprop.Target,
		}
	}

	if err = checkSignedProperties(&qprop.SignedProperties, c.Signer, o.profile); err != nil {
		return err
	}

	usp := &qprop.UnsignedProperties.UnsignedSignatureProperties
	ocsp := &usp.RevocationValues.OCSPValues
	switch o.profile {
	case TM:
		var nonce []byte
		if c.SigningTime, nonce, err = checkOCSP(ocsp, c.Signer, c.Issuer, o.ocsp); err != nil {
			return err
		}
		err = checkNonce(nonce, sigval)
	case TS:
		if _, _, err = checkOCSP(ocsp, c.Signer, c.Issuer, o.ocsp); err != nil {
			return err
		}
		c.SigningTime, err = checkTimestamp(&usp.SignatureTimeStamp, &s.SignatureValue, o.tsp)
	}
	return err
}

// checkReferences checks the references for SignedProperties and each file.
func checkReferences(s *signature, files map[string]*asiceFile) error {
	// Check the SignedProperties reference.
	sigPropID := s.Object.QualifyingProperties.SignedProperties.ID
	ref, err := findReference(s.SignedInfo.Reference, "#"+sigPropID)
	if err != nil {
		return err
	}

	if ref.Type != "http://uri.etsi.org/01903#SignedProperties" {
		return InvalidSignedPropertiesReferenceTypeError{Type: ref.Type}
	}

	// We use a fixed canonicalization algorithm, so Transforms should
	// either be missing or designate the algorithm that we use.
	for _, tr := range ref.Transforms.Transform {
		if tr.Algorithm != xmlc14n11 {
			return UnsupportedSignedPropertiesTransformError{
				Algorithm: tr.Algorithm,
			}
		}
	}

	sigprop := buffer()
	defer release(sigprop)
	writeXML(&s.Object.QualifyingProperties.SignedProperties, sigprop)
	if err = checkXMLDigest(ref.DigestMethod, sigprop.Bytes(), ref.DigestValue); err != nil {
		return SignedPropertiesReferenceDigestError{Err: err}
	}

	// Check the data file references.
	dofs := s.Object.QualifyingProperties.SignedProperties.
		SignedDataObjectProperties.DataObjectFormat
	for _, file := range files {
		ref, err = findReference(s.SignedInfo.Reference, file.name)
		if err != nil {
			return err
		}

		if len(ref.Type) > 0 {
			return UnsupportedReferenceTypeError{URI: file.name, Type: ref.Type}
		}
		if len(ref.Transforms.Transform) > 0 {
			return FileReferenceWithTransformsError{URI: file.name}
		}

		if err = checkXMLDigest(ref.DigestMethod, file.data.Bytes(), ref.DigestValue); err != nil {
			return FileReferenceDigestError{URI: file.name, Err: err}
		}

		// Since we already have the reference for file here, also
		// check the DataObjectFormat instead of having to look up the
		// reference again later when checking SignedProperties.
		dof, err := findDataObjectFormat(dofs, "#"+ref.ID)
		if err != nil {
			return err
		}
		if dof.MimeType.Value != file.mimetype {
			return DataObjectFormatMIMETypeMismatchError{
				URI:              file.name,
				Manifest:         file.mimetype,
				DataObjectFormat: dof.MimeType.Value,
			}
		}
	}

	// Check for extra References. We know that there cannot be any less,
	// because we had a reference for each unique file name.
	if count, expected := len(s.SignedInfo.Reference), len(files)+1; count > expected {
		return ExtraReferencesError{Count: count, Expected: expected}
	}

	// Check for extra DataObjectFormats.
	if count, expected := len(dofs), len(files); count > expected {
		return ExtraDataObjectFormatsError{Count: count, Expected: expected}
	}
	return nil
}

func findReference(refs []reference, uri string) (reference, error) {
	for _, ref := range refs {
		ruri, err := url.QueryUnescape(ref.URI)
		if err != nil {
			return ref, ReferenceURIUnescapeError{
				URI: ref.URI,
				Err: err,
			}
		}
		if ruri == uri {
			return ref, nil
		}
	}
	return reference{}, NoReferenceWithURIError{URI: uri}
}

func findDataObjectFormat(dofs []dataObjectFormat, ref string) (dataObjectFormat, error) {
	for _, dof := range dofs {
		if dof.ObjectReference == ref {
			return dof, nil
		}
	}
	return dataObjectFormat{}, NoDataObjectFormatWithReferenceError{Reference: ref}
}

// checkKeyInfo verifies the signing certificate against our trusted roots.
func (o *Opener) checkKeyInfo(c *x509.Certificate, time time.Time) (issuer *x509.Certificate, err error) {
	if c.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		return nil, NotANonRepudiationCertificateError{
			KeyUsage: c.KeyUsage,
		}
	}

	opts := x509.VerifyOptions{
		Roots:         o.rpool,
		Intermediates: o.ipool,
		CurrentTime:   time,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	chains, err := c.Verify(opts)
	if err != nil {
		return nil, SignerCertificateVerificationError{
			Certificate: c.Raw, // Log entire cert for diagnostics.
			SigningTime: time,
			Err:         err,
		}
	}

	// At least one chain is guaranteed: use the first one.
	issuer = c
	if len(chains[0]) > 1 {
		issuer = chains[0][1]
	}
	return
}

func checkSignatureValue(s *signature, c *x509.Certificate, alg x509.SignatureAlgorithm) (
	signature []byte, err error) {

	if signature, err = b64d(s.SignatureValue.Value); err != nil {
		return nil, DecodeSignatureValueError{Err: err}
	}

	siginfo := buffer()
	defer release(siginfo)
	writeXML(&s.SignedInfo, siginfo)

	// We need to re-encode ECDSA signatures for the Go standard library.
	recode := signature
	switch alg {
	case x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		var r, s *big.Int
		if r, s, err = cryptoutil.ParseECDSAXMLSignature(signature); err != nil {
			return nil, ECDSASignatureParseError{Signature: signature, Err: err}
		}
		if recode, err = asn1.Marshal(struct {
			R *big.Int
			S *big.Int
		}{r, s}); err != nil {
			return nil, ECDSASignatureASN1MarshalError{Err: err}
		}
	}

	if err = c.CheckSignature(alg, siginfo.Bytes(), recode); err != nil {
		return nil, SignatureVerificationError{Err: err}
	}
	return signature, nil
}

func checkSignedProperties(p *signedProperties, c *x509.Certificate, profile Profile) error {
	// Check that SigningCertificate is correct
	sc := &p.SignedSignatureProperties.SigningCertificate
	if err := checkSigningCertificate(sc, c); err != nil {
		return err
	}

	// DataObjectFormats were already checked in checkReferences.

	spi := &p.SignedSignatureProperties.SignaturePolicyIdentifier
	switch profile {
	case TM:
		// If using BDOC-specific timemarks, then check that the
		// signature policy identifier references BDOC.
		return checkSignaturePolicyIdentifier(spi)
	default:
		// In all other cases, require that there be no signature
		// policy which might affect the interpretation of the
		// signature container.
		if spi.XMLElement.isPresent() {
			return UnexpectedSignaturePolicyIdentifierError{
				Identifier: spi.SignaturePolicyID.SigPolicyID.Identifier.Value,
			}
		}
	}
	return nil
}

func checkSigningCertificate(s *signingCertificate, c *x509.Certificate) error {
	// Check CertDigest to ensure that KeyInfo was not replaced.
	method := s.Cert.CertDigest.DigestMethod
	digest := s.Cert.CertDigest.DigestValue
	if err := checkXMLDigest(method, c.Raw, digest); err != nil {
		return SigningCertificateDigestError{Err: err}
	}

	// Check IssuerSerial to ensure that the correct certificate is indicated.
	serial := strings.TrimSpace(s.Cert.IssuerSerial.X509SerialNumber.Value)
	if c.SerialNumber.String() != serial {
		return SigningCertificateSerialError{
			KeyInfo:            c.SerialNumber,
			SigningCertificate: serial,
		}
	}

	issuer, err := cryptoutil.DecodeRDNSequence(s.Cert.IssuerSerial.X509IssuerName.Value)
	if err != nil {
		return SigningCertificateIssuerParseError{
			Issuer: s.Cert.IssuerSerial.X509IssuerName,
			Err:    err,
		}
	}

	c.Issuer.ExtraNames = c.Issuer.Names
	if !cryptoutil.RDNSequenceEqual(c.Issuer.ToRDNSequence(), issuer) {
		return SigningCertificateIssuerError{
			KeyInfo:            c.Issuer,
			SigningCertificate: s.Cert.IssuerSerial.X509IssuerName,
		}
	}
	return nil
}

var signaturePolicyIDMap = map[string]string{
	"urn:oid:1.3.6.1.4.1.10015.1000.3.2.1": "3Tl1oILSvOAWomdI9VeWV6IA/32eSXRUri9kPEz1IVs=",
	"urn:oid:1.3.6.1.4.1.10015.1000.3.2.3": "ep8Ng0NXo+hG58oWow1XVPcs191tlphvqYFU13tkb6Y=",
}

func checkSignaturePolicyIdentifier(spi *signaturePolicyIdentifier) error {
	// Determine if we are dealing with BDOC 2.1 or 2.1.2 based on URN.
	identifier := spi.SignaturePolicyID.SigPolicyID.Identifier
	if identifier.Qualifier != "OIDAsURN" {
		return SigPolicyIDQualifierError{Qualifier: identifier.Qualifier}
	}
	value := strings.TrimSpace(identifier.Value)
	digest, ok := signaturePolicyIDMap[value]
	if !ok {
		return SigPolicyIDValueError{Value: value}
	}

	// Check that the signature policy digest matches.
	hash := spi.SignaturePolicyID.SigPolicyHash
	if hash.DigestMethod.Algorithm != "http://www.w3.org/2001/04/xmlenc#sha256" {
		return SigPolicyHashMethodError{Algorithm: hash.DigestMethod.Algorithm}
	}
	if value := strings.TrimSpace(hash.DigestValue.Value); value != digest {
		return SigPolicyHashValueError{
			Value:    value,
			Expected: digest,
		}
	}
	return nil
}

func checkOCSP(values *ocspValues, c, issuer *x509.Certificate, ocsp *ocsp.Client) (
	producedAt time.Time, nonce []byte, err error) {

	value := values.EncapsulatedOCSPValue.Value
	if len(value) == 0 {
		return producedAt, nil, OCSPResponseMissingError{}
	}

	response, err := b64d(value)
	if err != nil {
		return producedAt, nil, OCSPResponseDecodeError{Value: value, Err: err}
	}

	status, err := ocsp.CheckFullResponse(response, c, issuer, nil)
	if err != nil {
		return producedAt, nil, OCSPResponceVerificationError{Err: err}
	}
	if !status.Good {
		return producedAt, nil, OCSPStatusNotGoodError{Response: *status}
	}
	return status.ProducedAt, status.Nonce, nil
}

var hashOIDMap = map[string]crypto.Hash{
	"2.16.840.1.101.3.4.2.4": crypto.SHA224,
	"2.16.840.1.101.3.4.2.1": crypto.SHA256,
	"2.16.840.1.101.3.4.2.2": crypto.SHA384,
	"2.16.840.1.101.3.4.2.3": crypto.SHA512,
}

func checkNonce(nonce []byte, signature []byte) error {
	var tdd struct { // TBSDocumentDigest
		AlgorithmIdentifier pkix.AlgorithmIdentifier
		Value               []byte
	}
	rest, err := asn1.Unmarshal(nonce, &tdd)
	if err != nil {
		return OCSPNonceASN1UnmarshalError{Err: err}
	} else if len(rest) > 0 {
		return OCSPNonceTrailingDataError{Trailing: rest}
	}

	// Hash the signature using the same digest method as the nonce.
	h, ok := hashOIDMap[tdd.AlgorithmIdentifier.Algorithm.String()]
	if !ok {
		return UnsupportedNonceDigestAlgorithmError{
			Algorithm: tdd.AlgorithmIdentifier.Algorithm,
		}
	}
	hash := h.New()
	hash.Write(signature)
	calculated := hash.Sum(nil)
	if !bytes.Equal(tdd.Value, calculated) {
		return NonceMismatchError{Nonce: tdd.Value, Calculated: calculated}
	}
	return nil
}

func checkTimestamp(timestamp *xadesTimeStamp, sigval *signatureValue, tsp *tsp.Client) (
	genTime time.Time, err error) {

	value := timestamp.EncapsulatedTimeStamp.Value
	if len(value) == 0 {
		return genTime, TimestampMissingError{}
	}

	response, err := b64d(value)
	if err != nil {
		return genTime, TimestampDecodeError{Value: value, Err: err}
	}

	data := buffer()
	defer release(data)
	writeXML(sigval, data)

	if genTime, err = tsp.Check(response, data.Bytes(), nil); err != nil {
		return genTime, TimestampVerificationError{Err: err}
	}
	return genTime, nil
}

var hashXMLMap = map[string]crypto.Hash{
	"http://www.w3.org/2001/04/xmlenc#sha256":       crypto.SHA256,
	"http://www.w3.org/2001/04/xmldsig-more#sha384": crypto.SHA384,
	"http://www.w3.org/2001/04/xmlenc#sha512":       crypto.SHA512,
}

func checkXMLDigest(method digestMethod, data []byte, digest digestValue) error {
	h, ok := hashXMLMap[method.Algorithm]
	if !ok {
		return UnsupportedDigestAlgorithm{Algorithm: method.Algorithm}
	}

	decoded, err := b64d(digest.Value)
	if err != nil {
		return DigestValueDecodeError{Err: err}
	}

	hash := h.New()
	hash.Write(data)
	calculated := hash.Sum(nil)
	if !bytes.Equal(decoded, calculated) {
		return DigestMismatchError{Expected: decoded, Calculated: calculated}
	}
	return nil
}

func parseBase64Certificate(c string) (*x509.Certificate, error) {
	decoded, err := b64d(c)
	if err != nil {
		return nil, Base64DecodeError{Err: err}
	}
	cert, err := x509.ParseCertificate(decoded)
	if err != nil {
		return nil, CertificateParseError{Err: err}
	}
	return cert, nil
}

func b64d(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.TrimSpace(data))
}
