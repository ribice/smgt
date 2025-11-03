# loopnow analyzer

Warns when `time.Now()` is called inside a loop. Hoisting the call outside avoids repeated syscalls and keeps loops cheaper.

## Run it

```bash
go install github.com/ribice/smgt/cmd/loopnow@latest
loopnow ./...
```

## Typical finding

```go
for _, job := range jobs {
	if time.Now().After(job.Deadline) { // flagged: cache time outside the loop
		job.Cancel()
	}
}
```
