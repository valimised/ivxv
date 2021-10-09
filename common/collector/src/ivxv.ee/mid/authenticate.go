package mid

import (
	"context"
	"crypto/rand"
	"crypto/x509"

	"ivxv.ee/cryptoutil"
)

// MobileAuthenticate starts a Mobile-ID authentication session.
func (c *Client) MobileAuthenticate(ctx context.Context, idCode, phone string) (
	sesscode string, challengeRnd []byte, challenge []byte, err error) {

	// Generate random authentication challenge to sign. Although we could
	// use the challenge directly, we pass it through the hash function to
	// simplify VerifyAuthenticationSignature which requires the pre-image
	// of the signed data.
	challengeRnd = make([]byte, c.authHashFunction.Size())
	if _, err = rand.Read(challengeRnd); err != nil {
		err = GenerateAuthenticationChallengeError{Err: err}
		return
	}
	d := c.authHashFunction.New()
	d.Write(challengeRnd)
	challenge = d.Sum(nil)

	hashType := hashFunctionNames[c.authHashFunction]

	sesscode, err = c.startSession(ctx, sessAuth, idCode, phone, challenge, hashType)
	if err != nil {
		err = MobileAuthenticateError{Err: err}
		return
	}

	return
}

// GetMobileAuthenticateStatus queries the status of a Mobile-ID authentication
// session. If err is nil and signature is empty, then the transaction is still
// outstanding. If err is nil and signature is non-nil, then the user is
// authenticated, although callers should use VerifyAuthenticationSignature to
// double-check.
func (c *Client) GetMobileAuthenticateStatus(ctx context.Context, sesscode string) (
	cert *x509.Certificate, algorithm string, signature []byte, err error) {

	var certDER []byte
	algorithm, signature, certDER, err = c.getSessionStatus(ctx, sessAuth, sesscode)
	if err != nil {
		err = GetMobileAuthenticateStatusError{Err: err}
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
		}
		return
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
			Response: status,
			Err:      err,
		}
		return
	}
	if !status.Good {
		var certerr CertificateError
		certerr.Err = AuthenticationCertificateRevokedError{
			Reason: status.RevocationReason,
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
			Algorithm: algorithm,
		}
	}

	// We need to re-encode ECDSA signatures for the Go standard library.
	switch sigalg {
	case x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		if signature, err = cryptoutil.ReEncodeECDSASignature(signature); err != nil {
			return ReEncodeAuthenticationSignatureError{
				Signature: signature,
				Err:       err,
			}
		}
	}

	if err = cert.CheckSignature(sigalg, signed, signature); err != nil {
		return VerifyAuthenticationSignatureError{Err: err}
	}
	return nil
}
