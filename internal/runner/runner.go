package runner

import (
	"fmt"
	"log/slog"
	"os/exec"
)

// const TIMEOUT = "5s"

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Flush removes all files by filebase - request timestamp
func Flush(filebase string) error {
	command := fmt.Sprintf(`find /sandbox -type f -name "%s.*" -delete`, filebase)

	var err error
	_, err = isolate(command)
	return err
}

// Compile runs compilation process within the isolated environment
func Compile(filebase string) (*Result, error) {
	command := fmt.Sprintf(`gcc -o %s.out /sandbox/%s.c`, filebase, filebase)

	return isolate(command)
}

// ExecuteCompiled runs compiled program within the isolated environment
func ExecuteCompiled(filebase string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(`timeout %dms /sandbox/%s.out`, timeLimitMS, filebase)

	return isolate(command)
}

// Execute compiles and runs program within the isolated environment
func Execute(filebase string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; timeout %dms /sandbox/%s.out`,
		filebase, filebase, timeLimitMS, filebase,
	)

	return isolate(command)
}

// ExecuteInteractive compiles and runs program with provided input within the isolated environment
func ExecuteInteractive(filebase string, input string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; echo "%s" | timeout %dms /sandbox/%s.out`,
		filebase, filebase, input, timeLimitMS, filebase,
	)

	return isolate(command)
}

// ExecuteInteractiveCompiled runs compiled program with provided input within the isolated environment
func ExecuteInteractiveCompiled(filebase, input string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(`echo "%s" | timeout %dms /sandbox/%s.out`, input, timeLimitMS, filebase)

	return isolate(command)
}

// isolate runs provided command within isolated container
func isolate(command string) (*Result, error) {
	cmd := exec.Command("docker", "run", "--rm",
		"--cpus=0.5",
		"--memory=128m",
		"--memory-swap=256m",
		"--pids-limit=50",
		"--read-only",
		"--network=none",
		"-v", "./files:/sandbox",
		"jus1d/void-runner",
		"bash", "-c", command,
	)

	var r Result
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			r.ExitCode = ee.ExitCode()
			r.Stderr = string(ee.Stderr)
		} else {
			slog.Error("can't execute command", slog.String("error", err.Error()))
			return nil, err
		}
	}
	r.Stdout = string(out)

	return &r, nil
}
