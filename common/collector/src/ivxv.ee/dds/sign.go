package dds

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
)

// https://sk-eid.github.io/dds-documentation/api/api_docs/#mobilesignhash
type mobileSignHashRequest struct {
	XMLName          xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl MobileSignHashRequest"`
	IDCode           string
	PhoneNo          string
	Language         string
	ServiceName      string
	MessageToDisplay string
	Hash             string
	HashType         string
	KeyID            string
}

type mobileSignHashResponse struct {
	XMLName     xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl MobileSignHashResponse"`
	Sesscode    string
	ChallengeID string
	Status      string
}

// MobileSignHash starts a Mobile-ID signing session to sigh hash. The hash
// method used must be SHA-256.
func (c *Client) MobileSignHash(ctx context.Context, id, phone string, hash []byte) (
	sesscode, challengeID string, err error) {

	var resp mobileSignHashResponse
	if err = soapRequest(ctx, c.conf.URL, mobileSignHashRequest{
		IDCode:           id,
		PhoneNo:          phone,
		Language:         c.conf.Language,
		ServiceName:      c.conf.ServiceName,
		MessageToDisplay: c.conf.SignMessage,
		Hash:             hex.EncodeToString(hash),
		HashType:         "SHA256",
		KeyID:            "ECC",
	}, &resp); err != nil {
		err = MobileSignHashError{Err: err}
		return
	}

	if resp.Status != "OK" {
		err = MobileSignHashStatusError{Status: resp.Status}
		return
	}

	return resp.Sesscode, resp.ChallengeID, nil
}

// https://sk-eid.github.io/dds-documentation/api/api_docs/#getmobilesignhashstatusrequest
type getMobileSignHashStatusRequest struct {
	// nolint: lll, the XML namespace forces us to use a long line.
	XMLName       xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileSignHashStatusRequest"`
	Sesscode      string
	WaitSignature bool
}

type getMobileSignHashStatusResponse struct {
	// nolint: lll, the XML namespace forces us to use a long line.
	XMLName   xml.Name `xml:"http://www.sk.ee/DigiDocService/DigiDocService_2_3.wsdl GetMobileSignHashStatusResponse"`
	Status    string
	Signature string
}

// GetMobileSignHashStatus queries the status of a Mobile-ID signing session.
// If err is nil and signature is empty, then the transaction is still
// outstanding.
func (c *Client) GetMobileSignHashStatus(ctx context.Context, sesscode string) (
	signature []byte, err error) {

	var resp getMobileSignHashStatusResponse
	if err = soapRequest(ctx, c.conf.URL, getMobileSignHashStatusRequest{
		Sesscode:      sesscode,
		WaitSignature: false,
	}, &resp); err != nil {
		return nil, GetMobileSignHashStatusError{Err: err}
	}

	switch resp.Status {
	case "OUTSTANDING_TRANSACTION":
	case "SIGNATURE":
		signature, err = base64.StdEncoding.DecodeString(resp.Signature)
		if err != nil {
			return nil, DecodeSignatureError{
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
	case "REVOKED_CERTIFICATE":
		var cert CertificateError
		cert.Err = SigningCertificateRevokedError{}
		err = cert
	default:
		var status StatusError
		status.Err = UnexpectedSignatureStatusError{Status: resp.Status}
		err = status
	}
	return
}
