package htest

import (
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// HTMLChecker can be implemented to check HTML through custom logic
type HTMLChecker interface {
	CheckHTML(f Failer, sel *goquery.Selection)
}

// HTMLCheckerFunc allows for casting a function to implement the HTMLChecker
type HTMLCheckerFunc func(f Failer, se *goquery.Selection)

// CheckHTML implements ResponseChecker
func (f HTMLCheckerFunc) CheckHTML(fail Failer, se *goquery.Selection) {
	f(fail, se)
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

// Check allows for checking the response using resuable custom logic
func (ha *HTMLAsserter) Check(c HTMLChecker) *HTMLAsserter {
	c.CheckHTML(ha.fail, ha.Selection)
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
