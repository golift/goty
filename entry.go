// Package goty provides a way to convert Go structs to TypeScript interfaces.
package goty

import (
	"reflect"
	"slices"

	"golift.io/goty/gotyface"
)

const (
	// DefaultTag is the tag name used to find struct member names.
	DefaultTag = "json"
)

// UsePkgName is the behavior for the package name prefix being appended to the interface name.
type UsePkgName uint8

const (
	// UsePkgNameOnConflict will prefix the typescript interface name with the package name
	// if the interface name is already taken by another struct that had the same name.
	// This is useful when you embed a lot of package.Config{} structs into your own struct.
	// It's still possible to have conflicts, and those will have an integer suffix added.
	// This is the default behavior.
	UsePkgNameOnConflict UsePkgName = iota
	// UsePkgNameNever will never prefix the typescript interface name with the package name.
	// Conflicting names will have an integer suffix added.
	UsePkgNameNever
	// UsePkgNameAlways will always prefix the typescript interface name with the package name.
	// Conflicting names will have an integer suffix added.
	UsePkgNameAlways
)

// Config is the input config for the builder.
type Config struct {
	// Overrides is a map of go types to their typescript type and name.
	// These override the global overrides.
	Overrides Overrides `json:"overrides" toml:"overrides" xml:"overrides" yaml:"overrides"`
	// GlobalOverrides are applied to all structs unless a type-specific override exists.
	GlobalOverrides Override `json:"globalOverrides" toml:"global_overrides" xml:"global-override" yaml:"globalOverrides"`
	// DocHandler is the handler for go/doc comments. Comments are off by default.
	gotyface.Docs `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// Overrides is a map of go types to their typescript override values.
type Overrides map[any]Override

// Override is a struct that contains overrides for either a specific type or for all types (when global).
type Override struct {
	// Namer is a function that can be used to customize the typescript interface name.
	// Use this to add a prefix, suffix or any custom name changes you wish.
	Namer Namer `json:"-" toml:"-" xml:"-" yaml:"-"`
	// Typescript type. ie. string, number, boolean, etc.
	// This has no effect when set inside a global override; it's type specific.
	Type string `json:"type" toml:"type" xml:"type" yaml:"type"`
	// Typescript interface name. This does not work on field names.
	// This has no effect when set inside a global override; it's type specific.
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	// Tag is the tag name to use for the struct member(s). Default is "json".
	Tag string `json:"tag" toml:"tag" xml:"tag" yaml:"tag"`
	// Comment is a comment to add to the typescript interface.
	Comment string `json:"comment" toml:"comment" xml:"comment" yaml:"comment"`
	// Setting optional to true will add a question mark to the typescript name.
	// This has no effect when set inside a global override; it's type specific.
	Optional bool `json:"optional" toml:"optional" xml:"optional" yaml:"optional"`
	// Setting KeepBadChars to true will keep bad characters in the typescript name.
	// These include pretty much all those characters on the number keys on your keyboard.
	KeepBadChars bool `json:"keepBadChars" toml:"keep_bad_chars" xml:"keep-bad-chars" yaml:"keepBadChars"`
	// Setting KeepUnderscores to true will keep underscores in the typescript name.
	// Unlike other characters, underscores are valid. They are still removed by default.
	KeepUnderscores bool `json:"keepUnderscores" toml:"keep_underscores" xml:"keep-underscores" yaml:"keepUnderscores"`
	// Configure the UsePkgName value to control how typescript interface names are generated.
	UsePkgName UsePkgName `json:"usePkgName" toml:"use_pkg_name" xml:"use-pkg-name" yaml:"usePkgName"`
	// By default all typescript interfaces are exported. Set NoExport to true to prevent that.
	NoExport bool `json:"noExport" toml:"no_export" xml:"no-export" yaml:"noExport"`
}

// Namer is an interface that allows external interface naming.
type Namer func(refType reflect.Type, currentName string) string

// NewGoty creates a new Goty instance to build typescript interfaces from go structs.
// If config is nil, it will be initialized to an empty Override.
func NewGoty(config *Config) *Goty {
	return &Goty{
		structNames: make(map[string]bool),
		structTypes: make(map[reflect.Type]*DataStruct),
		config:      config.setup(),
		output:      make([]*DataStruct, 0),
		pkgPaths:    make(map[string]struct{}),
	}
}

// Values returns the output of the builder.
// These are raw values that can be used to generate typescript interfaces.
// Only useful after calling .Enums() and/or .Parse().
func (g *Goty) Values() []*DataStruct {
	return g.output
}

// Pkgs returns the list of package paths that we have parsed.
// This is useful for parsing docs after parsing structs.
func (g *Goty) Pkgs() []string {
	output := make([]string, len(g.pkgPaths))
	idx := 0

	for pkg := range g.pkgPaths {
		output[idx] = pkg
		idx++
	}

	slices.Sort(output)

	return output
}

func getType(fld any) reflect.Type {
	switch t := fld.(type) {
	case reflect.Type:
		return t
	default:
		return reflect.TypeOf(fld)
	}
}

// setup makes sure the config is initialized and all defaults are set in the global override.
func (c *Config) setup() *Config {
	if c == nil {
		c = &Config{}
	}

	c.GlobalOverrides.setup()
	// These are not used in global overrides, make that more obvious.
	c.GlobalOverrides.Type = ""
	c.GlobalOverrides.Name = ""

	if c.Docs == nil {
		c.Docs = gotyface.NoDocs()
	}

	return c
}

// setup makes sure an override has a tag value.
// Other default override options are initially validated and configured here.
func (o *Override) setup() *Override {
	if o.Tag == "" {
		o.Tag = DefaultTag // "json"
	}

	if o.UsePkgName == 0 {
		o.UsePkgName = UsePkgNameOnConflict // explicit.
	}

	if o.Namer == nil {
		o.Namer = func(_ reflect.Type, name string) string {
			return name
		}
	}

	return o
}

// override returns the override for a given type.
// If there is no override for the type, the global override is returned.
func (c *Config) override(typ reflect.Type) *Override {
	for loop, override := range c.Overrides {
		if t := getType(loop); t == typ {
			return override.setup()
		}
	}

	return &c.GlobalOverrides
}
