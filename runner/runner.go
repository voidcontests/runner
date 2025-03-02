package runner

import (
	"fmt"
	"log/slog"
	"os/exec"
)

const TIMEOUT = "5s"

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func Clean(filebase string) error {
	command := fmt.Sprintf(`find /sandbox -type f -name "%s.*" -delete`, filebase)

	var err error
	_, err = run_isolated(command)
	return err
}

func Compile(filebase string) (bool, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c`,
		filebase, filebase,
	)

	res, err := run_isolated(command)
	if err != nil {
		return false, err
	}
	return res.ExitCode == 0, nil
}

func Execute(filebase string) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; timeout %s /sandbox/%s.out`,
		filebase, filebase, TIMEOUT, filebase,
	)

	return run_isolated(command)
}

func ExecuteInteractive(filebase string) (*Result, error) {
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; cat /sandbox/%s.input.txt | timeout %s /sandbox/%s.out`,
		filebase, filebase, filebase, TIMEOUT, filebase,
	)

	return run_isolated(command)
}

func run_isolated(command string) (*Result, error) {
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
