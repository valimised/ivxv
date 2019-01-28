package dds

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"ivxv.ee/log"
	"ivxv.ee/safereader"
)

const maxResponseSize = 10240 // 10 KiB.

type soapEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    soapBody
}

type soapBody struct {
	XMLName  xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	embedded interface{} // The name of sub-elements is not known so use custom unmarshaling.
}

func (s soapBody) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if err := e.Encode(s.embedded); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

func (s *soapBody) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.Decode(&s.embedded); err != nil {
		return err
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	if _, ok := token.(xml.EndElement); !ok {
		return UnexpectedXMLToken{Token: token}
	}
	return nil
}

type soapFault struct {
	XMLName     xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Message     string   `xml:"detail>message"`
}

func soapRequest(ctx context.Context, url string, req interface{}, resp interface{}) error {
	soapReq, err := xml.Marshal(soapEnvelope{Body: soapBody{embedded: req}})
	if err != nil {
		return MarshalSOAPRequestError{Err: err}
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(soapReq))
	if err != nil {
		return CreateHTTPRequestError{URL: url, Err: err}
	}
	httpReq = httpReq.WithContext(ctx)

	reqDump, err := httputil.DumpRequestOut(httpReq, true)
	if err != nil {
		return DumpHTTPRequestError{Err: err}
	}
	log.Debug(ctx, HTTPRequest{Request: string(reqDump)})

	log.Log(ctx, SendingRequest{URL: url, Type: fmt.Sprintf("%T", req)})
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return log.Alert(SendRequestError{Err: err})
	}
	defer func() {
		if cerr := httpResp.Body.Close(); cerr != nil && err == nil {
			err = ResponseBodyCloseError{Err: cerr}
		}
	}()
	log.Log(ctx, ReceivedResponse{})

	respDump, err := httputil.DumpResponse(httpResp, false)
	if err != nil {
		return DumpHTTPResponseError{Err: err}
	}
	log.Debug(ctx, HTTPResponse{Response: string(respDump)})

	// XXX: Does encoding/xml.Unmarshal retain any references to the
	// original byte slice in the unmarshaled structure? If not, then
	// instead of allocating a new byte slice here we could reuse pooled
	// buffers for temporarily storing the XML between reading and
	// decoding.
	body, err := ioutil.ReadAll(safereader.New(httpResp.Body, maxResponseSize))
	if err != nil {
		return ReadHTTPResponseBodyError{Err: err}
	}
	log.Debug(ctx, HTTPResponseBody{Body: string(body)})

	var soapResp soapEnvelope
	if httpResp.StatusCode != http.StatusOK {
		if len(body) > 0 {
			var fault soapFault
			soapResp.Body.embedded = &fault
			if err = xml.Unmarshal(body, &soapResp); err != nil {
				return UnmarshalSOAPFaultError{Err: err}
			}

			err = SOAPFaultError{
				HTTPStatus: httpResp.Status,
				Code:       fault.FaultCode,
				String:     fault.FaultString,
				Message:    fault.Message,
			}
			// https://sk-eid.github.io/dds-documentation/api/api_docs/#soap-error-messages
			switch fault.FaultString {
			case "101", "102":
				var input InputError
				input.Err = err
				err = input
			case "301":
				var notuser NotMIDUserError
				err = notuser
			case "302", "303", "304", "305":
				var cert CertificateError
				cert.Err = err
				err = cert
			}
			return err

		}
		return HTTPStatusError{Status: httpResp.Status}
	}

	soapResp.Body.embedded = resp
	if err = xml.Unmarshal(body, &soapResp); err != nil {
		return UnmarshalSOAPResponseError{Err: err}
	}
	return nil
}
