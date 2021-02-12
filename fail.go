package htest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/Joker/hpp"
	"github.com/PuerkitoBio/goquery"
)

// TruncateBodyLength is the length after which the rest of the response print
// will be ignored and replaced with ...
var TruncateBodyLength = 512

// ErrTestFail is thrown when the failer fails for testing
var ErrTestFail = errors.New("test fail")

// Failer is used to fail cases with extra context that should help
// the developer figure out what went wrong.
type Failer interface {
	Fatalf(format string, args ...interface{})
	WithRequest(*http.Request) Failer
	WithResponse(*httptest.ResponseRecorder) Failer
	WithHTML(sel *goquery.Selection) Failer
	Copy() Failer
}

// Fail is a failer implementation that cals FF on fatalf
type Fail struct {
	FF func(format string, args ...interface{})

	req *http.Request
	rec *httptest.ResponseRecorder
	sel *goquery.Selection
}

// NewTestFail gives a failer that will panic and write to a buffer
func NewTestFail() (*Fail, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	return &Fail{FF: func(format string, args ...interface{}) {
		fmt.Fprintf(buf, format, args...)
		panic(ErrTestFail) // unwind the stack
	}}, buf
}

//NewStdFail will fail using the standard library tb interface
func NewStdFail(tb testing.TB) *Fail {
	return &Fail{FF: tb.Fatalf}
}

// Fatalf will end the test and fail
func (f *Fail) Fatalf(format string, args ...interface{}) {
	if f.req != nil {
		format += "\n ---> REQUEST: %s"
		args = append(args, fmtRequest(f.req))
	}

	if f.rec != nil {
		format += "\n <--- RESPONSE: %s"
		args = append(args, fmtResponse(f.rec))
	}

	if f.sel != nil {
		format += "\n ==== HTML SELECTION: %s"
		args = append(args, fmtHTML(f.sel))
	}

	f.FF(format, args...)
}

// WithRequest will configure the fail in context of the request
func (f *Fail) WithRequest(r *http.Request) Failer {
	f.req = r
	return f
}

// WithResponse will configure the fail in context of this response
func (f *Fail) WithResponse(w *httptest.ResponseRecorder) Failer {
	f.rec = w
	return f
}

// WithHTML will configure the fail in context of the selected html
func (f *Fail) WithHTML(sel *goquery.Selection) Failer {
	f.sel = sel
	return f
}

// Copy will return a new instance of the failer
func (f *Fail) Copy() Failer {
	return &Fail{
		FF:  f.FF,
		req: f.req,
		rec: f.rec,
		sel: f.sel,
	}
}

// fmtRequest will pretty format the provided request
func fmtRequest(req *http.Request) string {
	d, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic("htest: failed to dump response: " + err.Error())
	}

	s := string(d)
	if len(s) > TruncateBodyLength {
		s = s[:TruncateBodyLength] + "..."
	}

	return s
}

// fmtResponse will pretty format the provided response
func fmtResponse(rec *httptest.ResponseRecorder) string {
	d, err := httputil.DumpResponse(rec.Result(), true)
	if err != nil {
		panic("htest: failed to dump response: " + err.Error())
	}

	s := string(d)
	if len(s) > TruncateBodyLength {
		s = s[:TruncateBodyLength] + "..."
	}

	return s
}

// fmtHTML will pretty print the provided selection
func fmtHTML(sel *goquery.Selection) string {
	s, err := goquery.OuterHtml(sel)
	if err != nil {
		panic("htest: failed to turn selection into html: " + err.Error())
	}

	buf := bytes.NewBuffer(nil)
	hpp.Format(strings.NewReader(s), buf)
	return buf.String()
}
