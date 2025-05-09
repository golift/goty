<center><img src="https://raw.githubusercontent.com/wiki/golift/goty/goty.png"></center>

# golift.io/goty

Transform go structs into typescript interfaces.

Because no other package out there could handle my application configuration structs.

I use this to transform my go data into json and send it to my Svelte (typescript) front end.
You can see the initial [implementation into that project here](https://github.com/Notifiarr/notifiarr/pull/892).

## Others

None of these produce accurate and usable code. Most don't do either.
I tested with [this struct](https://github.com/Notifiarr/notifiarr/blob/0538806dd7753e357ee93d8eef39f640ba9dbc31/pkg/configfile/config.go#L53).

- https://github.com/StirlingMarketingGroup/go2ts
- https://github.com/tompston/gut
- https://github.com/newtoallofthis123/gotypes
- https://github.com/tkrajina/typescriptify-golang-structs
- https://github.com/gzuidhof/tygo
- https://github.com/csweichel/bel
- https://github.com/OneOfOne/struct2ts

I wrote a [justification](justification.md) explaining how I tested each of the above
projects before deciding to write another one.

## Example

```go
// Package main provides an example of how to use goty.
//
//go:generate go run .
package main

import (
	"log"
	"reflect"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"golift.io/cnfg"
	"golift.io/goty"
	"golift.io/goty/gotydoc"
)

func main() {
	weekdays := []goty.Enum{
		{Name: "Sunday", Value: time.Sunday},
		{Name: "Monday", Value: time.Monday},
		{Name: "Tuesday", Value: time.Tuesday},
		{Name: "Wednesday", Value: time.Wednesday},
		{Name: "Thursday", Value: time.Thursday},
		{Name: "Friday", Value: time.Friday},
		{Name: "Saturday", Value: time.Saturday},
	}

	docs := gotydoc.New() // Optionally, parse go/doc comments.
	goat := goty.NewGoty(&goty.Config{
		GlobalOverrides: goty.Override{
			Tag:        "json",                    // default.
			UsePkgName: goty.UsePkgNameOnConflict, // default.
			Namer: func(_ reflect.Type, name string) string {
				// Add a prefix to every interface name.
				return "Noti" + name
			},
		},
		Docs: docs,
		Overrides: goty.Overrides{
			cnfg.Duration{}: {Type: "string"},
			// Give the custom enum a JSDoc comment.
			reflect.TypeOf(time.Weekday(0)): {Comment: "The day of the week."},
		},
	})

	// Parse the weekday enums and then parse the config struct.
	goat.Enums(weekdays).Parse(configfile.Config{})
	// Make this folder by running `go mod vendor`. Delete it when you're finished.
	// This reads in all the docs and makes them available for printing/writing.
	// Do this before Printing and after parsing (so you have a list of package names).
	docs.AddMust("../vendor", goat.Pkgs()...)
	// goat.Print()
	goat.Write("notifiarrConfig.ts", true)
}
```

[This file](notifiarrConfig.ts) contains the output of the above app.
