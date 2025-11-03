# rot analyzer

Flags local `var` declarations that are separated from their first real use. Keeping declarations close to their usage reduces the chance of zero-value bugs and keeps scopes tight.

## Run it

```bash
go install github.com/ribice/smgt/cmd/rot@latest
rot ./...
```

## Typical finding

```go
var buf bytes.Buffer // flagged: declare this inside the branch where it is first used
if cond {
	buf.WriteString(input)
}
```
