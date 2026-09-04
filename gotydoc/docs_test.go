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

	for attempt := range 50 {
		docs := gotydoc.New()
		docs.AddPkgMust(src, pkg)

		got := docs.Type(reflect.TypeFor[sample.Config]())
		if got != "Config is the documented type." {
			t.Fatalf("iteration %d: Config doc = %q", attempt, got)
		}

		got = docs.Member(reflect.TypeFor[sample.Config](), "Name")
		if got != "Name is a field comment." {
			t.Fatalf("iteration %d: Name doc = %q", attempt, got)
		}

		got = docs.Member(reflect.TypeFor[sample.Config](), "Age")
		if got != "Age is a trailing comment." {
			t.Fatalf("iteration %d: Age doc = %q", attempt, got)
		}
	}
}

func TestTypeDocsForGenericInstantiation(t *testing.T) {
	t.Parallel()

	docs := gotydoc.New()
	docs.AddPkgMust(sampleDir(t), "golift.io/goty/gotydoc/internal/sample")

	got := docs.Type(reflect.TypeFor[sample.APIResponse[any]]())
	if got != "APIResponse is a standard response to our caller." {
		t.Fatalf("generic type doc = %q", got)
	}
}

func TestAnonymousFieldDocs(t *testing.T) {
	t.Parallel()

	docs := sampleDocs(t)
	got := docs.Member(reflect.TypeFor[sample.Config](), "Embedded")
	if got != "Nested is an anonymous field with a JSON name." {
		t.Fatalf("anonymous field doc = %q", got)
	}
}

func TestUnexportedTypeDocs(t *testing.T) {
	t.Parallel()

	docs := sampleDocs(t)
	got := docs.Type(sample.SecretHolderType())
	if got != "secretHolder is unexported so go/doc mode 0 would drop it." {
		t.Fatalf("unexported type doc = %q", got)
	}

	got = docs.Member(sample.SecretHolderType(), "Token")
	if got != "Token is a secret field." {
		t.Fatalf("unexported member doc = %q", got)
	}
}

func TestGroupedTypeMemberDocs(t *testing.T) {
	t.Parallel()

	docs := sampleDocs(t)

	got := docs.Type(reflect.TypeFor[sample.GroupedB]())
	if got != "GroupedB is the second grouped type." {
		t.Fatalf("grouped type doc = %q", got)
	}

	got = docs.Member(reflect.TypeFor[sample.GroupedB](), "Beta")
	if got != "Beta is on grouped B." {
		t.Fatalf("grouped member doc = %q", got)
	}

	got = docs.Member(reflect.TypeFor[sample.GroupedA](), "Alpha")
	if got != "Alpha is on grouped A." {
		t.Fatalf("first grouped member doc = %q", got)
	}
}

func sampleDocs(t *testing.T) *gotydoc.Docs {
	t.Helper()

	docs := gotydoc.New()
	docs.AddPkgMust(sampleDir(t), "golift.io/goty/gotydoc/internal/sample")

	return docs
}

func sampleDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Join(filepath.Dir(file), "internal", "sample")
}
