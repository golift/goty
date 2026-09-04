package goty //nolint:testpackage // tests unexported formatDocs.

import (
	"strings"
	"testing"
)

func TestFormatDocsIndentedContinuation(t *testing.T) {
	t.Parallel()

	got := formatDocs(true, "  ", "first line\nsecond line")
	want := "  /**\n   * first line\n   * second line\n   */\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	// The old replacer put a space *before* indent (`\n `+indent+`* `),
	// which breaks tab-indented JSDoc (`\n \t*` vs `\n\t *`).
	got = formatDocs(true, "\t", "first line\nsecond line")
	if !strings.Contains(got, "\t * first line\n\t * second line") {
		t.Fatalf("tab indent misaligned:\n%q", got)
	}
}
