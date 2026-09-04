package gotydoc_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golift.io/goty/gotydoc"
)

func TestAddPkgSkipsWrongGOOSAndTests(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeGo(t, dir, "keep.go", `package p

// Keep is compiled on every GOOS.
type Keep struct{}
`)
	writeGo(t, dir, "x_plan9.go", `package p

// Plan9Only must not load except on plan9.
type Plan9Only struct{}
`)
	writeGo(t, dir, "never.go", `//go:build ignore

package p

// NeverOnly is excluded by a build tag.
type NeverOnly struct{}
`)
	writeGo(t, dir, "keep_test.go", `package p

// TestOnly is a test file.
type TestOnly struct{}
`)

	docs := gotydoc.New()
	docs.AddPkgMust(dir, "example.com/p")

	got := docs.TypeNames("example.com/p")
	slices.Sort(got)

	if !slices.Contains(got, "Keep") {
		t.Fatalf("types = %v, missing Keep", got)
	}

	if slices.Contains(got, "NeverOnly") {
		t.Fatalf("build-tagged file was parsed: %v", got)
	}

	if slices.Contains(got, "TestOnly") {
		t.Fatalf("test file was parsed: %v", got)
	}

	if runtime.GOOS != "plan9" && slices.Contains(got, "Plan9Only") {
		t.Fatalf("plan9 file was parsed on %s: %v", runtime.GOOS, got)
	}
}

func writeGo(t *testing.T, dir, name, src string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
