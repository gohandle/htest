package htest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Case represents a set of assertions for an http.Handler
type Case struct {
	fail Failer
	h    http.Handler
}

// NewCase will setup a case with a custom failer
func NewCase(fail Failer, h http.Handler) *Case {
	return &Case{fail, h}
}

// New initiates a test case
func New(tb testing.TB, h http.Handler) *Case {
	fail := NewStdFail(tb)
	return NewCase(fail, h)
}

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

	rb.hdr.Set(k, v)
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

// HTMLAsserter asserts (parts of) an HTML document
type HTMLAsserter struct {
	Selection *goquery.Selection
	ra        *ResponseAsserter
	doc       *goquery.Document
	fail      Failer
}

func newHTMLAsserter(ra *ResponseAsserter, fail Failer) *HTMLAsserter {
	ha, err := &HTMLAsserter{ra: ra, fail: fail}, (error)(nil)
	ha.doc, err = goquery.NewDocumentFromReader(ra.rec.Body)
	if err != nil {
		ha.fail.Fatalf("failed to parse response body as HTML: %v", err)
	}

	ha.Selection = ha.doc.Selection
	ha.fail = ra.fail.Copy().WithHTML(ha.Selection)
	return ha
}

// Count asserts how many elements there are to be expected
func (ha *HTMLAsserter) Count(exp int) *HTMLAsserter {
	if act := ha.Selection.Length(); act != exp {
		ha.fail.Fatalf("count: exp '%d': got: '%d'", exp, act)
	}

	return ha
}

// AssertAll will return a new asserter that allows for checking ALL elements that
// match the selector.
func (ha *HTMLAsserter) AssertAll(selector string) *HTMLAsserter {
	sel := ha.Selection.Find(selector)
	if sel.Length() < 1 {
		ha.fail.Fatalf("selector '%s' didn't yield anything", selector)
	}

	return &HTMLAsserter{
		Selection: sel,
		ra:        ha.ra,
		doc:       ha.doc,
		fail:      ha.fail.Copy().WithHTML(sel),
	}
}

// TextMatch will assert that the text in the selection matches the pattern
func (ha *HTMLAsserter) TextMatch(expr string) *HTMLAsserter {
	r, err := regexp.Compile(expr)
	if err != nil {
		ha.fail.Fatalf("failed to compile match expr: %v", err)
	}

	if txt := ha.Selection.Text(); !r.MatchString(txt) {
		ha.fail.Fatalf("text match: exp text to match '%s': got: '%s'", expr, txt)
	}
	return ha
}

// Text will assert that the text context of the selection equals 'exp'
func (ha *HTMLAsserter) Text(exp string) *HTMLAsserter {
	if act := ha.Selection.Text(); act != exp {
		ha.fail.Fatalf("text: exp '%s': got: '%s'", exp, act)
	}

	return ha
}

// Attr will assert of the attribute named 'n' has expected value 'exp'
func (ha *HTMLAsserter) Attr(n, exp string) *HTMLAsserter {
	if act := ha.Selection.AttrOr(n, ""); act != exp {
		ha.fail.Fatalf("attr: exp '%s': got: '%s'", exp, act)
	}

	return ha
}
