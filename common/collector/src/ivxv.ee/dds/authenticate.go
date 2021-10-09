package dds

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"strconv"
	"strings"

	"ivxv.ee/cryptoutil"
)

// https://sk-eid.github.io/dds-documentation/api/api_docs/#mobileauthenticate
type mobileAuthenticate struct {
	XMLName              xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl MobileAuthenticate"`
	IDCode               string   `xml:",omitempty"`
	CountryCode          string
	PhoneNo              string `xml:",omitempty"`
	Language             string
	ServiceName          string
	MessageToDisplay     string
	SPChallenge          string
	MessagingMode        string
	ReturnCertData       bool
	ReturnRevocationData bool
}

type mobileAuthenticateResponse struct {
	//nolint:lll // XML namespace forces us to use a long line.
	XMLName         xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl MobileAuthenticateResponse"`
	Sesscode        int
	Status          string
	CertificateData string
	ChallengeID     string
	Challenge       string
	RevocationData  string
}

// MobileAuthenticate starts a Mobile-ID authentication session.
func (c *Client) MobileAuthenticate(ctx context.Context, idCode, phone string) (
	sesscode, challengeID string, challenge []byte, cert *x509.Certificate, err error) {

	// We cannot use a struct literal, because gen would report it
	// as a duplicate error type.
	var input InputError
	switch {
	case c.conf.IDCodeRequired && len(idCode) == 0:
		input.Err = MobileAuthenticateNoIDCodeError{}
		err = input
		return
	case c.conf.PhoneRequired && len(phone) == 0:
		input.Err = MobileAuthenticateNoPhoneError{}
		err = input
		return
	}

	// Generate our part of the challenge to sign.
	spBytes := make([]byte, 10)
	if _, err = rand.Read(spBytes); err != nil {
		err = GenerateAuthenticationChallengeError{Err: err}
		return
	}
	spChallenge := hex.EncodeToString(spBytes)

	var resp mobileAuthenticateResponse
	if err = soapRequest(ctx, c.conf.URL, mobileAuthenticate{
		IDCode:               idCode,
		CountryCode:          c.conf.CountryCode,
		PhoneNo:              phone,
		Language:             c.conf.Language,
		ServiceName:          c.conf.ServiceName,
		MessageToDisplay:     c.conf.AuthMessage,
		SPChallenge:          spChallenge,
		MessagingMode:        "asynchClientServer",
		ReturnCertData:       true,
		ReturnRevocationData: true,
	}, &resp); err != nil {
		err = MobileAuthenticateError{Err: err}
		return
	}

	if resp.Status != "OK" {
		err = MobileAuthenticateStatusError{Status: resp.Status}
		return
	}

	if !strings.HasPrefix(strings.ToLower(resp.Challenge), strings.ToLower(spChallenge)) {
		err = MobileAuthenticateChallengeMismatch{
			Prefix:    spChallenge,
			Challenge: resp.Challenge,
		}
		return
	}

	// The returned certificate is not PEM, just Base64-encoded.
	certDER, err := base64.StdEncoding.DecodeString(resp.CertificateData)
	if err != nil {
		err = DecodeAuthenticationCertificateError{
			Certificate: resp.CertificateData,
			Err:         err,
		}
		return
	}
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

	// Check the certificate revocation data, again Base64-encoded.
	ocspDER, err := base64.StdEncoding.DecodeString(resp.RevocationData)
	if err != nil {
		err = DecodeAuthenticationOCSPResponseError{
			Response: resp.RevocationData,
			Err:      err,
		}
		return
	}
	status, err := c.ocsp.CheckFullResponse(ocspDER, cert, issuer, nil)
	if err != nil {
		err = CheckAuthenticationOCSPResponseError{
			Response: ocspDER,
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

	// The challenge however, is represented in hexadecimal.
	if challenge, err = hex.DecodeString(resp.Challenge); err != nil {
		err = DecodeChallengeError{Challenge: resp.Challenge, Err: err}
		return
	}

	// DigiDocService uses integers for session codes during
	// authentication, but strings during signing. Try to make this more
	// consistent by converting authentication codes to strings too.
	sesscode = strconv.Itoa(resp.Sesscode)
	challengeID = resp.ChallengeID
	return
}

// https://sk-eid.github.io/dds-documentation/api/api_docs/#getmobileauthenticatestatus
type getMobileAuthenticateStatus struct {
	//nolint:lll // XML namespace forces us to use a long line.
	XMLName       xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileAuthenticateStatus"`
	Sesscode      int
	WaitSignature bool
}

type getMobileAuthenticateStatusResponse struct {
	//nolint:lll // XML namespace forces us to use a long line.
	XMLName   xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileAuthenticateStatusResponse"`
	Status    string
	Signature string
}

// GetMobileAuthenticateStatus queries the status of a Mobile-ID authentication
// session. If err is nil and signature is empty, then the transaction is still
// outstanding. If err is nil and signature is non-nil, then the user is
// authenticated, although callers should use VerifyAuthenticationSignature to
// double-check.
func (c *Client) GetMobileAuthenticateStatus(ctx context.Context, sesscode string) (
	signature []byte, err error) {

	sessInt, err := strconv.Atoi(sesscode)
	if err != nil {
		// We cannot use a struct literal, because gen would report it
		// as a duplicate error type.
		var input InputError
		input.Err = ParseSesscodeError{Sesscode: sesscode, Err: err}
		return nil, input
	}

	var resp getMobileAuthenticateStatusResponse
	if err = soapRequest(ctx, c.conf.URL, getMobileAuthenticateStatus{
		Sesscode:      sessInt,
		WaitSignature: false,
	}, &resp); err != nil {
		return nil, GetMobileAuthenticateStatusError{Err: err}
	}

	switch resp.Status {
	case "OUTSTANDING_TRANSACTION":
	case "USER_AUTHENTICATED":
		signature, err = base64.StdEncoding.DecodeString(resp.Signature)
		if err != nil {
			return nil, DecodeChallengeSignatureError{
				Signature: resp.Signature,
				Err:       err,
			}
		}
	case "EXPIRED_TRANSACTION":
		var expired ExpiredError
		err = expired
	case "USER_CANCEL":
		var canceled CanceledError
		err = canceled
	case "PHONE_ABSENT":
		var absent AbsentError
		err = absent
	default:
		var status StatusError
		status.Err = UnexpectedAuthenticationStatusError{Status: resp.Status}
		err = status
	}
	return
}

// VerifyAuthenticationSignature verifies the certificate signature on the
// authentication challenge.
func VerifyAuthenticationSignature(cert *x509.Certificate, challenge, signature []byte) error {
	key, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return VerifyUnsupportedAlgorithmError{Algorithm: cert.PublicKeyAlgorithm}
	}
	r, s, err := cryptoutil.ParseECDSAXMLSignature(signature)
	if err != nil {
		return ParseAuthenticationSignatureError{
			Signature: signature,
			Err:       err,
		}
	}
	if !ecdsa.Verify(key, challenge, r, s) {
		return VerifyAuthenticationSignatureError{}
	}
	return nil
}
