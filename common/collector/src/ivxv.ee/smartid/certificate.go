package smartid

import (
	"context"
	"crypto/x509"
)

// https://github.com/SK-EID/smart-id-documentation#2384-request-parameters
type getCertificate struct {
	RelyingPartyUUID string   `json:"relyingPartyUUID"`
	RelyingPartyName string   `json:"relyingPartyName"`
	CertificateLevel string   `json:"certificateLevel,omitempty"`
	Nonce            string   `json:"nonce,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
}

// GetCertificateChoice  starts a Smart-ID certificate choice session.
func (c *Client) GetCertificateChoice(ctx context.Context, identifier string) (sess string, err error) {

	// We cannot use a struct literal, because gen would report it
	// as a duplicate error type.
	var input InputError

	if len(identifier) == 0 {
		input.Err = GetCertificateNoIDCodeError{}
		err = input
	}
	if err != nil {
		return
	}

	var resp startSessionResponse
	if err = httpPost(ctx, c.url+"certificatechoice/etsi/"+convertToETSI(identifier), getCertificate{
		RelyingPartyUUID: c.conf.RelyingPartyUUID,
		RelyingPartyName: c.conf.RelyingPartyName,
		CertificateLevel: c.conf.CertificateLevel,
	}, &resp); err != nil {
		return "", GetMobileCertificateError{Err: err}
	}

	return resp.SessionID, nil
}

// GetCertificateChoiceStatus queries for a Smart-ID certificate choice session.
func (c *Client) GetCertificateChoiceStatus(ctx context.Context, sesscode string) (
	documentno string, cert *x509.Certificate, err error) {

	var certDER []byte
	documentno, _, _, certDER, err = c.getSessionStatus(ctx, sesscode)
	if err != nil {
		err = GetMobileCertificateStatusError{Err: err}
		return
	}

	if certDER != nil {
		cert, err = c.parseAndVerify(ctx, certDER)
	}

	return
}
