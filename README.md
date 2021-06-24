# commandresult

This package wraps the functionality of calling an `*exec.Cmd.CombinedOutput()`, capturing an easy-to-serialize result for either re-running, or writing out to disk.

## usage

```go
r, err := command.Capture(context.TODO(), []string{"PATH=/bin"}, "echo", "hello world")

```

## testing

This package only depends on Go core, so `make test` is sufficient to verify its contents.

