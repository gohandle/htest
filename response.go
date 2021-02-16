package htest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// RequestBuilder allows building a request for a testcase
type RequestBuilder struct {
	method string
	path   string
	cas    *Case
	hdr    http.Header
	body   io.Reader
	fail   Failer
}

// GET will start the building of a GET request
func (c *Case) GET(p string) *RequestBuilder {
	return &RequestBuilder{
		method: http.MethodGet,
		path:   p,
		cas:    c,
		fail:   c.fail,
	}
}

// POST will start the building of a POST request
func (c *Case) POST(p string) *RequestBuilder {
	return &RequestBuilder{
		method: http.MethodPost,
		path:   p,
		cas:    c,
		fail:   c.fail,
	}
}

// WithFormData will set the requests' body to the form encoded shape of
// the provided key value pairs and set the content type to
func (rb *RequestBuilder) WithFormData(pairs ...string) *RequestBuilder {
	vals := url.Values{}
	for i := 0; i < len(pairs); i += 2 {
		vals.Set(pairs[i], pairs[i+1])
	}

	rb.body = strings.NewReader(vals.Encode())
	return rb.WithHeader("Content-Type", "application/x-www-form-urlencoded")
}

// WithHeaders allows adding headers as string pairs
func (rb *RequestBuilder) WithHeaders(pairs ...string) *RequestBuilder {
	for i := 0; i < len(pairs); i += 2 {
		rb = rb.WithHeader(pairs[i], pairs[i+1])
	}

	return rb
}

// WithHeader sets a header value for the request building
func (rb *RequestBuilder) WithHeader(k, v string) *RequestBuilder {
	if rb.hdr == nil {
		rb.hdr = http.Header{}
	}

	rb.hdr.Add(k, v)
	return rb
}

// EXPECT will create and perform the request on the case's http handler and return
// an asserter for the http response.
func (rb *RequestBuilder) EXPECT() *ResponseAsserter {
	req := httptest.NewRequest(rb.method, rb.path, rb.body)
	for k, v := range rb.hdr {
		req.Header[k] = v
	}

	rec := httptest.NewRecorder()
	rb.cas.h.ServeHTTP(rec, req)

	return &ResponseAsserter{
		rec: rec,
		req: req,
		b:   rb,
		fail: rb.fail.
			Copy().
			WithRequest(req).
			WithResponse(rec),
	}
}

// ResponseAsserter allows for asserting an http response
type ResponseAsserter struct {
	rec  *httptest.ResponseRecorder
	req  *http.Request
	b    *RequestBuilder
	fail Failer
}

// ResponseChecker can be implemented to check responses through custom logic
type ResponseChecker interface {
	CheckResponse(f Failer, rec *httptest.ResponseRecorder)
}

// ResponseCheckerFunc allows casing a function to implement the ResponseChecker
type ResponseCheckerFunc func(f Failer, rec *httptest.ResponseRecorder)

// CheckResponse implements ResponseChecker
func (f ResponseCheckerFunc) CheckResponse(fail Failer, rec *httptest.ResponseRecorder) {
	f(fail, rec)
}

// Check allows for checking the response using resuable custom logic
func (ra *ResponseAsserter) Check(c ResponseChecker) *ResponseAsserter {
	c.CheckResponse(ra.fail, ra.rec)
	return ra
}

// Status asserts that the status equal c
func (ra *ResponseAsserter) Status(c int) *ResponseAsserter {
	if ra.rec.Code != c {
		ra.fail.Fatalf("status code: got: '%v' exp: '%v'", ra.rec.Code, c)
	}

	return ra
}

// Header asserts that the header 'k' equals 'exp'
func (ra *ResponseAsserter) Header(k, exp string) *ResponseAsserter {
	if act := ra.rec.Header().Get(k); act != exp {
		ra.fail.Fatalf("header '%s': got: '%v' exp: '%v'", k, act, exp)
	}

	return ra
}

// BodySize asserts the size of the response body
func (ra *ResponseAsserter) BodySize(exp int) *ResponseAsserter {
	if act := ra.rec.Body.Len(); act != exp {
		ra.fail.Fatalf("body size: got: '%d', exp: '%d'", act, exp)
	}

	return ra
}

// HTML will parse the response body as html and return an asserter for it
func (ra *ResponseAsserter) HTML() *HTMLAsserter {
	// @TODO do some basic html validation checking
	return newHTMLAsserter(ra, ra.fail)
}
