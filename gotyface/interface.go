// Package gotyface holds the documentation interface for the goty packages.
// It's a simple interface that allows you to pull go/doc comments into typescript as JSDoc.
// Stored in a standalone package to avoid circular imports.
package gotyface

import "reflect"

// Docs allows pulling go/doc comments into typescript as JSDoc.
type Docs interface {
	// Type retrieves documentation for a type
	Type(t reflect.Type) string

	// Member retrieves documentation for a struct or interface member.
	Member(parent reflect.Type, name string) string
}
