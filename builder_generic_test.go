package goty_test

import (
	"testing"

	"golift.io/goty"
)

// APIResponse is a standard response to our caller.
type APIResponse[T any] struct {
	Status string `json:"status"`
	Msg    T      `json:"message"`
}

func TestGenericStructName(t *testing.T) {
	t.Parallel()

	goat := goty.NewGoty(nil)
	goat.Parse(APIResponse[any]{})

	vals := goat.Values()
	if len(vals) != 1 {
		t.Fatalf("got %d structs, want 1", len(vals))
	}

	if vals[0].Name != "APIResponse" {
		t.Errorf("Name = %q, want APIResponse", vals[0].Name)
	}

	wantGo := "golift.io/goty_test.APIResponse"
	if vals[0].GoName != wantGo {
		t.Errorf("GoName = %q, want %q", vals[0].GoName, wantGo)
	}
}
