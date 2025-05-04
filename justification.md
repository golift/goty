# Goty Justification

There are at least 7 other project on GitHub that claim to do what Goty does.
If that many projects exist, surely converting Golang structs to typescript
interfaces is easy. Read on to learn why I chose to make yet another tool.

Look, I sort of feel bad. I love the open source community. I would really like
to contribute to existing projects, and I spent a couple hours trying to fix one
of them. The code was just too complex and the comments too sparse. I gave up and
decided to write my own converter/generator.

This type of generator, when written _in_ Go requires extensive use of the `reflect`
package. I happen to have a [good deal of experience](https://github.com/golift/cnfg)
with this package. I won't say I know it well; I don't think any sane person does.
But I have struggled with it for many hours and discovered all the features and pitfalls.

But of course one of those other packages should have worked! Dude, I agree.
Let's go through what I did. First, I discovered all the projects I could find.
You'll see them below, listed in the order I tested them.

Then, I wrote test code for each one that converted [this struct](https://github.com/Notifiarr/notifiarr/blob/c809169b5df9bd72e5d13931c709f34988a506ed/pkg/configfile/config.go#L53-L75)
to typescript.
Let me digress a moment and explain this struct. I've been working on the [Notifiarr](https://notifiarr.com)
client for over 4 years now. That struct is the configuration input for the application.
It contains types from a dozen go packages with anonymous structs, slices, and maps
scattered throughout. Needless to say, it's a challenge to get all that stuff right.
Notifiarr doesn't have any embedded structs in the config, but I made sure Goty handles those as well.
My config has one enum that happens to be from the stdlib time package, so I made sure enums work too.

What constitutes working? Glad you asked. Here's my basic requirements.

- Member names should mimic what the `encoding/json` package produces (by default).
- All interfaces have unique names.
- Interface names are easy to adjust or override on a per-type basis.
- Struct types are only declared once as a typescript interface (no dups).
- All exported non-ignored (`json:"-"`) struct members have valid typescript types.
- Common types like time.Time, []byte and time.Duration are given correct types.
- The entire tree is parsed and produced. ie. It needs to be fully recursive.
- I can trigger it in my build workflow, preferably with `go generate`.

## Tests

So what happened when you tested? Here we go...

In all of the tests I put the code in a main package and used `go generate` on it.

### [tompston/gut](https://github.com/tompston/gut)

Gut produce typescript that was completely unreadable because of lack of indenting.
It completely ignored the `-` json tag, and actually used `-` as the name for many
of the interface members. That made the code unusable. Any struct that existed inside
of two or more other structs was duplicated. But they weren't duplicated with non-unique names,
they were simply embedded in the main Config struct. In other words, this only output 1 interface
consisting of 1246 lines of typescript with 0 comments or indents.

I didn't go any further with this one. The repo hasn't been touched since May 18, 2023, and the
default options produced a pile of garbage (comparatively). I don't mind contributing but this project
was designed for much smaller data sets and requires a lot of refactoring to fit my use case.

Test code:

```go
//go:generate go run .
package main

import (
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/tompston/gut"
)

func main() {
	// define which structs you want to convert, and map them to their file name.
	var interfaces map[string]any = map[string]any{
		"configfile.Config": gut.Convert(configfile.Config{}),
	}

	for name, data := range interfaces {
		err := gut.Generate(fmt.Sprintf("./%s.ts", name), fmt.Sprintln(data))
		if err != nil {
			panic(err)
		}
	}
}
```

### [tkrajina/typescriptify-golang-structs](https://github.com/tkrajina/typescriptify-golang-structs)

This is the test code I played around with.
This produced a half dozen interfaces with the name `Config`.
I found an [open issue](https://github.com/tkrajina/typescriptify-golang-structs/issues/54)
where others had this problem. I spent a couple hours trying to add a feature to override
the Name for specific types. I got it half way working, but struggled for a long time with
the location for both names (the interface and the member type that refs it).

```go
//go:generate go run .
package main

import (
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	converter := typescriptify.New().Add(configfile.Config{})
	converter.BackupDir = ""
	converter.CreateInterface = true

	err := converter.ConvertToFile("models.ts")
	if err != nil {
		panic(err)
	}
}
```

### [StirlingMarketingGroup/go2ts](https://github.com/StirlingMarketingGroup/go2ts)

This is an interesting tool. Great for those one-off tasks of needing to quickly convert a small piece of data.
The drawback is that it doesn't do anything recursively. It seems to be web-based only. I dug in the code
to figure out if it wa something I could run locally, and I was simply not able to figure it out. I put
the Notifiarr config struct into the web tool and it happily converted that single struct. There was zero
output for any of struct members.

Without recursion and without an ability to run it in a pipeline, this was a dead-end for my needs.

### [newtoallofthis123/gotypes](https://github.com/newtoallofthis123/gotypes)

At first glance the repository has 2 commits, hasn't been touched in over a year and the readme says
the author doesn't know this stuff very well. He's had a year to learn it and hasn't come back to
improve this library. That means he got right the first time, so I gave it a whirl with high expectations.

I'm kidding. I looked at the code and didn't even try it. There's
[not enough code](https://github.com/newtoallofthis123/gotypes/blob/9ca2253fd77d31b998223f728c75294ab1c76b7d/transpile.go)
to do what I need, not even close. Honestly though, this project is still small enough that I could have leveled it up.
Without an active maintainer though, that's a shot in the dark I wasn't willing to waste time on.

### [gzuidhof/tygo](https://github.com/gzuidhof/tygo)

This package was interesting. Instead of passing in a data type, you pass in a package import path.
I had high hopes, but it fell short of my expectations. Since I pass in an entire import path,
it only knows about that one path. Meaning all the data structures in the other modules were simply
omitted and replaced with the dreaded `any`. In addition to many missing data structures, I also had
a bunch of other "junk" in my typescript likes constant that are declared in Go. None of that is useful
in my context.

Test code:

```go
//go:generate go run .
package main

import "github.com/gzuidhof/tygo/tygo"

func main() {
	config := &tygo.Config{
		Packages: []*tygo.PackageConfig{
			&tygo.PackageConfig{
				Path:       "github.com/Notifiarr/notifiarr/pkg/configfile",
				OutputPath: "models.ts",
			},
		},
	}
	gen := tygo.New(config)
	err := gen.Generate()
	if err != nil {
		panic(err)
	}
}
```
### [csweichel/bel](https://github.com/csweichel/bel)

Dude this looks like a solid project! Well structured, low on the issue (1) and PR (3) count.
It appears to have a ton of features. I jumped right in. WHy is the import path in the example
different than the username on github? Oh, they changed their username in the last 6 years since
updating the repo.

Once I got it loaded I wrote a small test for the notifiarr config you see below. The output was
1 single struct. The one I passed in. So maybe it doesn't do recursion? Then I found the `bel.FollowStructs`
option and tried that. It panic'd because one of my nested structs has an interface.
```
panic: cannot get primitive Typescript type for starr.APIer (interface)
```

I looked around in the docs and examples to find a fix, but I couldn't come up with one.
[This PR](https://github.com/csweichel/bel/pull/6) looks like it might provide a workaround.
I even edit all the packages in the starr module to add `json:"-"` to all the embedded
`starr.APIer` interface members. Added a `replace` to go.mod and tried again. I got the same panic.
Double-checked I saved all the files. At this point I don't know if it's failing because it
doesn't respect the json tag or if I missed one _somewhere else_. It's not giving me enough
info in the panic. Whelp.

There are 3 open pull requests that are years old at this point. The library hasn't been touched
in 6 years, so I have low hopes of the maintainer accepting more contributions. While this
is full featured and looks well put together, I'm afraid it's progress has stalled.

I really wish this one was maintained because it felt like a good option.
I could have forked it, but it's a massive codebase and would take me a long
time to familiarize myself. It's sad because it even supports go/doc. Something I've
never ventured into and want Goty to support. I will probably come back here for
some ideas.

```go
//go:generate go run .
package main

import (
	"github.com/32leaves/bel"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
)

func main() {
	ts, err := bel.Extract(configfile.Config{}, bel.FollowStructs)
	if err != nil {
		panic(err)
	}

	err = bel.Render(ts)
	if err != nil {
		panic(err)
	}
}
```

### [OneOfOne/struct2ts](https://github.com/OneOfOne/struct2ts)

This one doesn't show an example of how to use it as a module/library,
only how to run a binary. So I looked in a test file. Seems easy enough.

With default options it gave me some hefty output containing lots of functions
and classes. Honestly the classes looked pretty good. The problem is that they're
rather hard to read, and with the dozen or so I need to verify, it's a tedious task.
I much prefer interfaces. These classes are so complicated that the code generator
added three hefty helper functions to the top of the file.

```go
package main

import (
	"os"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/OneOfOne/struct2ts"
)

func main() {
	s2ts := struct2ts.New(nil)
	s2ts.Add(configfile.Config{})
	s2ts.RenderTo(os.Stdout)
}
```

It produced three classes with the same name `Config`.
The nearly 900 lines of code has a dozen or so lines
marked in red in vscode. Some of them were just warnings,
but a couple were outright syntax errors. See the example
problem below. It has troubles parsing slices.

```ts
// `atTimes` is a `[][3]uint` in go
// And []string isn't a thing in typescript. However, string[] is.
class Endpoint {
  atTimes: [3]uint[] | null;
  query: { [key: string]: []string };
  header: { [key: string]: []string };
}
```

It doesn't support typescript interface output at all. The repo hasn't been touched in 4 years.
There's a handful of open issues with things that obviously need to work. I don't feel comfortable
attempting to extend code for another purpose when there is no sufficient need there. Passed here.


## Goty

Oh, hey you made it down here? Sweet. <3

This is the minimum code to get Goty going.
```go
package main

import (
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"golift.io/goty"
)

func main() {
	goty.NewGoty(nil).Parse(configfile.Config{}).Print()
}
```

As long as your structs don't have any embedded anonymous primitives this
will produce valid typescript. Unfortunately one exists in the Notifiarr
config struct and this invalid line of code is produced:

```ts
export interface Duration extends number {};
```

It's easily fixed though. Just add one simple override for that weird type.
It marshalls into a string, so this is simple.

```go
package main

import (
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"golift.io/cnfg"
	"golift.io/goty"
)

func main() {
	goty.NewGoty(&goty.Config{
		Overrides: goty.Overrides{
			cnfg.Duration{}: {Type: "string"},
		},
	}).Parse(configfile.Config{}).Print()
}
```

This code produces an accurate representation of the go structure.
It's about 360 lines, and over 100 of that is generated (and useful) comments.
Try it on your data today!
