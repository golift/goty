// Package sample holds types used to test gotydoc package selection.
package sample

// Config is the documented type.
type Config struct {
	// Name is a field comment.
	Name string
}

// APIResponse is a standard response to our caller.
type APIResponse[T any] struct {
	Status string
	Msg    T
}
