# commandresult

This package wraps the functionality of calling an `*exec.Cmd.CombinedOutput()`, capturing an easy-to-serialize result
for either re-running, or writing out to disk.

## Structure

```{go}
type Result struct {
	Name     string   `json:"name"`
	Args     []string `json:"args,omitempty"`
	Env      []string `json:"env,omitempty"`
	Output   string   `json:"output,omitempty"`
	ExitCode int      `json:"exitCode"`
}
```

Most fields are `omitempty` because they disappear in certain situations. This cuts down noise on marshaling.

## Usage

```{go}
r, err := command.Capture([]string{"PATH=/bin", "echo", "hello world")
if err != nil {
    // re-capture output
    r.Capture()
    
    // compare output between runs
    reflect.DeepEqual(r, r2)
}
```

## Testing

This package is dependency free. Running `make test` is sufficient to verify its contents.

We include .golangci.yml configuration and a .drone.yaml for quality purposes.