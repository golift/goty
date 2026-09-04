package gotydoc_test

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"golift.io/goty/gotydoc"
	"golift.io/goty/gotydoc/internal/sample"
)

func TestAddPkgIgnoresExternalTests(t *testing.T) {
	t.Parallel()

	src := sampleDir(t)
	pkg := "golift.io/goty/gotydoc/internal/sample"

	for n := range 50 {
		docs := gotydoc.New()
		docs.AddPkgMust(src, pkg)

		got := docs.Type(reflect.TypeOf(sample.Config{}))
		if got != "Config is the documented type." {
			t.Fatalf("iteration %d: Config doc = %q", n, got)
		}

		got = docs.Member(reflect.TypeOf(sample.Config{}), "Name")
		if got != "Name is a field comment." {
			t.Fatalf("iteration %d: Name doc = %q", n, got)
		}
	}
}

func TestTypeDocsForGenericInstantiation(t *testing.T) {
	t.Parallel()

	docs := gotydoc.New()
	docs.AddPkgMust(sampleDir(t), "golift.io/goty/gotydoc/internal/sample")

	got := docs.Type(reflect.TypeOf(sample.APIResponse[any]{}))
	if got != "APIResponse is a standard response to our caller." {
		t.Fatalf("generic type doc = %q", got)
	}
}

func sampleDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Join(filepath.Dir(file), "internal", "sample")
}
