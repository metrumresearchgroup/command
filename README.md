# command

This package wraps the functionality of calling an `*exec.Cmd.CombinedOutput()`, capturing an easy-to-serialize result
for either re-running, or writing out to disk.

## Goal

The project's goal is ease of use when configuring, starting/stopping, and directly capturing output of a `exec.Cmd` call.

Its secondary goal is to allow use to re-run a command with its complete environment repeatedly.

## Use Cases

1. Saving a pre-defined command, and executing it repeatedly.
2. Checking applications for idempotence, with all environmental variability controlled by the package.
3. Serializing the configuration for future use as JSON and reloading it at startup.

## Structure

```{go}
type Capture struct {
	Name string `json:"name"`
	Args []string `json:"args,omitempty"`
	Dir string `json:"dir,omitempty"`
	Env []string `json:"env,omitempty"`
	Output []byte `json:"output,omitempty"`
	ExitCode int `json:"exit_code"`
}
```

Most fields are `omitempty` because they disappear in certain situations. This cuts down noise on marshaling.

## Usage

```{go}
cmd := command.New(command.WithEnv([]string{"PATH=/bin"}))
err := cmd.Run("echo", "hello world")
if err != nil {
    // re-capture output
    err = r.Rerun()
}
```

## Testing

This package is dependency free. Running `make test` is sufficient to verify its contents.

We include .golangci.yml configuration and a .drone.yaml for quality purposes.