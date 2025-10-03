package webeid

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"ivxv.ee/common/collector/cryptoutil"
	"ivxv.ee/common/collector/token"
)

const (
	// Major release of a Web eID authentication token.
	// This number should match the number that is retrieved from a Web eID
	// auth token while parsing. If numbers doesn't match, then backend should
	// refuse further operations with that particular Web eID auth token, since
	// major releases are incompatible.
	majorRelease = "1"
	// Regex for Format field in Web eID token.
	formatRegexp = `^web-eid:([0-9]+).([0-9]+)$`
	// Allowed signature algorithms:
	//
	// ECDSA:
	// "ES256", "ES384", "ES512",
	//
	// RSASSA-PSS:
	// "PS256", "PS384", "PS512",
	//
	// RSASSA-PKCS1-v1_5:
	// "RS256", "RS384", "RS512",
	algorithm = `^(ES|PS|RS)(256|384|512)$`
)

// webEidAuthToken is a Web eID authentication token that client
// sends to the webeid service.
//
// Reference:
// https://github.com/web-eid/web-eid-system-architecture-doc#web-eid-authentication-token-specification
type webEidAuthToken struct {
	// UnverifiedCertificate is a base64 encoded eID user's certificate
	UnverifiedCertificate string `json:"unverifiedCertificate"`
	// Algorithm that was used to produce a Signature
	Algorithm string `json:"algorithm"`
	// Signature = Sign(Hash(origin) + Hash(nonce))
	Signature string `json:"signature"`
	// Format specifies current version of Web eID (important to be major compatible)
	Format string `json:"format"`
	// AppVersion is just an additional info (not important)
	AppVersion string `json:"appVersion"`
}

// Certify returns Web eID token UnverifiedCertificate field.
func (w *fromRaw) Certify() *x509.Certificate {
	return w.Cert
}

// Header is not implemented.
func (w *fromRaw) Header() (string, error) {
	return "", nil
}

// Payload is not implemented.
func (w *fromRaw) Payload() (string, error) {
	return "", nil
}

// Signature returns Web eID token Signature field.
func (w *fromRaw) Signature() ([]byte, error) {
	return w.Sig, nil
}

func (w *fromRaw) Verify() error {
	// Regex over Format field
	release, err := formatRegex(w.Format)
	if err != nil {
		return VerifyFormatError{Err: err,
			Description: _WEBEID_REGEX}
	}

	// Web eID auth token major release should be supported by backend
	err = formatMajorRelease(release)
	if err != nil {
		return VerifyMajorReleaseError{Err: err,
			Description: _WEBEID_VER}
	}

	// Hash origin with specified algorithm
	hashOrigin, err := hashIt(w.Origin, w.Algorithm)
	if err != nil {
		return OriginHashError{Err: err,
			Description: _WEBEID_HASH_ALG}
	}

	// Hash nonce with specified algorithm
	hashNonce, err := hashIt(w.Nonce, w.Algorithm)
	if err != nil {
		return NonceHashError{Err: err,
			Description: _WEBEID_HASH_NONCE}
	}

	// [1, 2, 3, 4, 5] = append([1, 2, 3], [4, 5])
	hashOrigin = append(hashOrigin, hashNonce...)

	sha384SigAlgo := anySignatureAlgorithmToSHA384(w.Cert.SignatureAlgorithm)

	signature := w.Sig

	// ECDSA signatures can be in 2 forms:
	// a) ASN1-encoded
	// b) Encoded as 2 params (R and S)
	// Check whether w.Sig belongs to a) type signature
	err = cryptoutil.IsECDSAASN1EncodedSignature(w.Sig)
	if err != nil {
		// w.Sig doesn't belong to a) type, but maybe it is a) type?
		ecdsaSig, err := cryptoutil.ReEncodeECDSASignature(w.Sig)
		if err == nil {
			// Yes, it is an a) type
			signature = ecdsaSig
		}
		// No, w.Sig is an RSA-based signature
	}

	// This method does 2 things:
	// 1. data = SHA384(hashOrigin)
	// 2. VerifySignature(data, signature)
	err = w.Cert.CheckSignature(sha384SigAlgo, hashOrigin, signature)
	if err != nil {
		return CheckSignatureError{Err: err,
			Description: _WEBEID_SIG}
	}

	return nil
}

func (w *fromRaw) Unmarshal() (token.Token, error) {
	// Try to unmarshal raw Web eID auth token
	authToken := webEidAuthToken{}
	err := json.Unmarshal([]byte(w.Token), &authToken)
	if err != nil {
		return nil, JSONUnmarshalError{Err: err,
			Description: _WEBEID_JSON}
	}

	// Base64 decode eID user's certificate
	cert, err := cryptoutil.Base64Certificate(authToken.UnverifiedCertificate)
	if err != nil {
		return nil, Base64DecodeUnverifiedCertificateError{Err: err,
			Description: _WEBEID_CERT_B64}
	}

	// Base64 decode eID user's signature
	signature, err := base64.StdEncoding.DecodeString(authToken.Signature)
	if err != nil {
		return nil, Base64DecodeSignatureError{Err: err,
			Description: _WEBEID_SIG_B64}
	}

	w.Sig = signature
	w.Cert = cert
	w.Format = authToken.Format
	w.Algorithm = authToken.Algorithm
	return w, nil
}
