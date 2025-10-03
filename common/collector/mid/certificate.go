package mid

import (
	"context"
	"crypto/x509"
)

// https://github.com/SK-EID/MID#313-request-parameters
type getMobileCertificate struct {
	RelyingPartyUUID       string `json:"relyingPartyUUID"`
	RelyingPartyName       string `json:"relyingPartyName"`
	PhoneNumber            string `json:"phoneNumber"`
	NationalIdentityNumber string `json:"nationalIdentityNumber"`
}

// https://github.com/SK-EID/MID#316-response-structure
type getMobileCertificateResponse struct {
	Result  string `json:"result"`
	Cert    []byte `json:"cert"`
	Time    string `json:"time"`
	TraceID string `json:"traceId"`
}

// GetMobileCertificate queries for a Mobile-ID ECC signing certificate.
func (c *Client) GetMobileCertificate(ctx context.Context, id, phone string) (
	cert *x509.Certificate, err error) {

	// We cannot use a struct literal, because gen would report it
	// as a duplicate error type.
	var input InputError
	switch {
	case len(id) == 0:
		input.Err = GetCertificateNoIDCodeError{Description: _MID_N_ID}
		err = input
	case len(phone) == 0:
		input.Err = GetCertificateNoPhoneError{Description: _MID_N_PNO}
		err = input
	}
	if err != nil {
		return
	}

	var resp getMobileCertificateResponse
	if err = httpPost(ctx, c.url+"certificate", getMobileCertificate{
		RelyingPartyUUID:       c.conf.RelyingPartyUUID,
		RelyingPartyName:       c.conf.RelyingPartyName,
		NationalIdentityNumber: id,
		PhoneNumber:            phone,
	}, &resp); err != nil {
		return nil, GetMobileCertificateError{Err: err,
			Description: _MID_HTTP_CERT}
	}

	switch resp.Result {
	case "OK":
	case "NOT_FOUND":
		var certerr CertificateError
		certerr.Err = MobileCertificateNotFoundError{Description: _MID_RESP_NFOUND}
		return nil, certerr
	case "NOT_ACTIVE":
		var certerr CertificateError
		certerr.Err = MobileCertificateNotActiveError{Description: _MID_RESP_NACTIVE}
		return nil, certerr
	default:
		var status StatusError
		status.Err = UnknownMobileCertificateResultError{Status: resp.Result,
			Description: _MID_RESP_UNKNOWN_ERR}
		return nil, status
	}

	if cert, err = x509.ParseCertificate(resp.Cert); err != nil {
		return nil, ParseMobileCertificateError{
			Certificate: resp.Cert,
			Err:         err,
			Description: _MID_CERT,
		}
	}
	return
}
