# command

This package wraps the functionality of calling an `*exec.Cmd.CombinedOutput()`, capturing an easy-to-serialize result
for either re-running, or writing out to disk.

## Structure

```{go}
type Capture struct {
	Name string `json:"name"`
	Args []string `json:"args,omitempty"`
	Dir string `json:"working_dir,omitempty"`
	Output string `json:"output,omitempty"`
	ExitCode int `json:"exit_code"`
}
```

Most fields are `omitempty` because they disappear in certain situations. This cuts down noise on marshaling.

## Usage

```{go}
r, err := command.New(command.WithEnv([]string{"PATH=/bin"})).Run("echo", "hello world")
if err != nil {
    // re-capture output
    r2, _ := r.Rerun()
    
    // compare output between runs
    if !reflect.DeepEqual(r, r2) {
        // etc.
    }
}
```

## Testing

This package is dependency free. Running `make test` is sufficient to verify its contents.

We include .golangci.yml configuration and a .drone.yaml for quality purposes.