/*
Package tsp implements a PKIX timestamping client.
https://tools.ietf.org/html/rfc3161
*/
package tsp // import "ivxv.ee/tsp"

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/safereader"
)

const (
	// The maximum amount the response GenTime can differ from
	// the current time at the time of response validation
	maxAge = 1 * time.Minute

	// The maximum amount the response GenTime can be set in the future
	// allowing for correction for system clock inconsistencies
	//
	// One second for timestamps with only a second accuracy (i.e., that do
	// not contain fractions of seconds in Gentime) and one second for
	// actual clock skew.
	maxSkew = 2 * time.Second

	// Maximum size for the timestamping server response.
	maxResponseSize = 10240 // 10 KiB.
)

// The CMS content type used for timestamp tokens.
// https://tools.ietf.org/html/rfc3161#page-8
var idCTTSTInfo = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 1, 4}

// Conf contains the configurable options for the TSP client. It only contains
// serialized values such that it can easily be unmarshaled from a file.
type Conf struct {
	// The URL where the TSP requests are sent.
	URL string

	// The certificates used by the server to sign the token.
	Signers []string

	// The maximum time that GenTime and SignTime can differ in a timestamp
	DelayTime int64

	// Retry is the amount of times a TSP request is retried in case of
	// network or server errors.
	Retry uint64
}

// Client is used for performing TSP requests and checking responses.
type Client struct {
	url     string
	signers []*x509.Certificate
	delay   time.Duration
	retry   uint64
}

// New returns a new TSP client with the provided configuration.
func New(conf *Conf) (c *Client, err error) {
	if len(conf.Signers) == 0 {
		return nil, UnconfiguredSignersError{}
	}

	c = &Client{
		url:   conf.URL,
		delay: time.Duration(conf.DelayTime) * time.Second,
		retry: conf.Retry,
	}
	if c.signers, err = cryptoutil.PEMCertificates(conf.Signers...); err != nil {
		return nil, SignerParsingError{Err: err}
	}
	return
}

// Create requests and verifies a timestamp on data from the configured TSP
// server. If nonce is not nil, then that value will be used as the nonce in
// the request, otherwise a random value is generated. Create returns a
// DER-encoded timestamp token.
func (c *Client) Create(ctx context.Context, data, nonce []byte) ([]byte, error) {
	if len(c.url) == 0 {
		return nil, UnconfiguredURLError{}
	}

	// If there is no nonce, generate one
	if nonce == nil {
		nonce = make([]byte, 20)
		if _, err := rand.Read(nonce); err != nil {
			return nil, GenerateNonceError{Err: err}
		}
	}

	var tst timeStampToken
	var err error
retry:
	for attempt := uint64(0); ; attempt++ {
		// nolint: dupl, No need to deduplicate such a small snippet.
		switch tst, err = c.submitRequest(ctx, data, nonce); {
		case err == nil:
			break retry
		case attempt < c.retry && shouldRetry(err):
			log.Log(ctx, RetryingRequestSubmission{Attempt: attempt + 1, Err: err})
			time.Sleep(1 * time.Second)
		default:
			return nil, RequestSubmissionError{Err: err}
		}
	}

	info, err := checkTSTInfo(tst, data, nonce)
	if err != nil {
		return nil, TSTInfoCheckError{Err: err}
	}

	if err = checkGenTime(time.Now(), info.GenTime, info.Accuracy); err != nil {
		return nil, GenTimeCheckError{Err: err}
	}

	if err = c.checkSignedData(tst, info.GenTime); err != nil {
		return nil, SignedDataCheckError{Err: err}
	}

	return tst.Raw, nil
}

func shouldRetry(err error) bool {
	return errors.Walk(err, func(err error) error {
		switch t := err.(type) {

		// Sending the HTTP request or reading the response body failed.
		case SendRequestError, ResponseBodyReadError:
			return err

		// The HTTP status was a 5xx code.
		case ResponseStatusNotOK:
			if strings.HasPrefix(t.Status.(string), "5") {
				return err
			}
		}
		return nil
	}) != nil
}

// Check checks a stored DER-encoded timestamp token on data and returns the
// token generation time. If nonce is not nil, then the nonce in the timestamp
// token must match that value.
func (c *Client) Check(response, data, nonce []byte) (time.Time, error) {
	var tsToken timeStampToken
	rest, err := asn1.Unmarshal(response, &tsToken)
	if err != nil {
		return time.Time{}, TSTokenUnmarshalError{Err: err}
	}
	if len(rest) > 0 {
		return time.Time{}, TSTokenExcessBytes{ExcessBytes: rest}
	}

	info, err := checkTSTInfo(tsToken, data, nonce)
	if err != nil {
		return time.Time{}, CheckTSTInfoCheckError{Err: err}
	}

	if err = c.checkSignedData(tsToken, info.GenTime); err != nil {
		return time.Time{}, CheckSignedDataCheckError{Err: err}
	}
	return info.GenTime, nil
}

func (c *Client) submitRequest(ctx context.Context, data, nonce []byte) (
	tst timeStampToken, err error) {

	// Construct the request.
	hash := sha256.Sum256(data)

	var n *big.Int
	if nonce != nil {
		n = new(big.Int).SetBytes(nonce)
	}

	req := tsRequest{
		Version: 1,
		MessageImprint: messageImprint{
			HashAlgorithm: pkix.AlgorithmIdentifier{
				// Algorithm identifier for SHA-256.
				Algorithm: asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1},
			},
			HashedMessage: hash[:],
		},
		Nonce:   n,
		CertReq: true,
	}

	reqBytes, err := asn1.Marshal(req)
	if err != nil {
		err = MarshallingRequestError{Err: err}
		return
	}

	// Submit the request and read the response.
	httpReq, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(reqBytes))
	if err != nil {
		err = NewRequestError{Err: err}
		return
	}
	httpReq = httpReq.WithContext(ctx)
	httpReq.Header.Set("Content-Type", "application/timestamp-query")

	reqDump, err := httputil.DumpRequestOut(httpReq, true)
	if err != nil {
		err = ReqDumpError{Err: err}
		return
	}
	log.Debug(ctx, RequestDump{Request: string(reqDump)})

	log.Log(ctx, SendingRequest{
		URL:   c.url,
		Hash:  req.MessageImprint.HashedMessage,
		Nonce: n,
	})

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		err = log.Alert(SendRequestError{Err: err})
		return
	}
	log.Log(ctx, ReceivedResponse{})

	defer func() {
		if closeErr := httpResp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	respDump, err := httputil.DumpResponse(httpResp, false)
	if err != nil {
		err = RespDumpError{Err: err}
		return
	}
	log.Debug(ctx, ResponseDump{Dump: string(respDump)})

	if httpResp.StatusCode != http.StatusOK {
		err = ResponseStatusNotOK{Status: httpResp.Status}
		return
	}

	if ctype := httpResp.Header.Get("Content-Type"); ctype != "application/timestamp-reply" {
		err = UnexpectedResponseContentType{ContentType: ctype}
		return
	}

	// The entire response will be returned (as tst.Raw) so we need to
	// allocate a new byte slice using ioutil.ReadAll.
	body, err := ioutil.ReadAll(safereader.New(httpResp.Body, maxResponseSize))
	if err != nil {
		err = ResponseBodyReadError{Err: err}
		return
	}
	log.Debug(ctx, BodyDump{Body: body})

	// Parse the response and check TSP status.
	var resp tsResponse
	rest, err := asn1.Unmarshal(body, &resp)
	if err != nil {
		err = ResponseUnmarshalError{Err: err}
		return
	} else if len(rest) > 0 {
		err = ResponseUnmarshalExcessBytes{Bytes: rest}
		return
	}

	// https://tools.ietf.org/html/rfc3161#section-2.4.2
	// Status 0 means granted. We do not allow status 1 meaning granted with modifications
	if resp.Status.Status != 0 {
		err = TSStatusError{Status: resp.Status.Status}
		return
	}

	tst = resp.TimeStampToken
	return
}

func checkTSTInfo(tst timeStampToken, data, nonce []byte) (info tstInfo, err error) {
	encap := tst.Content.EncapContentInfo
	if !encap.EContentType.Equal(idCTTSTInfo) {
		return info, UnexpectedEcontentType{Type: encap.EContentType}
	}

	rest, err := asn1.Unmarshal(encap.EContent, &info)
	if err != nil {
		return info, EContentUnmarshalError{Err: err}
	}
	if len(rest) != 0 {
		return info, EContentUnmarshalExcessBytes{Bytes: rest}
	}

	// https://tools.ietf.org/html/rfc3161#page-8
	if info.Version != 1 {
		return info, UnexpectedTSTInfoVersion{Version: info.Version}
	}

	// https://tools.ietf.org/html/rfc3161#page-8
	// info.MessageImprint MUST have the same value as the similar field in
	// TimeStampReq. We relax this a little bit: the message imprint must
	// be a digest over data. This means that the response can use a
	// different algorithm than the request, but since tha TSA does not
	// know the data, then they should be unable to construct another
	// digest.
	chash, ok := digestAlgs[info.MessageImprint.HashAlgorithm.Algorithm.String()]
	if !ok {
		return info, UnsupportedMessageImprintAlgorithm{
			Algorithm: info.MessageImprint.HashAlgorithm,
		}
	}
	hash := chash.New()
	hash.Write(data)
	calculated := hash.Sum(nil)

	if !bytes.Equal(info.MessageImprint.HashedMessage, calculated) {
		return info, MessageImprintMismatch{
			Algorithm: info.MessageImprint.HashAlgorithm,
			Token:     info.MessageImprint.HashedMessage,
			Data:      calculated,
		}
	}

	// https://tools.ietf.org/html/rfc3161#page-11
	// If the nonce field is present in the request it MUST be present in the response.
	// Also both nonces MUST have the same value.
	if nonce != nil {
		if info.Nonce == nil {
			return info, ResponseNonceMissing{}
		}
		n := new(big.Int).SetBytes(nonce)
		if info.Nonce.Cmp(n) != 0 {
			return info, ResponseNonceMismatch{
				Expected: n,
				Got:      info.Nonce,
			}
		}
	}
	return
}

func checkGenTime(now, gen time.Time, acc accuracy) error {
	accuracy := time.Duration(acc.Seconds)*time.Second +
		time.Duration(acc.Millis)*time.Millisecond +
		time.Duration(acc.Micros)*time.Microsecond

	if age := now.Sub(gen) + accuracy; age > maxAge {
		return GenTimeTooOld{Age: age, GenTime: gen}
	}

	skewed := now.Add(maxSkew - accuracy)
	if gen.After(skewed) {
		return GenTimeSetInFuture{Skewed: skewed, GenTime: gen}
	}
	return nil
}
