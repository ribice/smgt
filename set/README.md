# set analyzer

Detects `map[string]bool` values that are only ever written the constant `true`, signalling they are being used as sets. Recommends switching to `map[string]struct{}` to avoid wasted heap.

## Run it

```bash
go install github.com/ribice/smgc/cmd/set@latest
set ./...
```

## Typical finding

```go
hosts := map[string]bool{} // flagged: prefer map[string]struct{}
hosts[h.Name] = true
```
