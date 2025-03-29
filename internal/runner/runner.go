package runner

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
)

const (
	LangC      = "c"
	LangPython = "python"
)

var ErrUnknownLanguage = errors.New("runner: unknown language")

func getCompilationCommand(language, filebase string) string {
	switch language {
	case LangC:
		return fmt.Sprintf(`gcc -o %s.out /sandbox/%s.c`, filebase, filebase)
	}

	return ""
}

func getExecuteCommand(language, filebase, input string, timeLimitMS int) string {
	switch language {
	case LangC:
		return fmt.Sprintf(`echo "%s" | timeout %0.3fs /sandbox/%s.out`, input, float64(timeLimitMS)/1000, filebase)
	case LangPython:
		return fmt.Sprintf(`echo "%s" | timeout %0.3fs python3 /sandbox/%s.py`, input, float64(timeLimitMS)/1000, filebase)
	}

	return ""
}

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func is_compiled(language string) bool {
	switch language {
	case LangC:
		return true
	default:
		return false
	}
}

func Exec(filebase, language string, timeLimitMS int, input string) (*Result, error) {
	var r *Result
	var err error
	if is_compiled(language) {
		command := getCompilationCommand(language, filebase)

		r, err = isolate(command)
		if err != nil {
			return nil, err
		}
		if r.ExitCode != 0 {
			return r, nil
		}
	}

	command := getExecuteCommand(language, filebase, input, timeLimitMS)
	if command == "" {
		return nil, ErrUnknownLanguage
	}

	return isolate(command)
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
	command := fmt.Sprintf(`timeout %0.3fs /sandbox/%s.out`, float64(timeLimitMS)/1000, filebase)

	return isolate(command)
}

// Execute compiles and runs program within the isolated environment
func Execute(filebase string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; timeout %0.3fs /sandbox/%s.out`,
		filebase, filebase, float64(timeLimitMS)/1000, filebase,
	)

	return isolate(command)
}

// ExecuteInteractive compiles and runs program with provided input within the isolated environment
func ExecuteInteractive(filebase string, input string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; echo "%s" | timeout %0.3fs /sandbox/%s.out`,
		filebase, filebase, input, float64(timeLimitMS)/1000, filebase,
	)

	return isolate(command)
}

func RunPython(filebase string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(
		`python3 %s.py`,
		filebase,
	)

	return isolate(command)
}

// ExecuteInteractiveCompiled runs compiled program with provided input within the isolated environment
func ExecuteInteractiveCompiled(filebase, input string, timeLimitMS int) (*Result, error) {
	command := fmt.Sprintf(`echo "%s" | timeout %0.3fs /sandbox/%s.out`, input, float64(timeLimitMS)/1000, filebase)

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
		"runner:latest",
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
