package judge

import "strings"

const (
	VerdictOK               = "ok"
	VerdictWrongAnswer      = "wrong_answer"
	VerdictRuntimeError     = "runtime_error"
	VerdictCompilationError = "compilation_error"
)

const SUFFIX = "\n"

func Match(actual, expected string) bool {
	return strings.TrimSuffix(actual, SUFFIX) == strings.TrimSuffix(expected, SUFFIX)
}
