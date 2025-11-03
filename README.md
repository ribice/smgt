# s(small) g(o) t(ools)

This repository hosts a handful of single-purpose Go static analysis helpers. Each lives in its own package with a matching CLI under `cmd/`, so you can run only what you need or wire several together.

## Tools

- [rot](rot/README.md): Flags local variable declarations that are separated from their first real use, keeping scopes tight and avoiding stale zero values.
- [set](set/README.md): Spots `map[string]bool` values that only store `true`, suggesting `map[string]struct{}` to save heap.
- [loopnow](loopnow/README.md): Warns when `time.Now()` is called inside loops and should be hoisted out.

Each README dives into typical findings, usage, and install commands.

## Installation

```bash
go install github.com/ribice/smgt/cmd/rot@latest
go install github.com/ribice/smgt/cmd/set@latest
go install github.com/ribice/smgt/cmd/loopnow@latest
```

## Running analyzers

Invoke the binary for the analyzer you want against a package pattern:

```bash
rot ./...
set ./...
loopnow ./...
```

Want to run several at once? Compose them with [`multichecker`](https://pkg.go.dev/golang.org/x/tools/go/analysis/multichecker):

```go
package main

import (
	"github.com/ribice/smgt/loopnow"
	"github.com/ribice/smgt/rot"
	"github.com/ribice/smgt/set"
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

## Development

Each analyzer ships with [`golang.org/x/tools/go/analysis/analysistest`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest) fixtures under `<analyzer>/testdata`. Every `// want` comment records the expected diagnostic. Run the full suite with:

```bash
go test ./...
```

Contributions are welcome—add more fixtures or new analyzers that keep runtime surprises out of production.
