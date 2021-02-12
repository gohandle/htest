package htest

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFailWithoutContext(t *testing.T) {
	fail, buf := NewTestFail()
	defer func() {
		recover()
		if buf.String() != "foo: bar" {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	fail.Fatalf("foo: %v", "bar")
}

func TestFailWithSelectionContext(t *testing.T) {
	fail, buf := NewTestFail()
	defer func() {
		recover()
		if !strings.Contains(buf.String(), "<html>") {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<p></p>`))

	fail.WithHTML(doc.Selection)
	fail.Fatalf("foo: %v", "bar")
}

func TestFailWithResponseContext(t *testing.T) {
	fail, buf := NewTestFail()
	defer func() {
		recover()
		if !strings.Contains(buf.String(), "HTTP/1.1") {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	rec := httptest.NewRecorder()
	rec.Body = bytes.NewBuffer(nil)
	fmt.Fprintf(rec.Body, "some body")

	fail.WithResponse(rec)
	fail.Fatalf("foo: %v", "bar")
}

func TestFailWithRequestContext(t *testing.T) {
	fail, buf := NewTestFail()
	defer func() {
		recover()
		if !strings.Contains(buf.String(), "POST") {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	req := httptest.NewRequest("POST", "/foo", strings.NewReader("body"))

	fail.WithRequest(req)
	fail.Fatalf("foo: %v", "bar")
}
