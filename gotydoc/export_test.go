package gotydoc

// Test helpers. Compiled only during `go test`.

// TypeNames is a test helper that lists go/doc type names for pkg.
func (d *Docs) TypeNames(pkg string) []string {
	indexed := d.pkgs[pkg]
	if indexed == nil {
		return nil
	}

	names := make([]string, 0, len(indexed.Types))
	for _, typ := range indexed.Types {
		names = append(names, typ.Name)
	}

	return names
}
