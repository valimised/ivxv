/*
Package ocsp implements an OCSP client
https://tools.ietf.org/html/rfc6960
*/
package ocsp // import "ivxv.ee/ocsp"

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
	"ivxv.ee/safereader"
)

const (
	// The maximum amount the response thisUpdate can be set in the future
	// allowing for correction for system clock inconsistencies
	maxSkew = 300 * time.Millisecond

	// The maximum amount the response thisUpdate can differ from
	// the current time at the time of response validation
	maxAge = 1 * time.Minute

	// Maximum size for the ocsp server response.
	maxResponseSize = 10240 // 10 KiB.
)

var (
	reader = safereader.New(maxResponseSize)

	// https://tools.ietf.org/html/rfc6960#section-4.4.1
	idPKIXOCSPNonce = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 48, 1, 2}

	// https://tools.ietf.org/html/rfc6960#section-4.2.1
	idPKIXOCSPBasic = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 48, 1, 1}

	// Map of signature algorithms
	sigMap = map[string]x509.SignatureAlgorithm{
		"1.2.840.113549.1.1.5":  x509.SHA1WithRSA,
		"1.2.840.113549.1.1.11": x509.SHA256WithRSA,
		"1.2.840.113549.1.1.12": x509.SHA384WithRSA,
		"1.2.840.113549.1.1.13": x509.SHA512WithRSA,
	}
)

// Conf contains the configurable options for the OCSP client. It only contains
// serialized values such that it can easily be unmarshaled from a file.
type Conf struct {
	URL        string
	Responders []string
}

// Client is used for performing OCSP requests and checking responses.
type Client struct {
	url        string
	responders []*x509.Certificate
}

// New returns a new OCSP client with the provided configuration.
func New(conf *Conf) (c *Client, err error) {
	c = &Client{url: conf.URL}
	if c.responders, err = cryptoutil.PEMCertificates(conf.Responders...); err != nil {
		return nil, ResponderParsingError{Err: err}
	}
	return
}

// CertStatus is the return value for OCSP commands containing all the
// relevant information about the OCSP response.
type CertStatus struct {
	ProducedAt       time.Time // The time this response was produced at.
	Nonce            []byte    // The nonce used in the response
	Good             bool      // Is the status of the requested certificate good?
	Unknown          bool      // Is the status of the certificate unknown
	RevocationReason int       // If the certificate is revoked, the reason of revocation.
}

// LiveCertStatus is an extension of CertStatus returned from live OCSP queries
// which also contains the raw response received from the server.
type LiveCertStatus struct {
	CertStatus
	RawResponse []byte // The raw ASN.1 DER-encoded response from the server.
}

// Check checks the status of the certificate against the configured OCSP
// server. If nonce is not nil, then that value will be used as the nonce in
// the request, otherwise no nonce is used.
func (c *Client) Check(ctx context.Context, cert, issuer *x509.Certificate, nonce []byte) (
	status *LiveCertStatus, err error) {

	if len(c.url) == 0 {
		return nil, UnconfiguredURLError{}
	}

	// Create request
	reqCert, err := newCertID(cert)
	if err != nil {
		return nil, RequestCertIDCreateError{Err: err}
	}

	// Submit the request to url and read the response.
	resp, basic, err := c.submitRequest(ctx, reqCert, nonce)
	if err != nil {
		return nil, SubmitRequestError{Err: err}
	}

	// Check response.
	respNonce, err := c.checkResponse(basic, reqCert, issuer, nonce)
	if err != nil {
		return nil, CheckResponseError{Err: err}
	}

	// Return status.
	single := basic.TBSResponseData.Responses[0]

	return &LiveCertStatus{
		RawResponse: resp,
		CertStatus: CertStatus{
			ProducedAt:       basic.TBSResponseData.ProducedAt,
			Good:             bool(single.CertStatusGood),
			Unknown:          bool(single.CertStatusUnknown),
			Nonce:            respNonce,
			RevocationReason: int(single.CertStatusRevoked.RevocationReason),
		},
	}, nil
}

// CheckFullResponse checks a DER-encoded full OCSP response. A full OCSP
// response envelopes a basic OCSP response with the status of the server
// response and the OID of the response type. CheckFullResponse unpacks the
// basic response and calls CheckResponse.
func (c *Client) CheckFullResponse(response []byte, cert, issuer *x509.Certificate, nonce []byte) (
	status *CertStatus, err error) {

	resp, err := unpackResponse(response)
	if err != nil {
		return nil, UnpackResponseError{Err: err}
	}

	return c.CheckResponse(resp, cert, issuer, nonce)
}

// CheckResponse checks a stored DER-encoded basic OCSP response. If nonce is
// not nil, then the nonce in the response must match that value.
func (c *Client) CheckResponse(response []byte, cert, issuer *x509.Certificate, nonce []byte) (
	status *CertStatus, err error) {

	// Unmarshal the basic response.
	var basic basicOCSPResponse
	if err = unmarshalResponse(response, &basic); err != nil {
		return nil, BasicResponseUnmarshalError{Err: err}
	}

	// Generate the CertID.
	reqCert, err := newCertID(cert)
	if err != nil {
		return nil, ResponseCertIDCreateError{Err: err}
	}

	// Check response.
	respNonce, err := c.checkStoredResponse(&basic, reqCert, issuer, nonce)
	if err != nil {
		return nil, CheckStoredResponseError{Err: err}
	}

	// Return status.
	single := basic.TBSResponseData.Responses[0]

	return &CertStatus{
		ProducedAt:       basic.TBSResponseData.ProducedAt,
		Good:             bool(single.CertStatusGood),
		Unknown:          bool(single.CertStatusUnknown),
		Nonce:            respNonce,
		RevocationReason: int(single.CertStatusRevoked.RevocationReason),
	}, nil
}

func (c *Client) submitRequest(ctx context.Context, reqCert *certID, nonce []byte) (
	response []byte, basic *basicOCSPResponse, err error) {

	r := ocspRequest{
		TBSRequest: tbsRequest{
			RequestList: []request{{ReqCert: *reqCert}},
		},
	}
	if nonce != nil {
		r.TBSRequest.RequestExtensions = []pkix.Extension{{
			Id:    idPKIXOCSPNonce,
			Value: nonce,
		}}
	}
	req, err := asn1.Marshal(r)
	if err != nil {
		err = RequestMarshalError{Err: err}
		return
	}

	var client http.Client

	httpReq, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(req))
	if err != nil {
		err = NewRequestError{Err: err}
		return
	}
	httpReq = httpReq.WithContext(ctx)
	httpReq.Header.Set("Content-Type", "application/ocsp-request")

	reqDump, err := httputil.DumpRequestOut(httpReq, true)
	if err != nil {
		err = ReqDumpError{Err: err}
		return
	}
	log.Debug(ctx, RequestDebugDump{Request: string(reqDump)})

	log.Log(ctx, SendingRequest{
		URL:            c.url,
		Serial:         reqCert.SerialNumber,
		IssuerNameHash: reqCert.IssuerNameHash,
	})

	httpResp, err := client.Do(httpReq)
	if err != nil {
		err = log.Alert(SendRequestError{Err: err})
		return
	}
	defer func() {
		if cerr := httpResp.Body.Close(); cerr != nil && err == nil {
			err = ResponseBodyCloseError{Err: cerr}
		}
	}()

	log.Log(ctx, ReceivedResponse{})

	respDump, err := httputil.DumpResponse(httpResp, false)
	if err != nil {
		err = RespDumpError{Err: err}
		return
	}
	log.Debug(ctx, ResponseDebugDump{Response: string(respDump)})

	if httpResp.StatusCode != http.StatusOK {
		err = UnexpectedResponseStatusError{Status: httpResp.Status}
		return
	}

	if ctype := httpResp.Header.Get("Content-Type"); ctype != "application/ocsp-response" {
		err = UnexpectedContentTypeError{ContentType: ctype}
		return
	}

	respRaw, err := reader.Read(httpResp.Body)
	if err != nil {
		err = ResponseBodyReadError{Err: err}
		return
	}
	defer reader.Recover(respRaw)
	log.Debug(ctx, BodyDump{Body: respRaw})

	// Unpack basicOCSPResponse
	resp, err := unpackResponse(respRaw)
	if err != nil {
		err = ResponseUnpackError{Err: err}
		return
	}

	// Unmarshal the basic response.
	basic = new(basicOCSPResponse)
	if err = unmarshalResponse(resp, basic); err != nil {
		err = UnmarshalResponseError{Err: err}
		return
	}

	// Create copy of raw response before the read buffer is recovered.
	response = make([]byte, len(respRaw))
	copy(response, respRaw)
	return
}

func unpackResponse(der []byte) (response []byte, err error) {
	var resp ocspResponse
	rest, err := asn1.Unmarshal(der, &resp)
	if err != nil {
		return nil, ResponseUnmarshalError{Err: err}
	} else if len(rest) > 0 {
		return nil, ResponseUnmarshalExcessBytesError{Bytes: rest}
	}

	if resp.ResponseStatus != ocspResponseStatusSuccessful {
		status, ok := ocspResponseStatus[int(resp.ResponseStatus)]
		if !ok {
			status = fmt.Sprint(resp.ResponseStatus)
		}
		return nil, UnexpectedOCSPResponseStatusError{Status: status}
	}

	if !resp.ResponseBytes.ResponseType.Equal(idPKIXOCSPBasic) {
		return nil, UnexpectedResponseTypeError{
			ResponseType: resp.ResponseBytes.ResponseType,
		}
	}

	return resp.ResponseBytes.Response, nil
}

func unmarshalResponse(der []byte, basic *basicOCSPResponse) (err error) {
	rest, err := asn1.Unmarshal(der, basic)
	if err != nil {
		return BasicOCSPResponseUnmarshalError{Err: err}
	} else if len(rest) > 0 {
		return BasicOCSPResponseUnmarshalExcessBytesError{Bytes: rest}
	}
	return
}

// checkResponse checks if the given response should be accepted.
func (c *Client) checkResponse(resp *basicOCSPResponse,
	cert *certID, issuer *x509.Certificate, nonce []byte) (
	respNonce []byte, err error) {

	// Perform common checks.
	if respNonce, err = c.checkResponseCommon(resp, cert, issuer, nonce); err != nil {
		return nil, CheckResponseCommonError{Err: err}
	}

	// Check thisUpdate skew and age.
	thisUpdate := resp.TBSResponseData.Responses[0].ThisUpdate
	current := time.Now()
	skewed := current.Add(maxSkew)
	if thisUpdate.After(skewed) {
		return nil, ThisUpdateSetInFutureError{
			ThisUpdate: thisUpdate,
			MaxSkewed:  skewed,
		}
	}
	if age := current.Sub(thisUpdate); age > maxAge {
		return nil, ThisUpdateExceedsMaxAgeError{
			Current:    current,
			ThisUpdate: thisUpdate,
		}
	}

	// Make sure producedAt is between thisUpdate and now + skew.
	if pat := resp.TBSResponseData.ProducedAt; pat.Before(thisUpdate) || pat.After(skewed) {
		return nil, ProducedAtWrongTimeError{
			ProducedAt: pat,
			ThisUpdate: thisUpdate,
			Skewed:     skewed,
		}
	}
	return
}

// checkStoredResponse checks if the given stored response should be accepted
// and returns the response nonce on success
func (c *Client) checkStoredResponse(resp *basicOCSPResponse,
	cert *certID, issuer *x509.Certificate, nonce []byte) (
	respNonce []byte, err error) {

	// Perform common checks.
	if respNonce, err = c.checkResponseCommon(resp, cert, issuer, nonce); err != nil {
		return nil, CheckStoredResponseCommonError{Err: err}
	}

	// Check time between thisUpdate and producedAt.
	thisUpdate := resp.TBSResponseData.Responses[0].ThisUpdate
	pat := resp.TBSResponseData.ProducedAt
	if pat.Before(thisUpdate) {
		return nil, ProducedAtBeforeThisUpdateError{
			ProducedAt: pat,
			ThisUpdate: thisUpdate,
		}
	}
	if age := pat.Sub(thisUpdate); age > maxAge {
		return nil, StoredThisUpdateExceedsMaxAgeError{
			ProducedAt: pat,
			ThisUpdate: thisUpdate,
		}
	}
	return
}

// checkResponseCommon performs checks on the basic response that are common
// for fresh and stored responses.
func (c *Client) checkResponseCommon(resp *basicOCSPResponse,
	cert *certID, issuer *x509.Certificate, nonce []byte) (
	respNonce []byte, err error) {

	// Compare certificate requested and in the response.
	if n := len(resp.TBSResponseData.Responses); n != 1 {
		return nil, UnexpectedTBSResponseCountError{Count: n}
	}
	single := resp.TBSResponseData.Responses[0]
	if !cert.equal(&single.CertID) {
		return nil, CertIDMismatchError{
			Stored:  single.CertID,
			Control: cert,
		}
	}

	// Get the nonce from the response, if present.
	for _, extension := range resp.TBSResponseData.ResponseExtensions {
		if extension.Id.Equal(idPKIXOCSPNonce) {
			respNonce = extension.Value
		}
	}

	// If nonce is not nil, then compare to respNonce.
	if nonce != nil && !bytes.Equal(nonce, respNonce) {
		return nil, ResponseNonceMismatchError{
			Request:  nonce,
			Response: respNonce,
		}
	}

	// Find responder certificate and check signature on the response.
	responder, err := responder(resp, issuer, c.responders)
	if err != nil {
		return nil, ResponderCertificateError{Err: err}
	}

	alg, ok := sigMap[resp.SignatureAlgorithm.Algorithm.String()]
	if !ok {
		return nil, SignatureAlgorithmNotSupportedError{
			Algorithm: resp.SignatureAlgorithm.Algorithm,
		}
	}
	if err = responder.CheckSignature(alg, resp.TBSResponseData.Raw, resp.Signature.RightAlign()); err != nil {
		return nil, CheckSignatureError{Err: err}
	}

	return
}

func responder(resp *basicOCSPResponse, issuer *x509.Certificate, responders []*x509.Certificate) (
	*x509.Certificate, error) {

	name := resp.TBSResponseData.ResponderIDByName

	// First check if the response is signed by a configured responder.
	for _, responder := range responders {
		responder.Subject.ExtraNames = responder.Subject.Names
		if cryptoutil.RDNSequenceEqual(responder.Subject.ToRDNSequence(), name) {
			return responder, nil
		}
	}

	// Otherwise check if its certificate is in the response, is issued by
	// the same issuer, and is allowed for OCSP signing.
	if issuer != nil {
		opts := x509.VerifyOptions{
			Roots:     cryptoutil.CertificatePool(issuer),
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
		}
		for _, der := range resp.Certs {
			responder, err := x509.ParseCertificate(der.FullBytes)
			if err != nil {
				return nil, ParseResponseCertificateError{
					DER: der.FullBytes,
					Err: err,
				}
			}

			responder.Subject.ExtraNames = responder.Subject.Names
			if !cryptoutil.RDNSequenceEqual(responder.Subject.ToRDNSequence(), name) {
				continue
			}

			if _, err = responder.Verify(opts); err != nil {
				return nil, VerifyResponderCertificateError{
					Subject: responder.Subject,
					Issuer:  responder.Issuer,
					Err:     err,
				}
			}
			return responder, nil
		}
	}

	return nil, ResponderCertificateNotFoundError{Responder: name}
}
