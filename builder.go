package goty

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"golift.io/goty/gotyface"
)

// Goty is the main struct for the builder.
// It's used to build typescript interfaces from go structs.
type Goty struct {
	// config is the config input for the builder.
	// Everything else in this struct is output.
	config *Config
	// structNames is a list of unique typescript interface names.
	// We keep track by name to ensure every interface gets a unique name.
	structNames map[string]bool
	// structTypes is a map of struct types to their typescript interface names.
	// We keep track by type so if a struct is embedded twice, it only gets saved once.
	structTypes map[reflect.Type]*DataStruct
	// pkgPaths is a list of package paths that we have parsed.
	// This is so you can parse docs after you parse structs.
	pkgPaths map[string]struct{}
	// output is what we build up as we parse the input struct(s).
	// We use a slice to preserve the order of the input structs.
	// Otherwise we could just use the structTypes map.
	output []*DataStruct
}

// DataStruct is the internal representation of a typescript interface
// that we build up as we parse the input struct(s).
type DataStruct struct {
	// Type is the go struct type that we are building the typescript interface for.
	Type reflect.Type
	// doc is the documentation handler to find docs for this struct and its members.
	doc gotyface.Docs
	// Overrides for this struct.
	ovr *Override
	// Name is generated from the struct name and package path, or from an override.
	// name is also the "type" where this is a member of a typescript interface.
	Name string
	// GoName is the full import path and name of the struct.
	GoName string
	// Members is a list of members in the struct, each with their own configuration.
	// If this is set there are no elements.
	Members []*StructMember
	// Elements is a map of enum values to their names.
	// If this is set there are no members.
	Elements []*Enum
	// Extends is a list of struct names that this struct extends.
	// This happens when a struct is anonymously embedded in another struct.
	Extends []string
}

// StructMember is the internal representation of a member of a typescript interface.
type StructMember struct {
	// doc, Member and parent are used to find the documentation for the member.
	doc    gotyface.Docs
	parent *DataStruct
	ovr    *Override
	// Name is the name of the member.
	Name string
	// Type is the typescript type of the member. Usually string, number, boolean, etc.
	Type string
	// Members is a list of members in this member if it's an anonymous struct.
	Members []*StructMember
	// Member is the struct field that we are building the typescript interface for.
	Member reflect.StructField
	// Optional is true if the member is optional.
	Optional bool
}

// Enum is used as an input to the Enum method.
// Use this to add an enum to the builder.
// Enums should be added before parsing the structs that use them.
// Do not mix enums, add each enum separately.
// Enums have no type. But maybe they could?
type Enum struct {
	// Value of the enum.
	Value any
	// Name of the enum.
	Name string
}

// Parse parses a struct and adds it to the builder.
func (g *Goty) Parse(elems ...any) *Goty {
	for _, elem := range elems {
		if elem == nil {
			continue
		}

		typ := getType(elem)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			panic("expected a struct, got " + typ.String())
		}

		g.parseStruct(typ)
	}

	return g
}

// Enums adds enums to the builder. The input is enum name and value pairs.
// Add enums before parsing the structs that use them.
func (g *Goty) Enums(enums ...[]Enum) *Goty {
	for _, enum := range enums {
		if enum != nil {
			g.enum(enum)
		}
	}

	return g
}

func (g *Goty) enum(enum []Enum) {
	var typ reflect.Type
	// Find the name of the Enum by looking at the type of the first value.
	for _, e := range enum {
		typ = reflect.TypeOf(e.Value)
		break
	}

	data := &DataStruct{
		Elements: make([]*Enum, len(enum)),
		doc:      g.config,
		Type:     typ,
		Name:     g.getStructName(typ),
		GoName:   typ.PkgPath() + "." + typ.Name(),
		ovr:      g.config.override(typ),
	}

	if g.structNames[data.Name] {
		panic("cannot find a suitable struct name for " +
			data.Type.PkgPath() + "." + data.Type.Name() + ": " + data.Name)
	}

	// Convert the enum values to typescript values using json Marshaller.
	for idx, enum := range enum {
		str, err := json.Marshal(enum.Value)
		if err != nil {
			panic("cannot marshal enum value: " + err.Error())
		}

		data.Elements[idx] = &Enum{Name: enum.Name, Value: string(str)}
	}

	g.structTypes[data.Type] = data
	g.structNames[data.Name] = true
	g.output = append(g.output, data)
}

// parseStruct adds a struct to the builder if it doesn't already exist.
// It will also add a unique suffix if the struct name is already taken.
// It returns the struct data that is used as a typescript interface.
func (g *Goty) parseStruct(elem reflect.Type) *DataStruct {
	if v, ok := g.structTypes[elem]; ok {
		return v
	}

	name := g.getStructName(elem)
	if g.structNames[name] {
		panic("cannot find a suitable struct name for " +
			elem.PkgPath() + "." + elem.Name() + ": " + name)
	}

	data := &DataStruct{
		Name:    name,
		Type:    elem,
		GoName:  elem.PkgPath() + "." + elem.Name(),
		Members: make([]*StructMember, 0),
		doc:     g.config,
		ovr:     g.config.override(elem),
	}

	// Add the struct to the builder if it has a name.
	// No name means it's embedded and all its members get added to the parent.
	if name != "" {
		g.structTypes[elem] = data
		g.structNames[name] = true
		g.output = append(g.output, data)
		g.pkgPaths[elem.PkgPath()] = struct{}{}
	}

	g.addStructMembers(data, elem)

	return data
}

// addStructMembers loops through the fields of a struct and adds them to the builder.
func (g *Goty) addStructMembers(data *DataStruct, field reflect.Type) {
	for idx := range field.NumField() { // Loop each struct member
		elem := field.Field(idx)
		ovr := g.config.override(elem.Type)
		tagval := strings.Split(elem.Tag.Get(ovr.Tag), ",")

		name := tagval[0]
		if name == "-" || !elem.IsExported() {
			continue
		} else if name == "" {
			name = elem.Name
		}

		member := &StructMember{
			Name:     g.stripBadChars(name, elem.Type),
			doc:      g.config, // hard to attach this later.
			Member:   elem,
			parent:   data,
			ovr:      ovr,
			Optional: ovr.Optional,
			Type:     ovr.Type,
		}

		if member.Type == "" {
			// We only parse the member if it didn't have a type override.
			member.Type, member.Optional = g.parseMember(data, elem.Type, member)
		}

		for _, tag := range tagval {
			if tag == "omitempty" {
				member.Optional = true
			}
		}

		data.addMember(member)
	}
}

// addMember adds a member to the struct.
// If the member is an anonymous struct, it extends the parent struct.
// Otherwise, it adds the member to the struct.
func (d *DataStruct) addMember(member *StructMember) {
	if member.Member.Anonymous && (member.Member.Type.Kind() == reflect.Struct ||
		member.Member.Type.Kind() == reflect.Ptr && member.Member.Type.Elem().Kind() == reflect.Struct) {
		d.Extends = append(d.Extends, member.Type)
	} else {
		d.Members = append(d.Members, member)
	}
}

// parseMember returns the typescript type for a given go type.
// It also returns a boolean indicating if the type is optional.
// Fully recursive.
//
//nolint:cyclop // This is a complex function, but really it's not that bad.
func (g *Goty) parseMember(parent *DataStruct, field reflect.Type, member *StructMember) (string, bool) {
	if g.structTypes[field] != nil {
		// This happens when there was a matching enum provided.
		return g.structTypes[field].Name, false
	}

	switch field.Kind() {
	case reflect.Ptr:
		s, _ := g.parseMember(parent, field.Elem(), member)
		return s, true
	case reflect.Struct:
		return g.checkStruct(field, member), false
	case reflect.Array, reflect.Slice:
		return g.parseSlice(parent, field, member), true
	case reflect.Map:
		return g.parseMap(parent, field, member), true
	case reflect.Bool:
		return "boolean", false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Uintptr:
		return "number", false
	case reflect.String:
		return "string", false
	case reflect.Interface:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	case reflect.Complex64, reflect.Complex128:
		fallthrough
	case reflect.Invalid:
		fallthrough
	default:
		return "any", true
	}
}

// checkStruct provides some logic to detect special struct types.
// Those include time.Time/Duration and embedded structs. Do we need others?
func (g *Goty) checkStruct(field reflect.Type, member *StructMember) string {
	switch field.String() {
	case "time.Time":
		return "Date"
	case "time.Duration":
		return "number"
	}

	structMember := g.parseStruct(field)
	if structMember.Name == "" { // Embedded struct.
		member.Members = append(member.Members, structMember.Members...)
		return "" // embedded structs don't have names; deal with it.
	}

	return structMember.Name
}

// parseSlice returns the typescript type for a given go slice.
func (g *Goty) parseSlice(parent *DataStruct, field reflect.Type, member *StructMember) string {
	// Go marshalls a byte slice into a base64 encoded string.
	if field.String() == "[]uint8" {
		return "string"
	}

	name, optional := g.parseMember(parent, field.Elem(), member)
	if optional && g.config.override(field).NullSlicePointers {
		name = "(null | " + name + ")"
	}

	// This doesn't really produce valid typescript. Any ideas?
	// size := ""
	// if field.Kind() == reflect.Array {
	// 	size = strconv.Itoa(field.Len())
	// }
	// return name + "[" + size + "]"
	return name + "[]"
}

// parseMap returns the typescript type for a given go map.
func (g *Goty) parseMap(parent *DataStruct, field reflect.Type, member *StructMember) string {
	// Parse both sides of the map.
	key, keyOptional := g.parseMember(parent, field.Key(), member)
	val, valOptional := g.parseMember(parent, field.Elem(), member)

	if keyOptional {
		key = "null | " + key
	}

	if valOptional {
		val = "null | " + val
	}

	return "Record<" + key + ", " + val + ">"
}

// getStructName returns a unique, capitalized name for a struct by appending a number to the end.
// The returned name is used as the interface name in the generated typescript code.
// If the struct has a name override, that is used instead.
// This is where we can add logic to manipulate the interface names. You could do things like:
// - Remove the package name; using only the struct name/tag value.
// - Use the struct name without the package name.
// - Make the name lowercase, uppercase, camelcase or snake_case.
func (g *Goty) getStructName(elem reflect.Type) string {
	ovr := g.config.override(elem)
	name := ovr.Namer(elem, capitalizeFirstLetter(elem.Name()))
	name = g.stripBadChars(name, elem)
	pkgParts := strings.Split(elem.PkgPath(), "/")

	if ovr.UsePkgName == UsePkgNameAlways ||
		(g.structNames[name] && ovr.UsePkgName == UsePkgNameOnConflict) {
		// We have to pass the original element name back in here so any name changes are repeated.
		name = ovr.Namer(elem, capitalizeFirstLetter(pkgParts[len(pkgParts)-1])+elem.Name())
	}

	// Name is elem name, or base pkg name + elem name. If there is an override, use it.
	if ovr.Name != "" {
		name = ovr.Name
	} else {
		name = g.stripBadChars(name, elem)
	}

	// The base name is the name of the struct without any suffix.
	// Usually there will not be a suffix added, so at this point we
	// have the name. It came from either an override or the pkgName + structName.
	base := name

	// Find a unique name for the struct by appending a number to the end.
	for i := range 1000 {
		if !g.structNames[name] {
			break
		}

		name = base + strconv.Itoa(i)
	}

	return name
}

// capitalizeFirstLetter capitalizes the first letter of a string.
func capitalizeFirstLetter(str string) string {
	if str == "" {
		return ""
	}

	return string(unicode.ToUpper(rune(str[0]))) + str[1:]
}

// stripBadChars strips underscores, dashes, dots, colons, slashes,
// and other invalid typescript interface name characters from a string.
func (g *Goty) stripBadChars(name string, typ reflect.Type) string {
	ovr := g.config.override(typ)
	if ovr.KeepBadChars && ovr.KeepUnderscores {
		return name
	}

	charsToRemove := ``

	if !ovr.KeepBadChars {
		charsToRemove += `-:./\(*&^%$#@)~"'[]{}<>,;+=|` + "`"
	}

	if !ovr.KeepUnderscores {
		charsToRemove += `_`
	}

	var output strings.Builder

	for _, r := range name {
		if !strings.ContainsRune(charsToRemove, r) {
			_, _ = output.WriteRune(r)
		}
	}

	return output.String()
}
