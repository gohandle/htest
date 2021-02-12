package htest_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gohandle/htest"
)

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

	t.Run("fail custom html checker", func(t *testing.T) {
		fail, buf := htest.NewTestFail()
		defer func() {
			recover()
			if !strings.Contains(buf.String(), "bar") {
				t.Fatalf("got: %v", buf.String())
			}
		}()

		c := htest.NewCase(fail, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<p class="bar">foo</p>`)
		}))

		c.GET("/").EXPECT().HTML().Check(htest.HTMLCheckerFunc(func(f htest.Failer, sel *goquery.Selection) {
			f.Fatalf("bar")
		}))
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
