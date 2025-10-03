package ocsp

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"tivi.io/core/crypto/util"
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
	// https://tools.ietf.org/html/rfc6960#section-4.4.1
	idPKIXOCSPNonce = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 48, 1, 2}

	// https://tools.ietf.org/html/rfc6960#section-4.2.1
	idPKIXOCSPBasic = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 48, 1, 1}

	// Map of signature algorithms
	sigMap = map[string]x509.SignatureAlgorithm{
		"1.2.840.113549.1.1.11": x509.SHA256WithRSA,
		"1.2.840.113549.1.1.12": x509.SHA384WithRSA,
		"1.2.840.113549.1.1.13": x509.SHA512WithRSA,
	}

	// OID of the hash function used to calculate CertID fields
	certIDHashOID = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
)

// VerifyOCSPCertificate verifies the status of the certificate against
// the configured OCSP server. If nonce is not nil, then that value will
// be used as the nonce in the request, otherwise no nonce is used.
func VerifyOCSPCertificate(ctx context.Context, cert, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte, conf Conf) (status LiveCertStatus, err error) {
	if len(conf.URL) == 0 {
		return LiveCertStatus{}, fmt.Errorf("no OCSP URL specified")
	}
	retries := util.DefaultUint64(conf.Retries, 1)
	reqCert, err := newCertIDRequest(cert)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("new cert id request: %w", err)
	}
	basicResponse, err := sendOCSPRequest(ctx, conf.URL, reqCert, nonce, retries)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("send ocsp request: %w", err)
	}
	respNonce, err := verifyOCSPResponse(basicResponse, reqCert, issuer, responders, nonce)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("verify ocsp response: %w", err)
	}
	single := basicResponse.TBSResponseData.Responses[0]
	return LiveCertStatus{
		RawResponse: basicResponse.Raw,
		CertStatus: CertStatus{
			ProducedAt:       basicResponse.TBSResponseData.ProducedAt,
			Good:             bool(single.CertStatusGood),
			Unknown:          bool(single.CertStatusUnknown),
			Nonce:            respNonce,
			RevocationReason: int(single.CertStatusRevokedInfo.RevocationReason),
		},
	}, nil
}

// VerifyResponse verifies a DER-encoded full OCSP response. a full OCSP
// response envelopes a basic OCSP response with the status of the server
// response and the OID of the response type. VerifyResponse unpacks the
// basic response and calls VerifyBasicResponse.
func VerifyResponse(bytes []byte, cert, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte) (status LiveCertStatus, err error) {
	resp, err := unmarshalOCSPResponse(bytes)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("unmarshal ocsp response: %w", err)
	}
	return VerifyBasicResponse(resp.ResponseBytes.Response, cert, issuer, responders, nonce)
}

// VerifyBasicResponse verifies a stored DER-encoded basic OCSP response. If nonce is
// not nil, then the nonce in the response must match that value.
func VerifyBasicResponse(response []byte, cert, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte) (status LiveCertStatus, err error) {
	basicResponse, err := unmarshalBasicResponse(response)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("unmarshal basic response: %w", err)
	}
	reqCert, err := newCertIDRequest(cert)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("new cert id request: %w", err)
	}
	respNonce, err := verifyStoredResponse(basicResponse, reqCert, issuer, responders, nonce)
	if err != nil {
		return LiveCertStatus{}, fmt.Errorf("verify stored response: %w", err)
	}
	single := basicResponse.TBSResponseData.Responses[0]
	return LiveCertStatus{
		RawResponse: basicResponse.Raw,
		CertStatus: CertStatus{
			ProducedAt:       basicResponse.TBSResponseData.ProducedAt,
			Good:             bool(single.CertStatusGood),
			Unknown:          bool(single.CertStatusUnknown),
			Nonce:            respNonce,
			RevocationReason: int(single.CertStatusRevokedInfo.RevocationReason)},
	}, nil
}

// ParseTime parses a DER-encoded basic OCSP response and returns the time it
// was produced at. ParseTime does not check the validity of the response, it only
// returns the producedAt time value.
func ParseTime(response []byte) (time.Time, error) {
	basicResponse, err := unmarshalBasicResponse(response)
	if err != nil {
		return time.Time{}, fmt.Errorf("unmarshal basic response: %w", err)
	}
	return basicResponse.TBSResponseData.ProducedAt, nil
}

const ocspReply = "application/ocsp-response"

// sendOCSPRequest creates, marshals and sends the HTTP POST request
// to the OCSP server, retrying in case of failure. If successful,
// it unmarshals and returns the basic OCSP response.
func sendOCSPRequest(ctx context.Context, url string, certID *certID, nonce []byte, retries uint64) (basicOCSPResponse, error) {
	// Create the request
	request := ocspRequest{
		TBSRequest: tbsRequest{
			RequestList: []request{{ReqCert: *certID}},
		},
	}
	if nonce != nil {
		request.TBSRequest.RequestExtensions = []pkix.Extension{{
			Id:    idPKIXOCSPNonce,
			Value: nonce,
		}}
	}
	// ASN11Marshal the request
	reqBytes, err := asn1.Marshal(request)
	if err != nil {
		return basicOCSPResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	// Create the HTTP POST request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return basicOCSPResponse{}, fmt.Errorf("new http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/ocsp-request")
	var httpRespBody []byte
	// Send the HTTP request
	// In case of (server) failure, retry the number of times specified
	// with Exponential Backoff sleep time in between
	for tries := uint64(0); ; tries++ {
		if tries > retries {
			return basicOCSPResponse{}, err
		}
		sleep := time.Duration(math.Pow(2, float64(tries))) * time.Second
		var httpResp *http.Response
		httpResp, err = http.DefaultClient.Do(httpReq)
		if err != nil {
			err = fmt.Errorf("send http request: %w", err)
			time.Sleep(sleep)
			continue
		}
		defer func() {
			if closeErr := httpResp.Body.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()
		if httpResp.StatusCode != http.StatusOK {
			err = fmt.Errorf("bad http status code: wanted '%d' got '%d'", http.StatusOK, httpResp.StatusCode)
			if httpResp.StatusCode/100 != 5 {
				return basicOCSPResponse{}, err
			}
			time.Sleep(sleep)
			continue
		}
		if ct := httpResp.Header.Get("Content-Type"); ct != ocspReply {
			err = fmt.Errorf("bad http response content type: wanted '%s' got '%s'", ocspReply, ct)
			time.Sleep(sleep)
			continue
		}
		httpRespBody, err = ioutil.ReadAll(io.LimitReader(httpResp.Body, maxResponseSize))
		if err != nil {
			err = fmt.Errorf("http response body: %w", err)
			time.Sleep(sleep)
			continue
		}
		break
	}
	resp, err := unmarshalOCSPResponse(httpRespBody)
	if err != nil {
		return basicOCSPResponse{}, fmt.Errorf("unmarshal ocsp response: %w", err)
	}
	basicResponse, err := unmarshalBasicResponse(resp.ResponseBytes.Response)
	if err != nil {
		return basicOCSPResponse{}, fmt.Errorf("unmarshal basic response: %w", err)
	}
	return basicResponse, nil
}

// unmarshalOCSPResponse unmarshals a full OCSP response, verifies the
// response status from the OCSP server, and returns the unmarshalled content.
func unmarshalOCSPResponse(bytes []byte) (ocspResponse, error) {
	var resp ocspResponse
	rest, err := asn1.Unmarshal(bytes, &resp)
	if err != nil {
		return ocspResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(rest) > 0 {
		return ocspResponse{}, fmt.Errorf("unmarshal response excess bytes: %d", len(rest))
	}
	if resp.ResponseStatus != ocspResponseStatusSuccessful {
		status, ok := ocspResponseStatus[int(resp.ResponseStatus)]
		if !ok {
			status = fmt.Sprint(resp.ResponseStatus)
		}
		return ocspResponse{}, fmt.Errorf("bad ocsp status code: wanted '%d' got '%d' (meaning '%s')", ocspResponseStatusSuccessful, resp.ResponseStatus, status)
	}
	if !resp.ResponseBytes.ResponseType.Equal(idPKIXOCSPBasic) {
		return ocspResponse{}, fmt.Errorf("bad ocsp response type: wanted '%s' got '%s'", idPKIXOCSPBasic, resp.ResponseBytes.ResponseType)
	}
	return resp, nil
}

// unmarshalBasicResponse unmarshals a basic OCSP response and
// returns the unmarshalled content.
func unmarshalBasicResponse(response []byte) (basicOCSPResponse, error) {
	var basicResponse basicOCSPResponse
	rest, err := asn1.Unmarshal(response, &basicResponse)
	if err != nil {
		return basicOCSPResponse{}, fmt.Errorf("unmarshal basic response: %w", err)
	}
	if len(rest) > 0 {
		return basicOCSPResponse{}, fmt.Errorf("unmarshal basic response excess bytes: %d", len(rest))
	}
	return basicResponse, err
}

// verifyOCSPResponse verifies if the given response should be accepted, based on
// the certID, the presence of the correct nonce, the responder signature (verifyResponseCommon),
// and the validity of the thisUpdate and producedAt times. Returns the response nonce on success.
func verifyOCSPResponse(resp basicOCSPResponse, cert *certID, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte) ([]byte, error) {
	respNonce, err := verifyResponseCommon(resp, cert, issuer, responders, nonce)
	if err != nil {
		return nil, fmt.Errorf("verify response common: %w", err)
	}
	thisUpdate := resp.TBSResponseData.Responses[0].ThisUpdate
	current := time.Now()
	skewed := current.Add(maxSkew)
	if thisUpdate.After(skewed) {
		return nil, fmt.Errorf("thisUpdate set in the future: thisUpdate '%s' skewed '%s'", thisUpdate, skewed)
	}
	if age := current.Sub(thisUpdate); age > maxAge {
		return nil, fmt.Errorf("thisUpdate too old: thisUpdate '%s' age '%s'", thisUpdate, age)
	}
	if pat := resp.TBSResponseData.ProducedAt; pat.Before(thisUpdate) || pat.After(skewed) {
		return nil, fmt.Errorf("producedAt wrong time: wanted after thisUpdate '%s' and before skewed '%s', got '%s'", thisUpdate, skewed, pat)
	}
	return respNonce, nil
}

// verifyStoredResponse verifies if the given stored response should be accepted, based on
// the certID, the presence of the correct nonce, the responder signature (verifyResponseCommon),
// and the validity of the thisUpdate and producedAt times. Returns the response nonce on success.
func verifyStoredResponse(resp basicOCSPResponse, cert *certID, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte) ([]byte, error) {
	respNonce, err := verifyResponseCommon(resp, cert, issuer, responders, nonce)
	if err != nil {
		return nil, fmt.Errorf("verify response common: %w", err)
	}
	thisUpdate := resp.TBSResponseData.Responses[0].ThisUpdate
	pat := resp.TBSResponseData.ProducedAt
	if pat.Before(thisUpdate) {
		return nil, fmt.Errorf("producedAt wrong time: producedAt '%s' before thisUpdate '%s'", pat, thisUpdate)
	}
	if age := pat.Sub(thisUpdate); age > maxAge {
		return nil, fmt.Errorf("producedAt too old: producedAt '%s' age '%s'", pat, age)
	}
	return respNonce, nil
}

// verifyResponseCommon verifies if the given response should be accepted, based on
// the certID, the presence of the correct nonce, and the responder signature.
// Returns the response nonce on success.
func verifyResponseCommon(basicResponse basicOCSPResponse, cert *certID, issuer *x509.Certificate, responders []*x509.Certificate, nonce []byte) ([]byte, error) {
	if n := len(basicResponse.TBSResponseData.Responses); n != 1 {
		return nil, fmt.Errorf("bad ocsp responses number: wanted '%d' got '%d'", 1, n)
	}
	singleResponse := basicResponse.TBSResponseData.Responses[0]
	if !cert.equal(&singleResponse.CertID) {
		return nil, fmt.Errorf("cert ids not equal")
	}
	var respNonce []byte
	for _, extension := range basicResponse.TBSResponseData.ResponseExtensions {
		if extension.Id.Equal(idPKIXOCSPNonce) {
			respNonce = extension.Value
		}
	}
	if nonce != nil && !bytes.Equal(nonce, respNonce) {
		return nil, fmt.Errorf("bad ocsp response nonce: wanted '%s' got '%s'", nonce, respNonce)
	}
	responder, err := getResponderCertificate(basicResponse, issuer, responders)
	if err != nil {
		return nil, fmt.Errorf("get responder certificate: %w", err)
	}
	algo, ok := sigMap[basicResponse.SignatureAlgorithm.Algorithm.String()]
	if !ok {
		return nil, fmt.Errorf("signature algorithm not supported: %s", basicResponse.SignatureAlgorithm.Algorithm)
	}
	if err = responder.CheckSignature(algo, basicResponse.TBSResponseData.Raw, basicResponse.Signature.RightAlign()); err != nil {
		return nil, fmt.Errorf("responder certificate check signature: %w", err)
	}
	return respNonce, nil
}

// getResponderCertificate returns the basic OCSP responder certificate, if it matches
// one of the certificates in the responders list of if it exists within the
// basic OCSP response.
func getResponderCertificate(basicResponse basicOCSPResponse, issuer *x509.Certificate, responders []*x509.Certificate) (*x509.Certificate, error) {
	responderName := basicResponse.TBSResponseData.ResponderIDByName
	// Verify if the response is signed by a configured responder
	for _, responder := range responders {
		responder.Subject.ExtraNames = responder.Subject.Names
		if util.IsRDNSequenceEqual(responder.Subject.ToRDNSequence(), responderName) {
			return responder, nil
		}
	}
	// Otherwise verify if the certificate is in the response, if it is issued by
	// the same issuer, and if it is allowed for OCSP signing.
	if issuer != nil {
		certPool := x509.NewCertPool()
		certPool.AddCert(issuer)
		opts := x509.VerifyOptions{
			Roots:     certPool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
		}
		for _, der := range basicResponse.Certs {
			responder, err := x509.ParseCertificate(der.FullBytes)
			if err != nil {
				return nil, fmt.Errorf("x509 parse certificate: %w", err)
			}
			responder.Subject.ExtraNames = responder.Subject.Names
			if !util.IsRDNSequenceEqual(responder.Subject.ToRDNSequence(), responderName) {
				continue
			}
			if _, err = responder.Verify(opts); err != nil {
				return nil, fmt.Errorf("responder verify: %w", err)
			}
			return responder, nil
		}
	}
	return nil, fmt.Errorf("responder not found: %s", responderName)
}

// newCertIDRequest returns a new certID object from the provided certificate.
func newCertIDRequest(cert *x509.Certificate) (id *certID, err error) {
	// Since certIDHash is SHA-1 then skip calculating the IssuerKeyHash
	// and simply use AuthorityKeyId.
	if len(cert.AuthorityKeyId) == 0 {
		return nil, fmt.Errorf("authority key id missing")
	}
	nameHash := sha1.Sum(cert.RawIssuer)
	return &certID{
		HashAlgorithm: pkix.AlgorithmIdentifier{
			Algorithm: certIDHashOID,
		},
		IssuerNameHash: nameHash[:],
		IssuerKeyHash:  cert.AuthorityKeyId,
		SerialNumber:   cert.SerialNumber,
	}, nil
}

func (c *certID) equal(other *certID) bool {
	return util.AlgorithmIdentifierCmp(c.HashAlgorithm, other.HashAlgorithm) &&
		bytes.Equal(c.IssuerNameHash, other.IssuerNameHash) &&
		bytes.Equal(c.IssuerKeyHash, other.IssuerKeyHash) &&
		c.SerialNumber.Cmp(other.SerialNumber) == 0
}
