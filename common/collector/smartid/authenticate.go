package smartid

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
)

var (
	// https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.12, SK specific OID for
	// Smart-ID authentication
	idKpClientAuthSK = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 62306, 5, 7, 0}
)

// Challenge generates a Smart-ID authentication session's challenge. Challenge returns a challenge and a
// verification code.
func (c *Client) Challenge() ([]byte, []byte, error) {
	challenge := make([]byte, c.authHashFunction.Size())
	if _, err := rand.Read(challenge); err != nil {
		return nil, nil, ChallengeError{Err: err, Description: _SID_RAND}
	}

	d := c.authHashFunction.New()
	d.Write(challenge)

	return challenge, d.Sum(nil), nil
}

// Authenticate starts a Smart-ID authentication session.
func (c *Client) Authenticate(ctx context.Context, identifer string, challengeRnd []byte) (
	sesscode string, err error) {

	d := c.authHashFunction.New()
	d.Write(challengeRnd)
	challenge := d.Sum(nil)

	hashType := hashFunctionNames[c.authHashFunction]

	sesscode, err = c.startSession(ctx, sessAuth, convertToETSI(identifer), challenge, hashType)
	if err != nil {
		err = AuthenticateError{Err: err, Description: _SID_SESS}
		return
	}

	return
}

// GetAuthenticateStatus queries the status of a Smart-ID authentication
// session. If err is nil and signature is empty, then the transaction is still
// outstanding. If err is nil and signature is non-nil, then the user is
// authenticated, although callers should use VerifyAuthenticationSignature to
// double-check.
func (c *Client) GetAuthenticateStatus(ctx context.Context, sesscode string) (
	documentno string, cert *x509.Certificate, algorithm string, signature []byte, err error) {

	var certDER []byte
	documentno, algorithm, signature, certDER, err = c.getSessionStatus(ctx, sesscode)
	if err != nil {
		err = GetAuthenticateStatusError{Err: err, Description: _SID_SESS_STAT}
		return
	}

	if certDER != nil {
		cert, err = c.parseAndVerify(ctx, certDER)
	}

	return
}

// parseAndVerify is a helper function to parse and verify the authentication
// certificate.
func (c *Client) parseAndVerify(ctx context.Context, certDER []byte) (
	cert *x509.Certificate, err error) {
	// Parse the authentication certificate.
	if cert, err = x509.ParseCertificate(certDER); err != nil {
		err = ParseAuthenticationCertificateError{
			Certificate: certDER,
			Err:         err,
			Description: _SID_CERT,
		}
		return
	}

	if cert.UnknownExtKeyUsage != nil {
		// We only expect one extended key usage, which is either standard id-kp-clientAuth
		// or id-kp-clientAuthSK, the latter must be marked as unknown to the RFC5280
		if cert.ExtKeyUsage != nil {
			err = UnexpectedExtKeyUsageError{
				Certificate: cert,
				ExtKeyUsage: cert.ExtKeyUsage[0],
				Description: _SID_EXTKEY_SK,
			}
			return
		}

		cert.ExtKeyUsage = make([]x509.ExtKeyUsage, len(cert.UnknownExtKeyUsage))
	}

	for i, ext := range cert.UnknownExtKeyUsage {
		switch ext.String() {
		case idKpClientAuthSK.String():
			cert.UnknownExtKeyUsage[i] = nil
			cert.ExtKeyUsage[i] = x509.ExtKeyUsageClientAuth
		default:
			err = UnknownExtKeyUsageError{
				Certificate: cert,
				ExtKeyUsage: ext.String(),
				Description: _SID_EXTKEY,
			}
			return
		}
	}

	// Verify the authentication certificate and get the issuer.
	opts := x509.VerifyOptions{
		Roots:         c.rpool,
		Intermediates: c.ipool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	chains, err := cert.Verify(opts)
	if err != nil {
		var certerr CertificateError
		certerr.Err = AuthenticationCertificateVerificationError{
			Certificate: cert,
			Err:         err,
			Description: _SID_CERT_VERIFY,
		}
		err = certerr
		return
	}
	issuer := cert
	if len(chains[0]) > 1 { // At least one chain is guaranteed.
		issuer = chains[0][1]
	}

	// Check OCSP status.
	status, err := c.ocsp.Check(ctx, cert, issuer, nil)
	if err != nil {
		err = CheckAuthenticationCertOCSPResponsError{
			Response:    status,
			Err:         err,
			Description: _SID_OCSP,
		}
		return
	}
	if !status.Good {
		var certerr CertificateError
		certerr.Err = AuthenticationCertificateRevokedError{
			Reason:      status.RevocationReason,
			Description: _SID_OCSP_N_GOOD,
		}
		err = certerr
		return
	}

	return
}

// VerifyAuthenticationSignature verifies the certificate signature on the
// authentication challenge.
func VerifyAuthenticationSignature(cert *x509.Certificate, algorithm string,
	signed, signature []byte) (err error) {

	sigalg, ok := signatureAlgs[algorithm]
	if !ok {
		return SigAlgorithmNotSupportedError{
			Algorithm:   algorithm,
			Description: _SID_ALG,
		}
	}

	if err = cert.CheckSignature(sigalg, signed, signature); err != nil {
		return VerifyAuthenticationSignatureError{Err: err,
			Description: _SID_SIG}
	}
	return nil
}

// convertToETSI is a helper function to make identifier to ETSI identifier.
func convertToETSI(identifier string) string {
	return "PNOEE-" + identifier
}
