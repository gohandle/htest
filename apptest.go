package htest

import (
	"net/http"
	"testing"
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
