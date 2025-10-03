package timestamp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"time"

	"tivi.io/core/crypto/cms"
	"tivi.io/core/crypto/util"
)

const (
	// maxAge is the maximum amount of time the response GenTime
	// can differ from the current time when validating the response.
	maxAge = 1 * time.Minute

	// maxSkew is the maximum amount of time the response GenTime
	// can be set in the future, correcting system clock inconsistencies.
	// One second for timestamps with only a second accuracy (i.e., that do
	// not contain fractions of seconds in Gentime) plus one second for
	// the actual clock skew.
	maxSkew = 2 * time.Second

	// maxResponseSize is the maximum size allowed for
	// the timestamping server response (10 KiB).
	maxResponseSize = 10240
)

const idSigningTime = "1.2.840.113549.1.9.5"

// RequestTimestamp sends a timestamp request for the given data to the specified TSA server url,
// verifying the response and returning the timestamp token.
func RequestTimestamp(ctx context.Context, data, nonce []byte, conf Conf) ([]byte, error) {
	if len(conf.URL) == 0 {
		return nil, fmt.Errorf("no TSA URL specified")
	}
	retries := util.DefaultUint64(conf.Retries, 2)
	tsToken, err := sendTimestampRequest(ctx, conf.URL, data, nonce, retries)
	if err != nil {
		return nil, fmt.Errorf("send timestamp request: %w", err)
	}
	info, err := verifyTokenInfo(tsToken, data, nonce)
	if err != nil {
		return nil, fmt.Errorf("verify token info: %w", err)
	}
	err = verifyGenTime(info.GenTime, info.Accuracy)
	if err != nil {
		return nil, fmt.Errorf("verify generation time: %w", err)
	}
	delay := time.Duration(util.DefaultUint64(conf.DelayTime, 1)) * time.Second
	otherSignedAttrs := map[string]cms.AttrToVerify{
		idSigningTime: {
			AttributeID: idSigningTime,
			VerifyCallback: func(b []byte) error {
				return verifySigningTime(b, info.GenTime, delay)
			}}}
	err = cms.VerifySignedData(tsToken.Raw, conf.Signers, nil, otherSignedAttrs, map[string]cms.AttrToVerify{})
	if err != nil {
		return nil, fmt.Errorf("verify signed data: %w", err)
	}
	return tsToken.Raw, nil
}

// VerifyTimestamp verifies a DER encoded timestamp response against the original data and nonce,
// returning the genTime, if the verification succeeds.
func VerifyTimestamp(derTimestamp, data, nonce []byte, conf Conf) (time.Time, error) {
	var tsToken timestampToken
	rest, err := asn1.Unmarshal(derTimestamp, &tsToken)
	if err != nil {
		return time.Time{}, fmt.Errorf("unmarshal timestamp token: %w", err)
	}
	if len(rest) > 0 {
		return time.Time{}, fmt.Errorf("unmarshal timestamp token excess bytes: %d", len(rest))
	}
	info, err := verifyTokenInfo(tsToken, data, nonce)
	if err != nil {
		return time.Time{}, fmt.Errorf("verify token info: %w", err)
	}
	delay := time.Duration(util.DefaultUint64(conf.DelayTime, 1)) * time.Second
	otherSignedAttrs := map[string]cms.AttrToVerify{
		idSigningTime: {
			AttributeID: idSigningTime,
			VerifyCallback: func(b []byte) error {
				return verifySigningTime(b, info.GenTime, delay)
			}}}
	err = cms.VerifySignedData(tsToken.Raw, conf.Signers, nil, otherSignedAttrs, map[string]cms.AttrToVerify{})
	if err != nil {
		return time.Time{}, fmt.Errorf("verify signed data: %w", err)
	}
	return info.GenTime, nil
}

// GetGenTime extracts and returns the genTime from a DER encoded timestamp response.
func GetGenTime(derTimestamp []byte) (time.Time, error) {
	var tsToken timestampToken
	rest, err := asn1.Unmarshal(derTimestamp, &tsToken)
	if err != nil {
		return time.Time{}, fmt.Errorf("unmarshal timestamp token: %w", err)
	}
	if len(rest) > 0 {
		return time.Time{}, fmt.Errorf("unmarshal timestamp token excess bytes: %d", len(rest))
	}
	// VerifyAll EContentType
	encap := tsToken.Content.EncapContentInfo
	if !encap.EContentType.Equal(idCTTSTInfo) {
		return time.Time{}, fmt.Errorf("bad econtent type: wanted '%s' got '%s'", idCTTSTInfo, encap.EContentType)
	}
	var info tstInfo
	// ASN1Unmarshal EContent
	rest, err = asn1.Unmarshal(encap.EContent, &info)
	if err != nil {
		return time.Time{}, fmt.Errorf("unmarshal econtent: %w", err)
	}
	if len(rest) > 0 {
		return time.Time{}, fmt.Errorf("unmarshal econtent excess bytes: %d", len(rest))
	}
	return info.GenTime, nil
}

const timestampReply = "application/timestamp-reply"

var sha256Identifier = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}

func sendTimestampRequest(ctx context.Context, url string, data, nonce []byte, retries uint64) (timestampToken, error) {
	// If nonce is not provided, generate a new one
	if nonce == nil {
		nonce = make([]byte, 20)
		if _, err := rand.Read(nonce); err != nil {
			return timestampToken{}, fmt.Errorf("generate nonce: %w", err)
		}
	}
	// Hash the data
	hash := sha256.Sum256(data)
	req := tsRequest{
		Version: 1,
		MessageImprint: messageImprint{
			HashAlgorithm: pkix.AlgorithmIdentifier{
				Algorithm: sha256Identifier,
			},
			HashedMessage: hash[:],
		},
		Nonce:   new(big.Int).SetBytes(nonce),
		CertReq: true,
	}
	// ASN11Marshal the request
	reqBytes, err := asn1.Marshal(req)
	if err != nil {
		return timestampToken{}, fmt.Errorf("marshal request: %w", err)
	}
	// Create the HTTP POST request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return timestampToken{}, fmt.Errorf("new http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/timestamp-query")
	var httpRespBody []byte
	// Send the HTTP request
	// In case of (server) failure, retry the number of times specified
	// with Exponential Backoff sleep time in between
	for tries := uint64(0); ; tries++ {
		if tries > retries {
			return timestampToken{}, err
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
				return timestampToken{}, err
			}
			time.Sleep(sleep)
			continue
		}
		if ct := httpResp.Header.Get("Content-Type"); ct != timestampReply {
			err = fmt.Errorf("bad http response content type: wanted '%s' got '%s'", timestampReply, ct)
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
	var resp tsResponse
	// ASN1Unmarshal the response
	rest, err := asn1.Unmarshal(httpRespBody, &resp)
	if err != nil {
		return timestampToken{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(rest) > 0 {
		return timestampToken{}, fmt.Errorf("unmarshal response excess bytes: %d", len(rest))
	}
	if resp.PKIStatus.Status != 0 {
		return timestampToken{}, fmt.Errorf("bad timestamp status code: wanted '%d' got '%d'", 0, resp.PKIStatus.Status)
	}
	return resp.TimeStampToken, nil
}

const (
	idSHA256 = "2.16.840.1.101.3.4.2.1"
	idSHA384 = "2.16.840.1.101.3.4.2.2"
	idSHA512 = "2.16.840.1.101.3.4.2.3"
)

var (
	idCTTSTInfo      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 1, 4}
	digestAlgorithms = map[string]crypto.Hash{
		idSHA256: crypto.SHA256,
		idSHA384: crypto.SHA384,
		idSHA512: crypto.SHA512,
	}
)

func verifyTokenInfo(tsToken timestampToken, data []byte, nonce []byte) (tstInfo, error) {
	// VerifyAll EContentType
	encap := tsToken.Content.EncapContentInfo
	if !encap.EContentType.Equal(idCTTSTInfo) {
		return tstInfo{}, fmt.Errorf("bad econtent type: wanted '%s' got '%s'", idCTTSTInfo, encap.EContentType)
	}
	var info tstInfo
	// ASN1Unmarshal EContent
	rest, err := asn1.Unmarshal(encap.EContent, &info)
	if err != nil {
		return tstInfo{}, fmt.Errorf("unmarshal econtent: %w", err)
	}
	if len(rest) > 0 {
		return tstInfo{}, fmt.Errorf("unmarshal econtent excess bytes: %d", len(rest))
	}
	if info.Version != 1 {
		return tstInfo{}, fmt.Errorf("bad timestamp token info version: wanted '%d' got '%d'", 1, info.Version)
	}
	// VerifyAll hashed data with messageImprint algorithm
	cHash, ok := digestAlgorithms[info.MessageImprint.HashAlgorithm.Algorithm.String()]
	if !ok {
		return tstInfo{}, fmt.Errorf("unsupported imprint algorithm: %s", info.MessageImprint.HashAlgorithm.Algorithm)
	}
	hash := cHash.New()
	hash.Write(data)
	calculated := hash.Sum(nil)
	if !bytes.Equal(info.MessageImprint.HashedMessage, calculated) {
		return tstInfo{}, fmt.Errorf("different hashed messages with algorithm: %s", info.MessageImprint.HashAlgorithm.Algorithm)
	}
	// VerifyAll nonce
	if nonce != nil {
		if info.Nonce == nil {
			return tstInfo{}, fmt.Errorf("no nonce returned")
		}
		n := new(big.Int).SetBytes(nonce)
		if info.Nonce.Cmp(n) != 0 {
			return tstInfo{}, fmt.Errorf("bad nonce returned: wanted '%s' got '%s'", n, info.Nonce)
		}
	}
	return info, nil
}

func verifyGenTime(gen time.Time, acc accuracy) error {
	now := time.Now()
	accuracy := time.Duration(acc.Seconds)*time.Second +
		time.Duration(acc.Millis)*time.Millisecond +
		time.Duration(acc.Micros)*time.Microsecond
	// VerifyAll GenTime age
	if age := now.Sub(gen) + accuracy; age > maxAge {
		return fmt.Errorf("generation time '%s' too old: aged '%s'", gen, age)
	}
	// VerifyAll GenTime skew
	skewed := now.Add(maxSkew - accuracy)
	if gen.After(skewed) {
		return fmt.Errorf("generation time '%s' too advanced: skewed '%s'", gen, skewed)
	}
	return nil
}

func verifySigningTime(value []byte, gen time.Time, delay time.Duration) error {
	var t time.Time
	rest, err := asn1.Unmarshal(value, &t)
	if err != nil {
		return fmt.Errorf("unmarshal signing time: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal signing time excess bytes: %d", len(rest))
	}
	diff := t.Sub(gen)
	if diff < 0 || diff > delay {
		return fmt.Errorf("signing time '%s' mismatches gen time '%s'", t, gen)
	}
	return nil
}
