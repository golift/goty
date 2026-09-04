package goty_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golift.io/goty"
)

type namedBytes []byte

type jsonTaggedEmbed struct {
	Inner string `json:"inner"`
}

type outerWithNamedEmbed struct {
	jsonTaggedEmbed `json:"nested"`

	Top string `json:"top"`
}

type outerWithPromotedEmbed struct {
	jsonTaggedEmbed

	Top string `json:"top"`
}

type omitZeroField struct {
	Name string `json:"name,omitzero"`
}

type byteSliceHolder struct {
	Raw  json.RawMessage `json:"raw"`
	Data []byte          `json:"data"`
	Blob namedBytes      `json:"blob"`
}

func TestAnonymousFieldWithJSONNameIsNested(t *testing.T) {
	t.Parallel()

	out := printType(t, outerWithNamedEmbed{})
	if strings.Contains(out, "extends ") {
		t.Fatalf("named anonymous field was inlined via extends:\n%s", out)
	}

	if !strings.Contains(out, "nested: JsonTaggedEmbed;") {
		t.Fatalf("named anonymous field missing as nested member:\n%s", out)
	}
}

func TestUnexportedAnonymousEmbedIsPromoted(t *testing.T) {
	t.Parallel()

	out := printType(t, outerWithPromotedEmbed{})
	if !strings.Contains(out, "interface OuterWithPromotedEmbed extends JsonTaggedEmbed") {
		t.Fatalf("unexported anonymous embed was dropped:\n%s", out)
	}
}

func TestOmitZeroIsOptional(t *testing.T) {
	t.Parallel()

	out := printType(t, omitZeroField{})
	if !strings.Contains(out, "name?: string;") {
		t.Fatalf("omitzero field should be optional:\n%s", out)
	}
}

func TestByteSlicesAreStrings(t *testing.T) {
	t.Parallel()

	out := printType(t, byteSliceHolder{})
	for _, want := range []string{
		"raw?: any;",
		"data?: string;",
		"blob?: string;",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}

	if strings.Contains(out, "number[]") {
		t.Fatalf("byte slice emitted as number[]:\n%s", out)
	}
}

type jsonStringTag struct {
	Count int  `json:"count,string,omitempty"`
	Flag  bool `json:"flag,string"`
}

type jsonNameString struct {
	Val int `json:"string"`
}

func TestJSONStringTagIsString(t *testing.T) {
	t.Parallel()

	out := printType(t, jsonStringTag{})
	for _, want := range []string{
		"count?: string;",
		"flag: string;",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestJSONNameStringIsNotStringOption(t *testing.T) {
	t.Parallel()

	out := printType(t, jsonNameString{})
	if !strings.Contains(out, "string: number;") {
		t.Fatalf("json name \"string\" should not force the string option:\n%s", out)
	}
}

func TestWriteStatErrorIsNotFileExists(t *testing.T) {
	t.Parallel()

	goat := goty.NewGoty(nil)
	goat.Parse(omitZeroField{})

	missingDir := filepath.Join(t.TempDir(), "nope", "out.ts")

	err := goat.Write(missingDir, false)
	if err == nil {
		t.Fatal("expected error writing into a missing directory")
	}

	if strings.Contains(err.Error(), "file exists") {
		t.Fatalf("stat failure reported as file exists: %v", err)
	}
}

func TestEmptyEnumDoesNotPanic(t *testing.T) {
	t.Parallel()

	goat := goty.NewGoty(nil)
	goat.Enums(nil, []goty.Enum{})
}

func TestWriteOverwriteFalse(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "out.ts")
	err := os.WriteFile(path, []byte("nope"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	goat := goty.NewGoty(nil)
	goat.Parse(omitZeroField{})

	err = goat.Write(path, false)
	if err == nil {
		t.Fatal("expected file exists error")
	}
}

func printType(t *testing.T, val any) string {
	t.Helper()

	goat := goty.NewGoty(nil)
	goat.Parse(val)

	var buf bytes.Buffer
	for _, s := range goat.Values() {
		s.Print("", &buf)
	}

	return buf.String()
}
