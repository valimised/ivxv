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
	"encoding/xml"
	"io"
	"math/big"
	"net/url"
	"strings"
	"time"

	// Import supported hash functions for calculating TM nonce.
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"

	"ivxv.ee/container"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
	"ivxv.ee/safereader"
	"ivxv.ee/yaml"
)

func init() {
	container.Register(container.BDOC, configure, unverifiedOpen)
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

func unverifiedOpen(encoded io.Reader) (c container.Container, err error) {
	// Due to the use-case of unverifiedOpen we expect it to only be called
	// once, so there is no need to set up long-lived safereaders and we
	// can use a single transient one. The bootstrap container and all
	// configuration files in it should fit inside 10 MiB.
	sr := safereader.New(10 * 1024 * 1024)
	o := &Opener{zipsr: sr, filesr: sr}

	files, err := o.openASiCE(encoded, false)
	if err != nil {
		return nil, UnverifiedOpenError{Err: err}
	}
	b := o.bdoc()
	for _, file := range files {
		// file.data will be recovered in BDOC.Close.
		b.files[file.name] = file.data
	}
	return b, nil
}

// Conf contains the configurable options for the BDOC container opener. It
// only contains serialized values such that it can easily be unmarshaled from
// a file.
type Conf struct {
	// BDOCSize is the maximum accepted size of BDOC containers read from
	// streams into memory. Note that this limit does not apply to sources
	// which already support random access.
	BDOCSize uint64

	// FileSize is the maximum accepted decompressed size of files in the container.
	FileSize uint64

	// Roots contains the root certificates used for verification in PEM format
	Roots []string

	// Roots contains the intermediate certificates used for verification in PEM format
	Intermediates []string

	// CheckTimeMark instructs the BDOC package to check the timestamp in the BDOC
	CheckTimeMark bool

	// OCSP is the configuration for the ocsp package
	OCSP ocsp.Conf
}

// Opener is a configured BDOC container opener.
type Opener struct {
	zipsr   *safereader.SafeReader
	filesr  *safereader.SafeReader
	rpool   *x509.CertPool
	ipool   *x509.CertPool
	ocsp    *ocsp.Client
	checkTM bool
}

// New returns a new BDOC container opener.
func New(c *Conf) (o *Opener, err error) {
	o = &Opener{checkTM: c.CheckTimeMark}

	if c.BDOCSize == 0 {
		return nil, UnconfiguredBDOCSizeError{}
	}
	o.zipsr = safereader.New(c.BDOCSize)

	if c.FileSize == 0 {
		return nil, UnconfiguredFileSizeError{}
	}
	o.filesr = safereader.New(c.FileSize)

	if len(c.Roots) == 0 {
		return nil, UnconfiguredRootsError{}
	}
	if o.rpool, err = cryptoutil.PEMCertificatePool(c.Roots...); err != nil {
		return nil, RootsParseError{Err: err}
	}
	if o.ipool, err = cryptoutil.PEMCertificatePool(c.Intermediates...); err != nil {
		return nil, IntermediatesParseError{Err: err}
	}

	// Create OCSP client if CheckTimeMark is true.
	if c.CheckTimeMark {
		if o.ocsp, err = ocsp.New(&c.OCSP); err != nil {
			return nil, OCSPClientError{Err: err}
		}
	}
	return
}

// Namespaces used by the BDOC signature XML files. The namespaces have
// to be sorted by name, in ascending order.
var bdocNS = []xmlns{
	{name: "asic", url: "http://uri.etsi.org/02918/v1.2.1#"},
	{name: "ds", url: "http://www.w3.org/2000/09/xmldsig#"},
	{name: "xades", url: "http://uri.etsi.org/01903/v1.3.2#"},
}

// BDOC is a parsed BDOC container
type BDOC struct {
	opener      *Opener
	signerInfos []container.Signature
	signatures  []*xadesSignatures
	files       map[string][]byte
}

func (o *Opener) bdoc() *BDOC {
	return &BDOC{opener: o, files: make(map[string][]byte)}
}

// Open opens and verifies a BDOC signature container.
func (o *Opener) Open(encoded io.Reader) (b *BDOC, err error) {
	// Open the container and get all the files from it.
	files, err := o.openASiCE(encoded, true)
	if err != nil {
		return nil, OpenBDOCContainerError{Err: err}
	}

	b = o.bdoc()
	defer func(b *BDOC) {
		// Recover all data on error.
		if err != nil {
			// Recover unprocessed ASiC-E files.
			for _, file := range files {
				o.filesr.Recover(file.data)
			}
			// nolint: errcheck, always returns nil.
			b.Close() // Recover parsed signatures.
		}
	}(b)

	// Parse the signature files from the container.
	for key, file := range files {
		if !file.signature() {
			continue
		}

		// FIXME: non-strict XML parsing.
		var s xadesSignatures
		if err = xml.Unmarshal(file.data, &s); err != nil {
			return nil, SignatureXMLError{Name: file.name, Err: err}
		}

		// file.data will be recovered in BDOC.Close.
		// XXX: We would not need to retain file.data here, if we would
		// store enough extra data during unmarshaling to reconstruct
		// the original XML for c14n.
		s.fileName = file.name
		s.raw = file.data

		// Parse certificate
		if s.Signature.KeyInfo.X509Data.parsedCert, err = parseBase64Certificate(
			s.Signature.KeyInfo.X509Data.X509Certificate); err != nil {

			return nil, SignerCertificateParseError{
				FileName:    file.name,
				SignatureID: s.Signature.ID,
				Err:         err,
			}
		}

		// Get signing time. If checkTM is true, then this will get
		// replaced by the OCSP response time.
		signingTime := s.Signature.Object.QualifyingProperties.SignedProperties.
			SignedSignatureProperties.SigningTime

		// The issuer certificate will be determined and inserted into
		// the container.Signature during certificate checking.
		b.signerInfos = append(b.signerInfos, container.Signature{
			ID:          s.Signature.ID,
			Signer:      s.Signature.KeyInfo.X509Data.parsedCert,
			SigningTime: signingTime,
		})

		delete(files, key) // Remove signatures so only data files remain.
		b.signatures = append(b.signatures, &s)
	}

	// Check the parsed signatures against the data files in the container.
	for i, s := range b.signatures {
		if err = b.check(s, &b.signerInfos[i], files); err != nil {
			return nil, CheckSignatureError{SignatureFile: s.fileName, Err: err}
		}
	}

	// Convert remaining data files from asiceFiles to byte slices as
	// expected by the container.Container interface.
	for _, file := range files {
		// file.data will be recovered in BDOC.Close.
		b.files[file.name] = file.data
	}
	return
}

// Signatures implements the container.Container interface.
func (b *BDOC) Signatures() []container.Signature {
	return b.signerInfos
}

// Data implements the container.Container interface.
func (b *BDOC) Data() map[string][]byte {
	return b.files
}

// Close implements the container.Container interface.
func (b *BDOC) Close() error {
	for _, f := range b.files {
		b.opener.filesr.Recover(f)
	}
	b.files = nil
	for _, s := range b.signatures {
		b.opener.filesr.Recover(s.raw)
	}
	b.signatures = nil
	return nil
}

// SignatureValue implements the ivxv.ee/q11n/ocsp.SignatureValuer interface.
// It returns the raw signature value for the requested ID.
func (b *BDOC) SignatureValue(id string) ([]byte, error) {
	for _, signature := range b.signatures {
		if signature.Signature.ID == id {
			encoding := base64.StdEncoding
			encoded := signature.Signature.SignatureValue.Value

			value := make([]byte, encoding.DecodedLen(len(encoded)))
			n, err := encoding.Decode(value, encoded)
			if err != nil {
				return nil, SignatureValueDecodeError{Err: err}
			}
			return value[:n], nil
		}
	}
	return nil, SignatureValueNoSuchIDError{ID: id}
}

// TimestampData implements the ivxv.ee/q11n/tsp.TimestampDataer interface. It
// returns the canonical SignatureValue XML element for the requested ID.
func (b *BDOC) TimestampData(id string) ([]byte, error) {
	for _, signature := range b.signatures {
		if signature.Signature.ID == id {
			data, err := canonicalize(string(signature.raw),
				"ds:SignatureValue", bdocNS)
			if err != nil {
				return nil, SignatureValueCanonicalizationError{Err: err}
			}
			return data, nil
		}
	}
	return nil, TimestampDataNoSuchIDError{ID: id}
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

func (b *BDOC) check(x *xadesSignatures, c *container.Signature, files map[string]*asiceFile) (err error) {
	c14n := x.Signature.SignedInfo.CanonicalizationMethod.Algorithm
	if c14n != xmlc14n11 {
		return UnsupportedCanonicalizationAlgorithmError{Algorithm: c14n}
	}

	signing := x.Signature.SignedInfo.SignatureMethod.Algorithm
	x509sa, ok := signAlgorithms[signing]
	if !ok {
		return UnsupportedSigningAlgorithmError{Algorithm: signing}
	}

	if err = checkReferences(x, files); err != nil {
		return
	}

	// We use the declared signing time here to check the validity of the
	// certificate, because the certificate might have expired by now.
	// However this is not enough, because the declared time cannot be
	// trusted and the certificate might have been revoked before
	// expiration: additional OCSP checks are performed for these cases,
	// either here if checkTM is set or later during signature
	// qualification.
	cert := x.Signature.KeyInfo.X509Data.parsedCert
	time := x.Signature.Object.QualifyingProperties.
		SignedProperties.SignedSignatureProperties.SigningTime
	if c.Issuer, err = b.checkKeyInfo(cert, time); err != nil {
		return
	}

	value, err := base64.StdEncoding.DecodeString(string(x.Signature.SignatureValue.Value))
	if err != nil {
		return DecodeSignatureValueError{Err: err}
	}
	if err = checkSignatureValue(cert, x509sa, string(x.raw), value); err != nil {
		return
	}

	// Check that QualifyingProperties targets this Signature.
	if len(x.Signature.ID) == 0 {
		return NoSignatureIDError{}
	}
	target := x.Signature.Object.QualifyingProperties.Target
	if target != "#"+x.Signature.ID {
		return QualifyingPropertiesTargetError{
			SignatureID: x.Signature.ID,
			Target:      target,
		}
	}

	sigprop := &x.Signature.Object.QualifyingProperties.SignedProperties
	if err = checkSignedProperties(sigprop, cert, b.opener.checkTM); err != nil {
		return
	}

	if b.opener.checkTM {
		ocsp := &x.Signature.Object.QualifyingProperties.
			UnsignedProperties.UnsignedSignatureProperties.RevocationValues.OCSPValues
		if c.SigningTime, err = checkOCSP(ocsp, cert, c.Issuer, value, b.opener.ocsp); err != nil {
			return err
		}
	}
	return
}

// checkReferences checks the references for SignedProperties and each file.
func checkReferences(x *xadesSignatures, files map[string]*asiceFile) (err error) {
	// Check the SignedProperties reference.
	sigPropID := x.Signature.Object.QualifyingProperties.SignedProperties.ID
	if len(sigPropID) == 0 {
		return NoSignedPropertiesIDError{}
	}
	ref, err := findReference(x.Signature.SignedInfo.Reference, "#"+sigPropID)
	if err != nil {
		return
	}

	if ref.Type != "http://uri.etsi.org/01903#SignedProperties" {
		return InvalidSignedPropertiesReferenceTypeError{Type: ref.Type}
	}

	// We use a fixed canonicalization algorithm, so Transforms should
	// either be missing or designate the algorithm that we use.
	for _, tr := range ref.Transforms {
		if tr.Algorithm != xmlc14n11 {
			return UnsupportedSignedPropertiesTransformError{
				Algorithm: tr.Algorithm,
			}
		}
	}
	sigprop, err := canonicalize(string(x.raw), "xades:SignedProperties", bdocNS)
	if err != nil {
		return SignedPropertiesCanonicalizationError{Err: err}
	}
	if err = checkXMLDigest(ref.DigestMethod, sigprop, ref.DigestValue); err != nil {
		return SignedPropertiesReferenceDigestError{Err: err}
	}

	// Check the data file references.
	seen := make(map[string]struct{})
	for _, file := range files {
		ref, err = findReference(x.Signature.SignedInfo.Reference, file.name)
		if err != nil {
			return
		}

		if len(ref.Type) > 0 {
			return UnsupportedReferenceTypeError{URI: file.name, Type: ref.Type}
		}
		if len(ref.Transforms) > 0 {
			return FileReferenceWithTransformsError{URI: file.name}
		}

		if err = checkXMLDigest(ref.DigestMethod, file.data, ref.DigestValue); err != nil {
			return FileReferenceDigestError{URI: file.name, Err: err}
		}

		// Since we already have the reference for file here, also
		// check the DataObjectFormat instead of having to look up the
		// reference again later when checking SignedProperties.
		// References must have an ID for DataObjectFormats to target.
		if len(ref.ID) == 0 {
			return NoFileReferenceIDError{URI: file.name}
		}
		if _, ok := seen[ref.ID]; ok {
			return DuplicateReferenceIDError{ID: ref.ID}
		}
		seen[ref.ID] = struct{}{}

		dof, err := findDataObjectFormat(x.Signature.Object.QualifyingProperties.
			SignedProperties.SignedDataObjectProperties.DataObjectFormat,
			"#"+ref.ID)
		if err != nil {
			return err
		}
		if dof.MimeType != file.mimetype {
			return DataObjectFormatMIMETypeMismatchError{
				URI:              file.name,
				Manifest:         file.mimetype,
				DataObjectFormat: dof.MimeType,
			}
		}
	}

	// Check for extra References. We know that there cannot be any less,
	// because we had a reference for each unique file name.
	if count, expected := len(x.Signature.SignedInfo.Reference), len(files)+1; count > expected {
		return ExtraReferencesError{Count: count, Expected: expected}
	}

	// Check for extra DataObjectFormats.
	count := len(x.Signature.Object.QualifyingProperties.
		SignedProperties.SignedDataObjectProperties.DataObjectFormat)

	if count > len(files) {
		return ExtraDataObjectFormatsError{Count: count, Expected: len(files)}
	}
	return
}

func findReference(refs []reference, uri string) (ref reference, err error) {
	for _, ref = range refs {
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
	return ref, NoReferenceWithURIError{URI: uri}
}

func findDataObjectFormat(dofs []dataObjectFormat, ref string) (dof dataObjectFormat, err error) {
	for _, dof = range dofs {
		if dof.ObjectReference == ref {
			return dof, nil
		}
	}
	return dof, NoDataObjectFormatWithReferenceError{Reference: ref}
}

// checkKeyInfo verifies the signing certificate against our trusted roots.
func (b *BDOC) checkKeyInfo(c *x509.Certificate, time time.Time) (issuer *x509.Certificate, err error) {
	if c.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		return nil, NotANonRepudiationCertificateError{
			KeyUsage: c.KeyUsage,
		}
	}

	opts := x509.VerifyOptions{
		Roots:         b.opener.rpool,
		Intermediates: b.opener.ipool,
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

func checkSignatureValue(c *x509.Certificate, alg x509.SignatureAlgorithm,
	xml string, sig []byte) (err error) {

	signedInfo, err := canonicalize(xml, "ds:SignedInfo", bdocNS)
	if err != nil {
		return CanonicalizeSignedInfoError{Err: err}
	}

	// We need to re-encode ECDSA signatures for the Go standard library.
	switch alg {
	case x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		var r, s *big.Int
		if r, s, err = cryptoutil.ParseECDSAXMLSignature(sig); err != nil {
			return ECDSASignatureParseError{Signature: sig, Err: err}
		}
		if sig, err = asn1.Marshal(struct {
			R *big.Int
			S *big.Int
		}{r, s}); err != nil {
			return ECDSASignatureASN1MarshalError{Err: err}
		}
	}

	if err = c.CheckSignature(alg, signedInfo, sig); err != nil {
		return SignatureVerificationError{Err: err}
	}
	return
}

func checkSignedProperties(p *signedProperties, c *x509.Certificate, checkTM bool) (err error) {
	// Check that SigningCertificate is correct
	sc := &p.SignedSignatureProperties.SigningCertificate
	if err = checkSigningCertificate(sc, c); err != nil {
		return
	}

	// DataObjectFormats were already checked in checkReferences.

	if checkTM {
		// If using BDOC-specific time marks, then check that the
		// signature policy identifier references BDOC.
		spi := &p.SignedSignatureProperties.SignaturePolicyIdentifier
		err = checkSignaturePolicyIdentifier(spi)
	}
	return
}

func checkSigningCertificate(s *signingCertificate, c *x509.Certificate) (err error) {
	// Check CertDigest to ensure that KeyInfo was not replaced.
	method := s.Cert.CertDigest.DigestMethod
	digest := s.Cert.CertDigest.DigestValue
	err = checkXMLDigest(method, c.Raw, digest)
	if err != nil {
		return SigningCertificateDigestError{Err: err}
	}

	// Check IssuerSerial to ensure that the correct certificate is indicated.
	serial := strings.TrimSpace(s.Cert.IssuerSerial.X509SerialNumber)
	if c.SerialNumber.String() != serial {
		return SigningCertificateSerialError{
			KeyInfo:            c.SerialNumber,
			SigningCertificate: serial,
		}
	}

	issuer, err := cryptoutil.DecodeRDNSequence(s.Cert.IssuerSerial.X509IssuerName)
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
	return
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
	if value := strings.TrimSpace(hash.DigestValue); value != digest {
		return SigPolicyHashValueError{
			Value:    value,
			Expected: digest,
		}
	}
	return nil
}

func checkOCSP(values *ocspValues, c, issuer *x509.Certificate, sig []byte, ocsp *ocsp.Client) (
	producedAt time.Time, err error) {

	// Check if the OCSP response status is good.
	value := values.EncapsulatedOCSPValue.Value
	if len(value) == 0 {
		return producedAt, OCSPResponseMissingError{}
	}

	response, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return producedAt, OCSPResponseDecodeError{Value: value, Err: err}
	}

	status, err := ocsp.CheckFullResponse(response, c, issuer, nil)
	if err != nil {
		return producedAt, OCSPResponceVerificationError{Err: err}
	}
	if !status.Good {
		return producedAt, OCSPStatusNotGoodError{Response: *status}
	}
	producedAt = status.ProducedAt

	// Check if the OCSP nonce is a digest of the signature (a time mark).
	err = checkNonce(status.Nonce, sig)
	return
}

var hashOIDMap = map[string]crypto.Hash{
	"1.3.14.3.2.26":          crypto.SHA1,
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
	hash.Write(signature) // nolint: errcheck, hash.Write never returns an error
	calculated := hash.Sum(nil)
	if !bytes.Equal(tdd.Value, calculated) {
		return NonceMismatchError{Nonce: tdd.Value, Calculated: calculated}
	}
	return nil
}

var hashXMLMap = map[string]crypto.Hash{
	"http://www.w3.org/2001/04/xmlenc#sha256":       crypto.SHA256,
	"http://www.w3.org/2001/04/xmldsig-more#sha384": crypto.SHA384,
	"http://www.w3.org/2001/04/xmlenc#sha512":       crypto.SHA512,
}

func checkXMLDigest(method digestMethod, data []byte, digest string) (err error) {
	decoded, err := base64.StdEncoding.DecodeString(digest)
	if err != nil {
		return DigestValueDecodeError{Err: err}
	}

	h, ok := hashXMLMap[method.Algorithm]
	if !ok {
		return UnsupportedDigestAlgorithm{Algorithm: method.Algorithm}
	}

	hash := h.New()
	hash.Write(data) // nolint: errcheck, hash.Write never returns an error
	calculated := hash.Sum(nil)
	if !bytes.Equal(decoded, calculated) {
		return DigestMismatchError{Expected: decoded, Calculated: calculated}
	}
	return
}

func parseBase64Certificate(c string) (cert *x509.Certificate, err error) {
	decoded, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return nil, Base64DecodeError{Err: err}
	}
	cert, err = x509.ParseCertificate(decoded)
	if err != nil {
		return nil, CertificateParseError{Err: err}
	}
	return
}
