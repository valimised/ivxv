package dds

import (
	"context"
	"crypto/x509"
	"encoding/xml"

	"ivxv.ee/cryptoutil"
)

// https://sk-eid.github.io/dds-documentation/api/api_docs/#getmobilecertificate
type getMobileCertificate struct {
	XMLName        xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileCertificate"`
	IDCode         string
	PhoneNo        string
	ReturnCertData string
}

type getMobileCertificateResponse struct {
	//nolint:lll // XML namespace forces us to use a long line.
	XMLName        xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileCertificateResponse"`
	SignCertStatus string
	SignCertData   string
}

// GetMobileCertificate queries for a Mobile-ID ECC signing certificate.
func (c *Client) GetMobileCertificate(ctx context.Context, id, phone string) (
	cert *x509.Certificate, err error) {

	var resp getMobileCertificateResponse
	if err = soapRequest(ctx, c.conf.URL, getMobileCertificate{
		IDCode:         id,
		PhoneNo:        phone,
		ReturnCertData: "signECC",
	}, &resp); err != nil {
		return nil, GetMobileCertificateError{Err: err}
	}

	switch resp.SignCertStatus {
	case "OK":
	case "REVOKED":
		var certerr CertificateError
		certerr.Err = MobileCertificateRevokedError{}
		return nil, certerr
	default:
		var status StatusError
		status.Err = UnknownMobileCertificateStatusError{Status: resp.SignCertStatus}
		return nil, status
	}

	if cert, err = cryptoutil.PEMCertificate(resp.SignCertData); err != nil {
		return nil, ParseMobileCertificateError{
			Certificate: resp.SignCertData,
			Err:         err,
		}
	}
	return
}
