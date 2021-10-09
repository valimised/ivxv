package mid

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
)

type sessType string

const (
	sessAuth sessType = "authentication"
	sessSign sessType = "signature"
)

var (
	// hashFunctionNames is map from hash algorithm to it's name for MID
	// REST API.
	hashFunctionNames = map[crypto.Hash]string{
		crypto.SHA256: "SHA256",
		crypto.SHA384: "SHA384",
		crypto.SHA512: "SHA512",
	}

	// signatureAlgs is map from signature algorithm name in MID REST API
	// to algorithm type.
	signatureAlgs = map[string]x509.SignatureAlgorithm{
		"SHA256WithRSAEncryption": x509.SHA256WithRSA,
		"SHA384WithRSAEncryption": x509.SHA384WithRSA,
		"SHA512WithRSAEncryption": x509.SHA512WithRSA,

		"SHA256WithECEncryption": x509.ECDSAWithSHA256,
		"SHA384WithECEncryption": x509.ECDSAWithSHA384,
		"SHA512WithECEncryption": x509.ECDSAWithSHA512,
	}
)

// https://github.com/SK-EID/MID#323-request-parameters
type startSessionRequest struct {
	RelyingPartyUUID       string `json:"relyingPartyUUID"`
	RelyingPartyName       string `json:"relyingPartyName"`
	PhoneNumber            string `json:"phoneNumber"`
	NationalIdentityNumber string `json:"nationalIdentityNumber"`
	Hash                   []byte `json:"hash"`
	HashType               string `json:"hashType"`
	Language               string `json:"language"`
	DisplayText            string `json:"displayText,omitempty"`
	DisplayTextFormat      string `json:"displayTextFormat,omitempty"`
}

// https://github.com/SK-EID/MID#325-example-response
type startSessionResponse struct {
	SessionID string `json:"sessionID"`
}

// startSession is helper function to start either authentication or signing
// dialog with the user.
func (c *Client) startSession(ctx context.Context, t sessType, idCode, phone string,
	hash []byte, hashType string) (sesscode string, err error) {

	// We cannot use a struct literal, because gen would report it
	// as a duplicate error type.
	var input InputError
	switch {
	case len(idCode) == 0:
		input.Err = StartSessionNoIDCodeError{}
		err = input
	case len(phone) == 0:
		input.Err = StartSessionNoPhoneError{}
		err = input
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

	message := c.conf.AuthMessage
	if t == sessSign {
		message = c.conf.SignMessage
	}

	var resp startSessionResponse
	if err = httpPost(ctx, c.url+string(t), startSessionRequest{
		RelyingPartyUUID:       c.conf.RelyingPartyUUID,
		RelyingPartyName:       c.conf.RelyingPartyName,
		NationalIdentityNumber: idCode,
		PhoneNumber:            phone,
		Hash:                   hash,
		HashType:               hashType,
		Language:               c.conf.Language,
		DisplayText:            message,
		DisplayTextFormat:      c.conf.MessageFormat,
	}, &resp); err != nil {
		err = StartSessionError{Err: err}
		return
	}

	sesscode = resp.SessionID

	return
}

// https://github.com/SK-EID/MID#335-response-structure
type sessionStatusResponse struct {
	State     string            `json:"state"`
	Result    string            `json:"result"`
	Signature signatureResponse `json:"signature"`
	Cert      []byte            `json:"cert"`
	Time      string            `json:"time"`
	TraceID   string            `json:"traceId"`
}

type signatureResponse struct {
	Value     []byte `json:"value"`
	Algorithm string `json:"algorithm"`
}

// getSessionStatus is helper function to get the status of either
// authentication or signing dialog with the user.
func (c *Client) getSessionStatus(ctx context.Context, t sessType, sesscode string) (
	algorithm string, signature []byte, certDER []byte, err error) {

	var resp sessionStatusResponse
	url := fmt.Sprintf("%s%s/session/%s?timeoutMs=%d", c.url, t, sesscode, c.conf.StatusTimeoutMS)
	if err = httpGet(ctx, url, &resp); err != nil {
		return "", nil, nil, GetSessionStatusError{Err: err}
	}

	switch resp.State {
	case "RUNNING":
		return
	case "COMPLETE":
	default:
		var status StatusError
		status.Err = UnexpectedSessionStateError{State: resp.State}
		return "", nil, nil, status
	}

	switch resp.Result {
	case "OK":
		return resp.Signature.Algorithm, resp.Signature.Value, resp.Cert, nil
	case "TIMEOUT":
		var expired ExpiredError
		return "", nil, nil, expired
	case "NOT_MID_CLIENT":
		var notMID NotMIDUserError
		return "", nil, nil, notMID
	case "USER_CANCELLED":
		var canceled CanceledError
		return "", nil, nil, canceled
	case "SIM_ERROR", "SIGNATURE_HASH_MISMATCH":
		var sim SIMError
		sim.Result = resp.Result
		return "", nil, nil, sim
	case "PHONE_ABSENT":
		var absent AbsentError
		return "", nil, nil, absent
	default:
		var status StatusError
		status.Err = UnexpectedSessionResultError{Result: resp.Result}
		return "", nil, nil, status
	}
}
