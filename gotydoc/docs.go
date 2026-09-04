// Package gotydoc parses Go doc documentation from a vendor folder.
// Provides methods to retrieve the documentation for a type or struct/interface member.
package gotydoc

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"golift.io/goty/gotyface"
)

/* Lifted from here:
 * https://github.com/csweichel/bel/blob/66b16680e6929086857458da3fa411c16d14d871/doc.go
 * MIT license.
 * Copyright (c) 2020 Christoph Weichel
 * With some changes.
 */

// Docs provides Go doc documentation from a vendor folder.
type Docs struct {
	pkgs map[string]*doc.Package
}

// New creates a new doc handler ready to add packages.
func New() *Docs {
	return &Docs{pkgs: make(map[string]*doc.Package)}
}

// AddPkg adds a package to the handler's index.
// src is the path to the package.
// pkg is the name of the package. Must be full package name.
func (d *Docs) AddPkg(src string, pkg string) error {
	fset := token.NewFileSet()

	// Skip _test.go and files that fail GOOS/GOARCH/build tags so a Windows
	// file cannot overwrite docs from the file that actually compiles here.
	ps, err := parser.ParseDir(fset, src, packageFiles(src), parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing go/doc in file %s: %w", src, err)
	}

	if p := pickPackage(ps, pkg); p != nil {
		// AllDecls so unexported embedded structs (which encoding/json promotes)
		// still have type and field comments available.
		d.pkgs[pkg] = doc.New(p, pkg, doc.AllDecls)
	}

	return nil
}

func packageFiles(dir string) func(fs.FileInfo) bool {
	return func(info fs.FileInfo) bool {
		if strings.HasSuffix(info.Name(), "_test.go") {
			return false
		}

		ok, err := build.Default.MatchFile(dir, info.Name())

		return err == nil && ok
	}
}

// pickPackage chooses the production package for an import path.
// parser.ParseDir groups files by the `package` clause, so a directory can
// contain both `foo` and `foo_test`. Ranging that map is non-deterministic;
// the last package used to win and silently drop all type docs.
//
//nolint:staticcheck // parser.ParseDir still returns *ast.Package.
func pickPackage(pkgs map[string]*ast.Package, importPath string) *ast.Package {
	if pkg, ok := pkgs[filepath.Base(importPath)]; ok {
		return pkg
	}

	names := make([]string, 0, len(pkgs))
	for name := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	if len(names) == 0 {
		return nil
	}

	return pkgs[names[0]]
}

// AddPkgMust adds a package to the handler's index like AddPkg but panics if there is an error.
// See AddPkg for more details.
func (d *Docs) AddPkgMust(src string, pkg string) *Docs {
	err := d.AddPkg(src, pkg)
	if err != nil {
		panic(err)
	}

	return d
}

// Add multiple packages to the handler's index.
// Vendor folder should contain full-module name paths.
// ie. They begin with github.com/username.
// Running `go mod vendor` is a good way to create this folder.
func (d *Docs) Add(vendorFolder string, pkg ...string) error {
	for _, p := range pkg {
		err := d.AddPkg(filepath.Join(vendorFolder, p), p)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddMust adds a package to the handler like Add but panics if there is an error.
// See Add for more details.
func (d *Docs) AddMust(vendorFolder string, pkg ...string) *Docs {
	err := d.Add(vendorFolder, pkg...)
	if err != nil {
		panic(err)
	}

	return d
}

// Type retrieves documentation for a top-level type using the handler's index.
func (d *Docs) Type(typ reflect.Type) string {
	doct := d.findDoc(typ)
	if doct == nil {
		return ""
	}

	return strings.TrimSpace(doct.Doc)
}

// Member retrieves documentation for a struct member using the handler's index.
func (d *Docs) Member(parent reflect.Type, name string) string {
	doct := d.findDoc(parent)
	if doct == nil {
		return ""
	}

	tspec := namedTypeSpec(doct)
	if tspec == nil {
		return ""
	}

	switch typ := tspec.Type.(type) {
	case *ast.InterfaceType:
		return findFieldName(typ.Methods.List, name)
	case *ast.StructType:
		return findFieldName(typ.Fields.List, name)
	default:
		return ""
	}
}

// namedTypeSpec returns the TypeSpec whose name matches doct.
// go/doc usually synthesizes a one-spec GenDecl, but we still search by name.
func namedTypeSpec(doct *doc.Type) *ast.TypeSpec {
	if doct.Decl == nil {
		return nil
	}

	for _, spec := range doct.Decl.Specs {
		tspec, ok := spec.(*ast.TypeSpec)
		if ok && tspec.Name != nil && tspec.Name.Name == doct.Name {
			return tspec
		}
	}

	return nil
}

func findFieldName(fields []*ast.Field, name string) string {
	for _, astField := range fields {
		if !fieldHasName(astField, name) {
			continue
		}

		if astField.Doc != nil {
			if text := strings.TrimSpace(astField.Doc.Text()); text != "" {
				return text
			}
		}

		if astField.Comment != nil {
			return strings.TrimSpace(astField.Comment.Text())
		}

		return ""
	}

	return ""
}

func fieldHasName(field *ast.Field, name string) bool {
	if len(field.Names) == 0 {
		return embeddedFieldName(field.Type) == name
	}

	for _, ident := range field.Names {
		if ident.Name == name {
			return true
		}
	}

	return false
}

func embeddedFieldName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return embeddedFieldName(typed.X)
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.IndexExpr:
		return embeddedFieldName(typed.X)
	case *ast.IndexListExpr:
		return embeddedFieldName(typed.X)
	case *ast.ParenExpr:
		return embeddedFieldName(typed.X)
	default:
		return ""
	}
}

func (d *Docs) findDoc(typ reflect.Type) *doc.Type {
	pkg, ok := d.pkgs[typ.PkgPath()]
	if !ok {
		return nil
	}

	want := instantiatedName(typ.Name())
	for _, doct := range pkg.Types {
		if doct.Name == want {
			return doct
		}
	}

	return nil
}

// instantiatedName strips type parameters from a reflect type name.
// APIResponse[any] is named "APIResponse[interface {}]" at runtime.
func instantiatedName(name string) string {
	base, _, _ := strings.Cut(name, "[")

	return base
}

// Validate the interface implementation.
var _ gotyface.Docs = &Docs{}
