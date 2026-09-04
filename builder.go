package goty

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golift.io/goty/gotyface"
)

const (
	tsString  = "string"
	tsDate    = "Date"
	tsAny     = "any"
	tsNumber  = "number"
	tsBoolean = "boolean"
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
	// Extends is a list of struct names that this member extends if it's a struct with anonymous members.
	Extends []string
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
	if len(enum) == 0 {
		return
	}

	typ := reflect.TypeOf(enum[0].Value)
	if typ == nil {
		return
	}

	data := &DataStruct{
		Elements: make([]*Enum, len(enum)),
		doc:      g.config,
		Type:     typ,
		Name:     g.getStructName(typ),
		GoName:   typ.PkgPath() + "." + instantiatedName(typ.Name()),
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
		GoName:  elem.PkgPath() + "." + instantiatedName(elem.Name()),
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
		name, opts, hasComma := splitTag(elem.Tag.Get(ovr.Tag))

		switch {
		case name == "-" && !hasComma:
			continue
		case !elem.IsExported() && !isAnonymousStructField(elem):
			// encoding/json promotes exported fields from unexported anonymous
			// structs, but ignores unexported anonymous scalars and other kinds.
			continue
		case name == "":
			name = elem.Name
		}

		tsName := g.stripBadChars(name, elem.Type)
		if name == "-" {
			// json:"-," is the escape for a field literally named "-".
			tsName = `"-"`
		}

		member := &StructMember{
			Name:     tsName,
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

		applyJSONOptions(member, opts, ovr.Type)
		data.addMember(member)
	}
}

func splitTag(tag string) (string, []string, bool) {
	name, rest, hasComma := strings.Cut(tag, ",")
	if !hasComma {
		return name, nil, false
	}

	if rest == "" {
		return name, nil, true
	}

	return name, strings.Split(rest, ","), true
}

func applyJSONOptions(member *StructMember, options []string, typeOverride string) {
	for _, opt := range options {
		switch opt {
		case "omitempty", "omitzero":
			member.Optional = true
		case tsString:
			if typeOverride == "" && jsonStringKind(member.Member.Type) {
				member.Type = tsString
			}
		}
	}
}

// jsonStringKind reports whether encoding/json honors the `string` tag option.
func jsonStringKind(typ reflect.Type) bool {
	if implementsIface(typ, reflect.TypeFor[json.Marshaler]()) {
		return false
	}

	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	switch typ.Kind() { //nolint:exhaustive // encoding/json only honors string on bool, string, int, and float.
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// addMember adds a member to the struct.
// If the member is an anonymous struct, it extends the parent struct.
// Otherwise, it adds the member to the struct.
func (d *DataStruct) addMember(member *StructMember) {
	if isAnonymousEmbed(member) {
		d.Extends = append(d.Extends, member.Type)
	} else {
		d.Members = append(d.Members, member)
	}
}

func isAnonymousEmbed(member *StructMember) bool {
	if !isAnonymousStructField(member.Member) {
		return false
	}

	jsonName, _, _ := strings.Cut(member.Member.Tag.Get(member.ovr.Tag), ",")
	if jsonName != "" {
		return false
	}

	// Marshaler/time.Time embeds become scalars; a type merely named Date still extends.
	_, _, ok := specialType(member.Member.Type)

	return !ok
}

func isAnonymousStructField(elem reflect.StructField) bool {
	if !elem.Anonymous {
		return false
	}

	typ := elem.Type

	return typ.Kind() == reflect.Struct ||
		typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct
}

// parseMember returns the typescript type for a given go type.
// It also returns a boolean indicating if the type is optional.
// Fully recursive.
//
//nolint:cyclop // This is a complex function, but really it's not that bad.
func (g *Goty) parseMember(parent *DataStruct, field reflect.Type, member *StructMember) (string, bool) {
	if data := g.structTypes[field]; data != nil && len(data.Elements) > 0 {
		return data.Name, false
	}

	if name, optional, ok := specialType(field); ok {
		return name, optional
	}

	if data := g.structTypes[field]; data != nil {
		return data.Name, false
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
		return tsBoolean, false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Uintptr:
		return tsNumber, false
	case reflect.String:
		return tsString, false
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
		return tsAny, true
	}
}

var (
	errNotJSONMarshaler = errors.New("not a json.Marshaler")
	errJSONMarshalPanic = errors.New("json marshaler panicked")
)

// specialType maps types whose JSON form is not their Go kind.
// time.Time stays Date. json.Marshaler is probed on the zero value.
// encoding.TextMarshaler (without MarshalJSON) is a JSON string.
func specialType(field reflect.Type) (string, bool, bool) {
	optional := field.Kind() == reflect.Ptr || field.Kind() == reflect.Slice || field.Kind() == reflect.Map

	for field.Kind() == reflect.Ptr {
		field = field.Elem()
		optional = true
	}

	switch {
	case field == reflect.TypeFor[time.Time]():
		return tsDate, optional, true
	case field == reflect.TypeFor[json.RawMessage]():
		return tsAny, true, true
	default:
		return specialMarshalerType(field, optional)
	}
}

func specialMarshalerType(field reflect.Type, optional bool) (string, bool, bool) {
	if implementsIface(field, reflect.TypeFor[json.Marshaler]()) {
		if name, ok := inferJSONMarshalerType(field); ok {
			return name, optional, true
		}
	}

	if implementsIface(field, reflect.TypeFor[encoding.TextMarshaler]()) {
		return tsString, optional, true
	}

	return "", false, false
}

func implementsIface(typ, iface reflect.Type) bool {
	if typ.Implements(iface) {
		return true
	}

	return typ.Kind() != reflect.Ptr && reflect.PointerTo(typ).Implements(iface)
}

func inferJSONMarshalerType(typ reflect.Type) (string, bool) {
	raw, err := marshalJSONZero(typ)
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return tsAny, true
	}

	switch bytes.TrimSpace(raw)[0] {
	case '"':
		return tsString, true
	case 't', 'f':
		return tsBoolean, true
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return tsNumber, true
	case 'n':
		// Nil-receiver marshalers often return null; keep the Go element type.
		return "", false
	default:
		return tsAny, true
	}
}

func marshalJSONZero(typ reflect.Type) ([]byte, error) {
	var (
		raw []byte
		err error
	)

	func() {
		defer func() {
			if recover() != nil {
				err = errJSONMarshalPanic
			}
		}()

		raw, err = callJSONMarshaler(typ)
	}()

	return raw, err
}

func callJSONMarshaler(typ reflect.Type) ([]byte, error) {
	iface := reflect.TypeFor[json.Marshaler]()

	var val reflect.Value
	switch {
	case typ.Kind() == reflect.Ptr:
		val = reflect.New(typ.Elem())
	case reflect.PointerTo(typ).Implements(iface):
		val = reflect.New(typ)
	default:
		val = reflect.New(typ).Elem()
	}

	marshaler, ok := val.Interface().(json.Marshaler)
	if !ok {
		return nil, errNotJSONMarshaler
	}

	raw, err := marshaler.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	return raw, nil
}

// checkStruct provides some logic to detect special struct types.
// Those include time.Time/Duration and embedded structs. Do we need others?
func (g *Goty) checkStruct(field reflect.Type, member *StructMember) string {
	switch field.String() {
	case "time.Time":
		return tsDate
	case "time.Duration":
		return tsNumber
	}

	structMember := g.parseStruct(field)
	if structMember.Name == "" { // Embedded struct.
		member.Members = append(member.Members, structMember.Members...)
		member.Extends = append(member.Extends, structMember.Extends...)

		return "" // embedded structs don't have names; deal with it.
	}

	return structMember.Name
}

// parseSlice returns the typescript type for a given go slice.
func (g *Goty) parseSlice(parent *DataStruct, field reflect.Type, member *StructMember) string {
	// Go marshalls a byte slice into a base64 encoded string.
	// json.RawMessage is a named []byte inserted as raw JSON. Since Go 1.26 it is
	// an alias of encoding/json/jsontext.Value, so compare types, not PkgPath/Name.
	if field == reflect.TypeFor[json.RawMessage]() {
		return tsAny
	}

	if field.Kind() == reflect.Slice && field.Elem().Kind() == reflect.Uint8 {
		return tsString
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
	elemName := instantiatedName(elem.Name())
	name := ovr.Namer(elem, capitalizeFirstLetter(elemName))
	name = g.stripBadChars(name, elem)
	pkgParts := strings.Split(elem.PkgPath(), "/")

	if ovr.UsePkgName == UsePkgNameAlways ||
		(g.structNames[name] && ovr.UsePkgName == UsePkgNameOnConflict) {
		// We have to pass the original element name back in here so any name changes are repeated.
		name = ovr.Namer(elem, capitalizeFirstLetter(pkgParts[len(pkgParts)-1])+elemName)
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

// instantiatedName strips type parameters from a reflect type name.
// APIResponse[any] is named "APIResponse[interface {}]" at runtime; those
// brackets would otherwise be stripped into "APIResponseinterface".
func instantiatedName(name string) string {
	base, _, _ := strings.Cut(name, "[")

	return base
}
