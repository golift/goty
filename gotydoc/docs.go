// Package gotydoc parses Go doc documentation from a vendor folder.
// Provides methods to retrieve the documentation for a type or struct/interface member.
package gotydoc

import (
	"fmt"
	"go/ast"
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

	// Skip _test.go so external tests (package foo_test) cannot overwrite foo.
	ps, err := parser.ParseDir(fset, src, skipTestFiles, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing go/doc in file %s: %w", src, err)
	}

	if p := pickPackage(ps, pkg); p != nil {
		d.pkgs[pkg] = doc.New(p, pkg, 0)
	}

	return nil
}

func skipTestFiles(info fs.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
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

	specs := doct.Decl.Specs
	if len(specs) < 1 {
		return ""
	}

	tspec, ok := specs[0].(*ast.TypeSpec)
	if !ok {
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

func findFieldName(fields []*ast.Field, name string) string {
	for _, dm := range fields {
		if len(dm.Names) > 0 && dm.Names[0].Name == name {
			return strings.TrimSpace(dm.Doc.Text())
		}
	}

	return ""
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
