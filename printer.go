package goty

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Header is printed before anything else.
const Header = `/* Auto-generated. DO NOT EDIT. Generator: https://golift.io/goty
 * Edit the source code and run goty again to make updates.` + "\n */\n\n"

// ErrNoStructs is returned if no structs are found to print or write.
var ErrNoStructs = errors.New("no structs to write, run Parse() first")

// Print prints all the structs in the output to stdout as typescript interfaces.
func (g *Goty) Print() {
	g.print(os.Stdout)
}

// Write writes all the structs in the output to a file as typescript interfaces.
// If the file exists and overwrite is false, returns an error.
func (g *Goty) Write(fileName string, overwrite bool) error {
	if len(g.output) == 0 {
		return ErrNoStructs
	}

	if _, err := os.Stat(fileName); !os.IsNotExist(err) && !overwrite {
		return fmt.Errorf("file exists: %s: %w", fileName, os.ErrExist)
	}

	file, err := os.Create(fileName) //nolint:gosec // user chooses their own demise.
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	g.print(file)

	return nil
}

// Print a struct as a typescript interface to an io.Writer.
func (s *DataStruct) Print(indent string, output io.Writer) {
	golangRef := "\n * @see golang: <" + s.GoName + ">"
	doc := formatDocs(false, indent, s.doc.Type(s.Type), s.ovr.Comment)
	fmt.Fprintln(output, `/**`+doc+golangRef+"\n"+` */`)

	if len(s.Elements) > 0 {
		s.printElements(indent, output)
		return
	}

	exported := "export "
	if s.ovr.NoExport {
		exported = ""
	}

	if len(s.Extends) > 0 {
		fmt.Fprintf(output, indent+exported+`interface %s extends %s {`,
			s.Name, strings.Join(s.Extends, `, `))
	} else {
		fmt.Fprint(output, indent+exported+`interface `+s.Name+` {`)
	}

	if len(s.Members) > 0 {
		fmt.Fprintln(output)
	}

	for _, m := range s.Members {
		m.Print(indent+`  `, output)
	}

	fmt.Fprintln(output, indent+"};\n")
}

// Print prints a struct member as a typescript interface member.
func (m *StructMember) Print(indent string, output io.Writer) {
	optional := ""
	if m.Optional {
		optional = "?"
	}

	doc := formatDocs(true, indent, m.doc.Member(m.parent.Type, m.Member.Name), m.ovr.Comment)

	if m.Members == nil {
		fmt.Fprintln(output, doc+indent+m.Name+optional+`: `+m.Type+`;`)
		return
	}

	if optional = ""; m.Optional {
		optional = "null | "
	}

	fmt.Fprintln(output, indent+m.Name+`: `+optional+`{`)

	for _, m := range m.Members {
		m.Print(indent+`  `, output)
	}

	fmt.Fprintln(output, indent+`};`)
}

func (g *Goty) print(output io.Writer) {
	if len(g.output) == 0 {
		panic(ErrNoStructs)
	}

	fmt.Fprint(output, Header)

	for _, s := range g.output {
		s.Print("", output)
	}

	if len(g.pkgPaths) < 1 {
		return
	}

	fmt.Fprintln(output, "// Packages parsed:")

	for idx, pkg := range g.Pkgs() {
		fmt.Fprintf(output, "// %3d. %s\n", idx+1, pkg)
	}
}

// formatDocs formats the documentation for an interface and an interface member.
// It wraps the documentation in JSDoc format if wrap is true.
func formatDocs(wrap bool, indent, doc string, extra ...string) string {
	for _, e := range extra {
		if e != "" {
			doc += strings.Trim(e, "\n")
		}
	}

	if doc == "" {
		return ""
	}

	output := ""

	if wrap {
		output = indent + "/**\n"
	}

	// sorry. :( it tries to make pretty JSDoc.
	output += strings.ReplaceAll(indent+" * "+doc, "\n", "\n "+indent+"* ")

	if wrap {
		return output + "\n" + indent + " */\n"
	}

	return "\n" + output
}

func (s *DataStruct) printElements(indent string, output io.Writer) {
	longest := 0
	for _, v := range s.Elements {
		if len(v.Name) > longest {
			longest = len(v.Name)
		}
	}

	exported := "export "
	if s.ovr.NoExport {
		exported = ""
	}

	fmt.Fprintln(output, indent+exported+`enum `+s.Name+` {`)
	// We use the formatter to align the enum values visually.
	formatter := fmt.Sprintf("%s  %%-%ds = %%s,\n", indent, longest)
	for _, v := range s.Elements {
		fmt.Fprintf(output, formatter, v.Name, v.Value)
	}

	fmt.Fprintln(output, indent+"};\n")
}
