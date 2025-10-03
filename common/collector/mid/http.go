package mid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/safereader"
)

const maxResponseSize = 10240 // 10 KiB.

// https://github.com/SK-EID/MID#327-error-response-contents
type errorResponse struct {
	Error   string `json:"error"`
	Time    string `json:"time"`
	TraceID string `json:"traceId"`
}

func httpGet(ctx context.Context, url string, resp interface{}) error {
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return CreateHTTPGetRequestError{URL: url, Err: err,
			Description: _MID_URI}
	}
	httpReq = httpReq.WithContext(ctx)

	return httpDo(ctx, "", httpReq, resp)
}

func httpPost(ctx context.Context, url string, req interface{}, resp interface{}) error {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return MarshalJSONRequestError{Err: err,
			Description: _MID_JSON}
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonReq))
	if err != nil {
		return CreateHTTPPostRequestError{URL: url, Err: err,
			Description: _MID_HTTP}
	}
	httpReq = httpReq.WithContext(ctx)

	httpReq.Header.Set("Content-Type", "application/json")

	return httpDo(ctx, fmt.Sprintf("%T", req), httpReq, resp)
}

func httpDo(ctx context.Context, tag string, httpReq *http.Request, resp interface{}) error {
	reqDump, err := httputil.DumpRequestOut(httpReq, true)
	if err != nil {
		return DumpHTTPRequestError{Err: err,
			Description: _MID_HTTP}
	}
	log.Debug(ctx, HTTPRequest{Request: string(reqDump),
		Description: _MID_HTTP_READY})

	log.Log(ctx, SendingRequest{URL: httpReq.URL, Method: httpReq.Method,
		BodyType: tag, Description: _MID_HTTP_SEND})
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return log.Alert(SendRequestError{Err: err,
			Description: _MID_RESP_ERR})
	}
	defer func() {
		if cerr := httpResp.Body.Close(); cerr != nil && err == nil {
			err = ResponseBodyCloseError{Err: cerr,
				Description: _MID_CLOSE}
		}
	}()
	log.Log(ctx, ReceivedResponse{Description: _MID_RESP})

	respDump, err := httputil.DumpResponse(httpResp, false)
	if err != nil {
		return DumpHTTPResponseError{Err: err,
			Description: _MID_RESP_PARSE}
	}
	log.Debug(ctx, HTTPResponse{Response: string(respDump),
		Description: _MID_RESP_PARSE_OK})

	// Does encoding/json.Unmarshal retain any references to the
	// original byte slice in the unmarshaled structure? If not, then
	// instead of allocating a new byte slice here we could reuse pooled
	// buffers for temporarily storing the JSON between reading and
	// decoding.
	body, err := io.ReadAll(safereader.New(httpResp.Body, maxResponseSize))
	if err != nil {
		return ReadHTTPResponseBodyError{Err: err,
			Description: _MID_RESP_READ}
	}
	log.Debug(ctx, HTTPResponseBody{Body: string(body),
		Description: _MID_RESP_READ_OK})

	if httpResp.StatusCode != http.StatusOK {
		if len(body) > 0 {
			var jsonErr errorResponse
			if err = json.Unmarshal(body, &jsonErr); err != nil {
				return UnmarshalJSONErrorResponseError{Err: err,
					Description: _MID_RESP_JSONCODE}
			}

			err = ErrorResponseError{
				HTTPStatus:  httpResp.Status,
				Message:     jsonErr.Error,
				Time:        jsonErr.Time,
				TraceID:     jsonErr.TraceID,
				Description: _MID_RESP_ERR,
			}
			return err

		}
		return HTTPStatusError{Status: httpResp.Status, Description: _MID_ERRCODE}
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return UnmarshalJSONResponseError{Err: err,
			Description: _MID_RESP_JSON}
	}
	return nil
}
