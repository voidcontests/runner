package runner

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strings"
)

const TIMEOUT = "5s"

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func Execute(filename string) (*Result, error) {
	extension := path.Ext(filename)
	base := strings.Replace(filename, extension, "", 1)
	cmd := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; timeout %s /sandbox/%s.out ; EXIT_CODE=$? ; find /sandbox -type f -name "%s.*" -delete ; exit $EXIT_CODE`,
		base, base, TIMEOUT, base, base,
	)

	return isolate(cmd)
}

func isolate(command string) (*Result, error) {
	cmd := exec.Command("docker", "run", "--rm",
		"--cpus=0.5",
		"--memory=128m",
		"--memory-swap=256m",
		"--pids-limit=50",
		"--read-only",
		"--network=none",
		"-v", "./files:/sandbox",
		"runner",
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
