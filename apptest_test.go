package htest_test

import (
	"fmt"
	"net/http"
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
}

// We don't curretly validate if the response is valid html because that is pretty hard without
// running a Jar or contacting an outside API. We do expect that if the response is some text that
// it shows up in the error when asserting for any HTML
func TestHTMLValidation(t *testing.T) {
	fail, buf := htest.NewTestFail()
	defer func() {
		recover()
		if !strings.Contains(buf.String(), "some text") {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `some text`)
	}))

	c.GET("/").EXPECT().HTML().AssertAll("p")
}

func TestAssertHTML(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		c := htest.New(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().
			AssertAll("p").Count(1).Text("foo").Attr("class", "bar").TextMatch(`.o.`)
	})

	t.Run("failing html assert", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "bogus") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("bogus")
	})

	t.Run("failing html count", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "2") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("p").Count(2)
	})

	t.Run("failing html count", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "2") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("p").Count(2)
	})

	t.Run("failing html attr", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "bogus") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("p").Attr("class", "bogus")
	})

	t.Run("failing html text", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "fail-text") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("p").Text("fail-text")
	})

	t.Run("failing html text match", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), ".f.") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().AssertAll("p").TextMatch(`.f.`)
	})
}

func TestMultipleHTMLAsserts(t *testing.T) {
	fail, buf := htest.NewTestFail()
	defer func() {
		recover()
		if !strings.Contains(buf.String(), "<html>") {
			t.Fatalf("got: %v", buf.String())
		}
	}()

	c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<p class="bar">foo</p>`)
	}))

	ht := c.GET("/").EXPECT().HTML()
	ht.AssertAll("p")

	// this second assert should still be scoped to the whole html document
	ht.AssertAll("bogus")
}
