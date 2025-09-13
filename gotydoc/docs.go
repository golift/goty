// Package gotydoc parses Go doc documentation from a vendor folder.
// Provides methods to retrieve the documentation for a type or struct/interface member.
package gotydoc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
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

	ps, err := parser.ParseDir(fset, src, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing go/doc in file %s: %w", src, err)
	}

	for _, p := range ps {
		d.pkgs[pkg] = doc.New(p, pkg, 0)
	}

	return nil
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

	for _, doct := range pkg.Types {
		if doct.Name == typ.Name() {
			return doct
		}
	}

	return nil
}

// Validate the interface implementation.
var _ gotyface.Docs = &Docs{}
