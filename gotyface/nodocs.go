package gotyface

import "reflect"

// NoDocs is a fake doc handler that returns empty strings for all methods.
// This is useful when you don't want to pull in go/doc comments.
//
//nolint:revive // This is on purpose.
func NoDocs() *noDocs {
	return &noDocs{}
}

type noDocs struct{}

func (n *noDocs) Type(_ reflect.Type) string {
	return ""
}

func (n *noDocs) Member(_ reflect.Type, _ string) string {
	return ""
}

// Validate the interface implementation.
var _ DocHandler = &noDocs{}
