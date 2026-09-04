package goty_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
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
		"raw: any;",
		"data: string;",
		"blob: string;",
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
		"id: any;",
		"blob: any;",
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

type money struct{}

func (money) MarshalJSON() ([]byte, error) {
	return []byte(`"5"`), nil
}

type wallet struct {
	Balance money `json:"balance"`
}

func TestMarshalerFieldTypeIsOrderIndependent(t *testing.T) {
	t.Parallel()

	for _, vals := range [][]any{
		{money{}, wallet{}},
		{wallet{}, money{}},
	} {
		out := printTypes(t, vals...)
		if !strings.Contains(out, "balance: any;") {
			t.Fatalf("Parse(%v) should type the marshaler field as any:\n%s", vals, out)
		}

		if strings.Contains(out, "balance: Money;") {
			t.Fatalf("Parse(%v) used the struct name instead of the wire type:\n%s", vals, out)
		}
	}
}

type tags []string

func (t *tags) MarshalJSON() ([]byte, error) {
	if t == nil || *t == nil {
		return []byte("null"), nil
	}

	raw, err := json.Marshal([]string(*t))
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}

	return raw, nil
}

type tagHolder struct {
	Tags tags   `json:"tags"`
	List []tags `json:"list"`
}

func TestJSONMarshalerSliceIsAny(t *testing.T) {
	t.Parallel()

	out := printType(t, tagHolder{})
	for _, want := range []string{
		"tags: any;",
		"list: any[];",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

type sDual []string

func (t *sDual) MarshalJSON() ([]byte, error) {
	if t == nil || *t == nil {
		return []byte("null"), nil
	}

	raw, err := json.Marshal([]string(*t))
	if err != nil {
		return nil, fmt.Errorf("marshal sDual: %w", err)
	}

	return raw, nil
}

func (*sDual) MarshalText() ([]byte, error) {
	return []byte("text"), nil
}

type dualHolder struct {
	D  sDual   `json:"d"`
	LD []sDual `json:"ld"`
}

func TestJSONMarshalerDoesNotFallThroughToText(t *testing.T) {
	t.Parallel()

	out := printType(t, dualHolder{})
	for _, want := range []string{
		"d: any;",
		"ld: any[];",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}

	if strings.Contains(out, "d: string;") || strings.Contains(out, "ld: string[];") {
		t.Fatalf("json.Marshaler fell through to TextMarshaler:\n%s", out)
	}
}

type date struct {
	Y int `json:"y"`
	M int `json:"m"`
}

type cal struct {
	date

	Z string `json:"z"`
}

func TestNamedDateEmbedStillExtends(t *testing.T) {
	t.Parallel()

	out := printType(t, cal{})
	if !strings.Contains(out, "interface Cal extends Date") {
		t.Fatalf("embed named Date should still extend:\n%s", out)
	}
}

type timeEmbed struct {
	time.Time

	X string `json:"x"`
}

type timeEmbedHolder struct {
	E timeEmbed `json:"e"`
}

func TestEmbeddedTimeIsWireTypeNotInterface(t *testing.T) {
	t.Parallel()

	out := printType(t, timeEmbed{})
	if strings.Contains(out, "interface TimeEmbed") ||
		strings.Contains(out, "Time: Date") ||
		strings.Contains(out, "x: string") {
		t.Fatalf("embedded time.Time was expanded:\n%s", out)
	}

	if !strings.Contains(out, "type TimeEmbed = any") {
		t.Fatalf("root time embed should be an any alias:\n%s", out)
	}
}

func TestEmbeddedTimeFieldIsAny(t *testing.T) {
	t.Parallel()

	out := printType(t, timeEmbedHolder{})
	if !strings.Contains(out, "e: any;") {
		t.Fatalf("embedded time.Time field should be any:\n%s", out)
	}

	if strings.Contains(out, "x: string") || strings.Contains(out, "interface TimeEmbed {") {
		t.Fatalf("outer marshaler was expanded:\n%s", out)
	}
}

type ownMarshal struct {
	time.Time

	Name string `json:"name"`
}

func (o ownMarshal) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(struct {
		TS   time.Time `json:"ts"`
		Name string    `json:"name"`
	}{TS: o.Time, Name: o.Name})
	if err != nil {
		return nil, fmt.Errorf("marshal ownMarshal: %w", err)
	}

	return raw, nil
}

type ownMarshalHolder struct {
	E ownMarshal `json:"e"`
}

func TestOwnMarshalJSONOnTimeEmbedIsNotDate(t *testing.T) {
	t.Parallel()

	out := printType(t, ownMarshalHolder{})
	if strings.Contains(out, "e: Date") {
		t.Fatalf("own MarshalJSON was typed as Date:\n%s", out)
	}

	if !strings.Contains(out, "e: any;") {
		t.Fatalf("own object marshaler should stay any:\n%s", out)
	}

	root := printType(t, ownMarshal{})
	if strings.Contains(root, "interface OwnMarshal") || strings.Contains(root, "name: string") {
		t.Fatalf("root object marshaler was expanded:\n%s", root)
	}

	if !strings.Contains(root, "type OwnMarshal = any") {
		t.Fatalf("root object marshaler should be an any alias:\n%s", root)
	}
}

type ptrTimeEmbed struct {
	*time.Time

	X string `json:"x"`
}

type ptrTimeEmbedHolder struct {
	E ptrTimeEmbed `json:"e"`
}

func TestPointerTimeEmbedFieldIsAny(t *testing.T) {
	t.Parallel()

	out := printType(t, ptrTimeEmbedHolder{})
	if !strings.Contains(out, "e: any;") {
		t.Fatalf("*time.Time embed field should be any:\n%s", out)
	}
}

func TestRootTextMarshalerIsAlias(t *testing.T) {
	t.Parallel()

	out := printType(t, pathLike{})
	if strings.Contains(out, "interface PathLike") || strings.Contains(out, "Dir") {
		t.Fatalf("root TextMarshaler was expanded:\n%s", out)
	}

	if !strings.Contains(out, "type PathLike = string") {
		t.Fatalf("root TextMarshaler should be a string alias:\n%s", out)
	}
}

type objectInt int

func (objectInt) MarshalJSON() ([]byte, error) {
	return []byte(`{"n":1}`), nil
}

func TestJSONStringOptionDoesNotOverrideMarshaler(t *testing.T) {
	t.Parallel()

	type holder struct {
		N objectInt `json:"n,string"`
	}

	out := printType(t, holder{})
	if strings.Contains(out, "n: string;") || strings.Contains(out, "n?: string;") {
		t.Fatalf("string tag overrode json.Marshaler:\n%s", out)
	}

	if !strings.Contains(out, "n: any;") {
		t.Fatalf("marshaler object should stay any:\n%s", out)
	}
}

func TestJSONStringOptionIgnoresUnsupportedKinds(t *testing.T) {
	t.Parallel()

	type tagged struct {
		Items []int `json:"items,string"` //nolint:staticcheck // json ignores string on slices; goty must too
	}

	out := printType(t, tagged{})
	if !strings.Contains(out, "items: number[];") {
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

func TestSlicesAndMapsFollowJSONOmitempty(t *testing.T) {
	t.Parallel()

	type holder struct {
		Items []int          `json:"items"`
		More  []int          `json:"more,omitempty"`
		Pair  [2]int         `json:"pair"`
		Meta  map[string]int `json:"meta"`
		Skip  map[string]int `json:"skip,omitempty"`
		Ptr   *[]int         `json:"ptr"`
	}

	out := printType(t, holder{})
	for _, want := range []string{
		"items: number[];",
		"more?: number[];",
		"pair: number[];",
		"meta: Record<string, number>;",
		"skip?: Record<string, number>;",
		"ptr?: number[];",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestDurationIsNumber(t *testing.T) {
	t.Parallel()

	type holder struct {
		Wait time.Duration  `json:"wait"`
		Idle *time.Duration `json:"idle"`
	}

	out := printType(t, holder{})
	for _, want := range []string{
		"wait: number;",
		"idle?: number;",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}

	if strings.Contains(out, "wait: string") || strings.Contains(out, "interface Duration") {
		t.Fatalf("Duration should be a nanosecond number:\n%s", out)
	}
}

func TestURLIsJSONStructNotString(t *testing.T) {
	t.Parallel()

	type holder struct {
		Home url.URL `json:"home"`
	}

	out := printType(t, holder{})
	if strings.Contains(out, "home: string;") || strings.Contains(out, "type URL = string") {
		t.Fatalf("url.URL is not a TextMarshaler; JSON is a struct:\n%s", out)
	}

	if !strings.Contains(out, "home: URL;") {
		t.Fatalf("url.URL field should use the struct type:\n%s", out)
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

	return printTypes(t, val)
}

func printTypes(t *testing.T, vals ...any) string {
	t.Helper()

	goat := goty.NewGoty(nil)
	goat.Parse(vals...)

	var buf bytes.Buffer
	for _, s := range goat.Values() {
		s.Print("", &buf)
	}

	return buf.String()
}
