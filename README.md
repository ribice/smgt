# rot analyzers

Right On Time is a collection of lightweight Go static analyzers aimed at catching subtle runtime costs and logic bugs before they ship. Each analyzer lives in its own package and binary so you can run only what you need—or wire them into your own multichecker.

Available analyzers (package → CLI):

- `github.com/ribice/smgc/rot` → `github.com/ribice/smgc/cmd/rot`
- `github.com/ribice/smgc/set` → `github.com/ribice/smgc/cmd/set`
- `github.com/ribice/smgc/loopnow` → `github.com/ribice/smgc/cmd/loopnow`

## Installation

```bash
go install github.com/ribice/smgc/cmd/rot@latest
go install github.com/ribice/smgc/cmd/set@latest
go install github.com/ribice/smgc/cmd/loopnow@latest
```

Running any of the binaries against a package pattern executes the corresponding analyzer:

```bash
rot ./...
set ./...
loopnow ./...
```

Want all analyzers at once? Compose them with [`multichecker`](https://pkg.go.dev/golang.org/x/tools/go/analysis/multichecker):

```go
package main

import (
	"github.com/ribice/smgc/loopnow"
	"github.com/ribice/smgc/rot"
	"github.com/ribice/smgc/set"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		rot.NewAnalyzer(),
		set.NewAnalyzer(),
		loopnow.NewAnalyzer(),
	)
}
```

## How it Works

`rot` tracks block-scoped variable declarations—plain `var` statements, short declarations (`:=`), range assignments, and type-switch bindings—and checks when each identifier is first referenced. If any other statements intervene before that first use, `rot` reports a diagnostic similar to:

```
simple.go:6:6: variable name should be declared right before it is used
```

## Example

```go
var name string // rot: variable name should be declared right before it is used
if age < 18 {
	return "minor"
}
name = strconv.Itoa(age)
return name
```

Move the declaration down so that the first use happens immediately:

```go
if age < 18 {
	return "minor"
}
name := strconv.Itoa(age)
return name
```

## Development

Each analyzer ships with [`golang.org/x/tools/go/analysis/analysistest`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest) fixtures under `<analyzer>/testdata`. Every `// want` comment records the expected diagnostic. Run the full suite with:

```bash
go test ./...
```

Contributions are welcome—add more fixtures or new analyzers that help developers avoid latency surprises.
