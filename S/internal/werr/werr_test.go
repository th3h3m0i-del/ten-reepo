// Package werr test cho New, Wrap, KindOf.
package werr

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(KindSyntax, "line %d fail", 5)
	if err.Kind != KindSyntax {
		t.Errorf("Kind = %v, want syntax", err.Kind)
	}
	if err.Msg != "line 5 fail" {
		t.Errorf("Msg = %q", err.Msg)
	}
}

func TestWrap(t *testing.T) {
	inner := errors.New("io error")
	err := Wrap(KindIO, inner, "không đọc")
	if err.Kind != KindIO {
		t.Errorf("Kind = %v, want io", err.Kind)
	}
	if err.Msg != "không đọc: io error" {
		t.Errorf("Msg = %q", err.Msg)
	}
}

func TestKindOf(t *testing.T) {
	e := New(KindRuntime, "x")
	if KindOf(e) != KindRuntime {
		t.Error("KindOf sai")
	}
	plain := errors.New("plain")
	if KindOf(plain) != KindUnknown {
		t.Error("plain error phải KindUnknown")
	}
}
