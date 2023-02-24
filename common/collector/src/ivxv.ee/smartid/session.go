package smartid

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
)

type sessType string

const (
	sessAuth  sessType = "authentication/etsi"
	sessSign  sessType = "signature/document"
	QSCD      string   = "QSCD"
	QUALIFIED string   = "QUALIFIED"
)

var (
	// hashFunctionNames is map from hash algorithm to it's name for Smart-ID
	// REST API.
	hashFunctionNames = map[crypto.Hash]string{
		crypto.SHA256: "SHA256",
		crypto.SHA384: "SHA384",
		crypto.SHA512: "SHA512",
	}

	// signatureAlgs is map from signature algorithm name in Smart-ID REST API
	// to algorithm type.
	signatureAlgs = map[string]x509.SignatureAlgorithm{
		"sha256WithRSAEncryption": x509.SHA256WithRSA,
		"sha384WithRSAEncryption": x509.SHA384WithRSA,
		"sha512WithRSAEncryption": x509.SHA512WithRSA,
	}
)

// https://github.com/SK-EID/smart-id-documentation#2394-request-parameters
type startSessionRequest struct {
	RelyingPartyUUID         string                     `json:"relyingPartyUUID"`
	RelyingPartyName         string                     `json:"relyingPartyName"`
	CertificateLevel         string                     `json:"certificateLevel,omitempty"`
	Hash                     []byte                     `json:"hash"`
	HashType                 string                     `json:"hashType"`
	AllowedInteractionsOrder []allowedInteractionsOrder `json:"allowedInteractionsOrder"`
	Nonce                    string                     `json:"nonce,omitempty"`
	RequestProperties        []byte                     `json:"requestProperties,omitempty"`
	Capabilities             []string                   `json:"capabilities,omitempty"`
}

type allowedInteractionsOrder struct {
	Type           string `json:"type"`
	DisplayText60  string `json:"displayText60,omitempty"`
	DisplayText200 string `json:"displayText200,omitempty"`
}

// https://github.com/SK-EID/smart-id-documentation#2395-example-response
type startSessionResponse struct {
	SessionID string `json:"sessionID"`
}

// startSession is helper function to start either authentication or signing
// dialog with the user.
func (c *Client) startSession(ctx context.Context, t sessType, identifier string,
	hash []byte, hashType string) (sesscode string, err error) {

	// We cannot use a struct literal, because gen would report it
	// as a duplicate error type.
	var input InputError
	switch {

	case len(hash) == 0:
		input.Err = StartSessionNoHashError{}
		err = input
	case len(hashType) == 0:
		input.Err = StartSessionNoHashTypeError{}
		err = input
	}
	if err != nil {
		return
	}

	interactionsOrder := c.conf.SignInteractionsOrder
	certLevel := c.conf.CertificateLevel
	if t == sessAuth {
		interactionsOrder = c.conf.AuthInteractionsOrder
		if certLevel == QSCD {
			certLevel = QUALIFIED
		}
	}

	var resp startSessionResponse
	if err = httpPost(ctx, c.url+string(t)+"/"+identifier, startSessionRequest{
		RelyingPartyUUID:         c.conf.RelyingPartyUUID,
		RelyingPartyName:         c.conf.RelyingPartyName,
		Hash:                     hash,
		HashType:                 hashType,
		CertificateLevel:         certLevel,
		AllowedInteractionsOrder: interactionsOrder,
	}, &resp); err != nil {
		err = StartSessionError{Err: err}
		return
	}
	sesscode = resp.SessionID

	return
}

// https://github.com/SK-EID/smart-id-documentation#23114-response-structure
type sessionStatusResponse struct {
	State               string            `json:"state"`
	Result              resultResponse    `json:"result"`
	Signature           signatureResponse `json:"signature"`
	Cert                certResponse      `json:"cert"`
	IgnoredProperties   []string          `json:"ignoredProperties"`
	InteractionFlowUsed string            `json:"interactionFlowUsed"`
	DeviceIPAddress     string            `json:"deviceIpAddress"`
}

type resultResponse struct {
	EndResult      string `json:"endResult"`
	DocumentNumber string `json:"documentNumber"`
}

type certResponse struct {
	Value            []byte `json:"value"`
	CertificateLevel string `json:"certificateLevel"`
}

type signatureResponse struct {
	Value     []byte `json:"value"`
	Algorithm string `json:"algorithm"`
}

// getSessionStatus is helper function to get the status of either
// authentication or signing dialog with the user.
func (c *Client) getSessionStatus(ctx context.Context, sesscode string) (
	documentNo string, algorithm string, signature []byte, certDER []byte, err error) {

	var resp sessionStatusResponse
	url := fmt.Sprintf("%ssession/%s?timeoutMs=%d", c.url, sesscode, c.conf.StatusTimeoutMS)
	if err = httpGet(ctx, url, &resp); err != nil {
		return "", "", nil, nil, GetSessionStatusError{Err: err}
	}

	switch resp.State {
	case "RUNNING":
		return
	case "COMPLETE":
	default:
		var status StatusError
		status.Err = UnexpectedSessionStateError{State: resp.State}
		return "", "", nil, nil, status
	}
	switch resp.Result.EndResult {
	case "OK":
		return resp.Result.DocumentNumber, resp.Signature.Algorithm, resp.Signature.Value, resp.Cert.Value, nil
	case "TIMEOUT":
		var expired ExpiredError
		return "", "", nil, nil, expired
	case "DOCUMENT_UNUSABLE":
		var account AccountError
		return "", "", nil, nil, account
	case "REQUIRED_INTERACTION_NOT_SUPPORTED_BY_APP":
		var account AccountError
		return "", "", nil, nil, account
	case "WRONG_VC":
		var verification VerificationError
		return "", "", nil, nil, verification
	case "USER_REFUSED":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	case "USER_REFUSED_CERT_CHOICE":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	case "USER_REFUSED_DISPLAYTEXTANDPIN":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	case "USER_REFUSED_VC_CHOICE":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	case "USER_REFUSED_CONFIRMATIONMESSAGE":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	case "USER_REFUSED_CONFIRMATIONMESSAGE_WITH_VC_CHOICE":
		var canceled CanceledError
		return "", "", nil, nil, canceled
	default:
		var status StatusError
		status.Err = UnexpectedSessionResultError{Result: resp.Result}
		return "", "", nil, nil, status
	}
}
