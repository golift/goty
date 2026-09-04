package goty_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

type jsonNameOmitempty struct {
	Val string `json:"omitempty"`
}

func TestJSONNameOmitemptyIsNotOption(t *testing.T) {
	t.Parallel()

	out := printType(t, jsonNameOmitempty{})
	if !strings.Contains(out, "omitempty: string;") {
		t.Fatalf("json name \"omitempty\" should not force optional:\n%s", out)
	}
}

type dashNamedField struct {
	Minus string `json:"-,"` //nolint:staticcheck // encoding/json escape for a field named "-"
	Skip  string `json:"-"`
	Keep  string `json:"keep"`
}

func TestJSONDashCommaIsLiteralName(t *testing.T) {
	t.Parallel()

	out := printType(t, dashNamedField{})
	if !strings.Contains(out, `"-": string;`) {
		t.Fatalf("json \"-,\" should emit a field named -:\n%s", out)
	}

	if strings.Contains(out, "Skip") || strings.Contains(out, "skip:") {
		t.Fatalf("json \"-\" field should be omitted:\n%s", out)
	}

	if !strings.Contains(out, "keep: string;") {
		t.Fatalf("keep field missing:\n%s", out)
	}
}

type pathLike struct {
	Dir  string
	File string
}

func (p pathLike) MarshalText() ([]byte, error) {
	return []byte(p.Dir + "/" + p.File), nil
}

type hexHash [4]byte

func (h hexHash) MarshalJSON() ([]byte, error) {
	return []byte(`"deadbeef"`), nil
}

type rawBlob []byte

func (b rawBlob) MarshalJSON() ([]byte, error) {
	return []byte(`{"n":1}`), nil
}

type marshalerHolder struct {
	When time.Time `json:"when"`
	Path pathLike  `json:"path"`
	ID   hexHash   `json:"id"`
	Blob rawBlob   `json:"blob"`
}

func TestMarshalersUseJSONWireType(t *testing.T) {
	t.Parallel()

	out := printType(t, marshalerHolder{})
	for _, want := range []string{
		"when: Date;",
		"path: string;",
		"id: string;",
		"blob?: any;",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}

	if strings.Contains(out, "interface PathLike") {
		t.Fatalf("TextMarshaler struct was expanded as an interface:\n%s", out)
	}

	if strings.Contains(out, "number[]") {
		t.Fatalf("json.Marshaler byte array was emitted as number[]:\n%s", out)
	}
}

func TestJSONStringOptionIgnoresUnsupportedKinds(t *testing.T) {
	t.Parallel()

	type tagged struct {
		Items []int `json:"items,string"` //nolint:staticcheck // json ignores string on slices; goty must too
	}

	out := printType(t, tagged{})
	if !strings.Contains(out, "items?: number[];") {
		t.Fatalf("slice with string option should stay a slice:\n%s", out)
	}
}

func TestJSONStringOptionOnAnonymousStructIsIgnored(t *testing.T) {
	t.Parallel()

	type tagged struct {
		jsonTaggedEmbed `json:",string"` //nolint:staticcheck // json ignores string on structs; goty must too
	}

	out := printType(t, tagged{})
	if strings.Contains(out, "extends string") {
		t.Fatalf("string option on anonymous struct produced extends string:\n%s", out)
	}

	if !strings.Contains(out, "extends JsonTaggedEmbed") {
		t.Fatalf("anonymous struct should still be promoted:\n%s", out)
	}
}

func TestUnexportedAnonymousScalarIsSkipped(t *testing.T) {
	t.Parallel()

	type hiddenInt int //nolint:unused // reached through reflect
	type outer struct {
		hiddenInt //nolint:unused // reached through reflect

		Top string `json:"top"`
	}

	out := printType(t, outer{})
	if strings.Contains(out, "hiddenInt") || strings.Contains(out, "HiddenInt") {
		t.Fatalf("unexported anonymous scalar was emitted:\n%s", out)
	}

	if !strings.Contains(out, "top: string;") {
		t.Fatalf("exported field missing:\n%s", out)
	}
}

func TestWriteStatErrorIsNotFileExists(t *testing.T) {
	t.Parallel()

	goat := goty.NewGoty(nil)
	goat.Parse(omitZeroField{})

	path := statErrorPath(t)
	err := goat.Write(path, false)
	if err == nil {
		t.Fatal("expected stat error")
	}

	if strings.Contains(err.Error(), "file exists") {
		t.Fatalf("stat failure reported as file exists: %v", err)
	}

	if !strings.Contains(err.Error(), "stat file") {
		t.Fatalf("expected stat file error, got: %v", err)
	}
}

// statErrorPath returns a path whose Stat error is not fs.ErrNotExist.
// On Unix, a child of a regular file is ENOTDIR. Windows maps that to
// ERROR_PATH_NOT_FOUND (IsNotExist), so we use an illegal filename instead.
func statErrorPath(t *testing.T) string {
	t.Helper()

	notDir := filepath.Join(t.TempDir(), "not-a-dir")
	err := os.WriteFile(notDir, []byte("x"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(notDir, "out.ts"),
		filepath.Join(t.TempDir(), "foo*bar.ts"),
		filepath.Join(t.TempDir(), "foo|bar.ts"),
		filepath.Join(t.TempDir(), "foo?bar.ts"),
	} {
		_, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return path
		}
	}

	t.Fatal("no path produced a Stat error other than NotExist")

	return ""
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
