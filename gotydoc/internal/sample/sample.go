// Package sample holds types used to test gotydoc package selection.
package sample

import "reflect"

// Config is the documented type.
type Config struct {
	// Nested is an anonymous field with a JSON name.
	Embedded `json:"nested"`

	// Name is a field comment.
	Name string
	Age  int // Age is a trailing comment.
}

// Embedded is nested into Config.
type Embedded struct {
	Inner string
}

// APIResponse is a standard response to our caller.
type APIResponse[T any] struct {
	Status string
	Msg    T
}

type (
	// GroupedA is the first grouped type.
	GroupedA struct {
		// Alpha is on grouped A.
		Alpha string
	}
	// GroupedB is the second grouped type.
	GroupedB struct {
		// Beta is on grouped B.
		Beta string
	}
)

// secretHolder is unexported so go/doc mode 0 would drop it.
type secretHolder struct {
	// Token is a secret field.
	Token string
}

// SecretHolderType is reflect.Type of the unexported secretHolder.
func SecretHolderType() reflect.Type {
	return reflect.TypeFor[secretHolder]()
}
