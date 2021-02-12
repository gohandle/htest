package htest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gohandle/htest"
)

func TestBasicRequestResponseAssertion(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		var n int
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if act := r.Header.Get("foo"); act != "bar" {
				t.Fatalf("got: %v", act)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("got: %v", err)
			}

			if act := r.FormValue("Rab"); act != "Dar" {
				t.Fatalf("got: %v", act)
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><p class="bar">foo</p></body></html>`)
			n++
		})

		c := htest.New(t, h)
		c.POST("/").
			WithFormData("Rab", "Dar").
			WithHeaders("Foo", "bar").
			EXPECT().
			Status(200).
			BodySize(48).
			Header("Content-Type", "text/html")

		if n != 1 {
			t.Fatalf("got: %v", n)
		}
	})

	t.Run("GET", func(t *testing.T) {
		var m string
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m = r.Method
			w.Header().Set("Foo", "Bar")
			w.WriteHeader(303)
			fmt.Fprintf(w, `<html><body><p class="bar">foo</p></body></html>`)
		})

		c := htest.New(t, h)
		c.GET("/").EXPECT().
			Status(303).
			BodySize(48).
			Header("foo", "Bar")

		if m != http.MethodGet {
			t.Fatalf("got: %v", m)
		}
	})

	t.Run("fail status", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "400") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		c.GET("/").EXPECT().Status(400)
	})

	t.Run("fail header", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "bar") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

		c.GET("/").EXPECT().Header("foo", "bar")
	})

	t.Run("fail body size", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "48") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

		c.GET("/").EXPECT().BodySize(48)
	})

	t.Run("fail custom response checker", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "foo") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		c.GET("/").EXPECT().Check(htest.ResponseCheckerFunc(func(f htest.Failer, rec *httptest.ResponseRecorder) {
			f.Fatalf("foo")
		}))
	})
}
