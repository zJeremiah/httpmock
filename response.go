package httpmock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ResponderFromResponse wraps an *http.Response in a Responder
func ResponderFromResponse(resp *http.Response) Responder {
	return func(req *http.Request) (*http.Response, error) {
		return resp, nil
	}
}

// NewStringResponse creates an *http.Response with a body based on the given string.  Also accepts
// an http status code.
func NewStringResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       NewRespBodyFromString(body),
		Header:     http.Header{},
	}
}

// NewStringResponder creates a Responder from a given body (as a string) and status code.
func NewStringResponder(status int, body string) Responder {
	return ResponderFromResponse(NewStringResponse(status, body))
}

// NewBytesResponse creates an *http.Response with a body based on the given bytes.  Also accepts
// an http status code.
func NewBytesResponse(status int, body []byte) *http.Response {
	return &http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       NewRespBodyFromBytes(body),
		Header:     http.Header{},
	}
}

// NewBytesResponder creates a Responder from a given body (as a byte slice) and status code.
func NewBytesResponder(status int, body []byte) Responder {
	return ResponderFromResponse(NewBytesResponse(status, body))
}

// NewJsonResponse creates an *http.Response with a body that is a json encoded representation of
// the given interface{}.  Also accepts an http status code.
func NewJsonResponse(status int, body interface{}) (*http.Response, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	response := NewBytesResponse(status, encoded)
	response.Header.Set("Content-Type", "application/json")
	return response, nil
}

// NewJsonResponder creates a Responder from a given body (as an interface{} that is encoded to
// json) and status code.
func NewJsonResponder(status int, body interface{}) (Responder, error) {
	resp, err := NewJsonResponse(status, body)
	if err != nil {
		return nil, err
	}
	return ResponderFromResponse(resp), nil
}

// NewXmlResponse creates an *http.Response with a body that is an xml encoded representation
// of the given interface{}.  Also accepts an http status code.
func NewXmlResponse(status int, body interface{}) (*http.Response, error) {
	encoded, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}
	response := NewBytesResponse(status, encoded)
	response.Header.Set("Content-Type", "application/xml")
	return response, nil
}

// NewXmlResponder creates a Responder from a given body (as an interface{} that is encoded to xml)
// and status code.
func NewXmlResponder(status int, body interface{}) (Responder, error) {
	resp, err := NewXmlResponse(status, body)
	if err != nil {
		return nil, err
	}
	return ResponderFromResponse(resp), nil
}

// NewRespBodyFromString creates an io.ReadCloser from a string that is suitable for use as an
// http response body.
func NewRespBodyFromString(body string) io.ReadCloser {
	return &dummyReadCloser{strings.NewReader(body)}
}

// NewRespBodyFromBytes creates an io.ReadCloser from a byte slice that is suitable for use as an
// http response body.
func NewRespBodyFromBytes(body []byte) io.ReadCloser {
	return &dummyReadCloser{bytes.NewReader(body)}
}

// same as ResponderFromResponse with delay support
func ResponderFromDelayResponse(delay time.Duration, resp *http.Response) Responder {
	return func(req *http.Request) (*http.Response, error) {
		time.Sleep(delay)
		return resp, nil
	}
}

// NewStringResponder creates a Responder from a given body (as a string) and status code.
// it use delay to test cancellation incoming request
func NewStringResponderWithDelay(delay time.Duration, status int, body string) Responder {
	return ResponderFromDelayResponse(delay, NewStringResponse(status, body))
}

// NewStringResponse creates an *http.Response with a body based on the given string.  Also accepts
// an http status code.
func NewSlowStringResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       NewSlowRespBodyFromString(body),
		Header:     http.Header{},
	}
}

func NewSlowRespBodyFromString(body string) io.ReadCloser {
	return &dummyReadCloser{NewSlowReader(strings.NewReader(body), 4096)}
}

func NewSlowStringResponder(delay time.Duration, status int, body string) Responder {
	return ResponderFromResponse(NewSlowStringResponse(status, body))
}

type dummyReadCloser struct {
	body io.ReadSeeker
}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	n, err = d.body.Read(p)
	if err == io.EOF {
		d.body.Seek(0, 0)
	}
	return n, err
}

func (d *dummyReadCloser) Close() error {
	return nil
}

type SlowReader struct {
	delay time.Duration
	r     io.ReadSeeker
}

func (sr SlowReader) Read(p []byte) (int, error) {
	time.Sleep(sr.delay)
	return sr.r.Read(p[:1])
}

func (sr SlowReader) Seek(offset int64, whence int) (int64, error) {
	return sr.r.Seek(offset, whence)
}

func NewSlowReader(r io.ReadSeeker, bps int) io.ReadSeeker {
	delay := time.Second / time.Duration(bps)
	return SlowReader{
		r:     r,
		delay: delay,
	}
}
