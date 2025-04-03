package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"runner/internal/judge"
	"runner/internal/runner"
	"runner/internal/runner/language"

	"github.com/gofiber/fiber/v2"
)

func Healthcheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("ok")
}

type TestRequest struct {
	Code        string `json:"code"`
	Language    string `json:"language"`
	TimeLimitMS int    `json:"time_limit_ms"`
	TCs         []struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	} `json:"tcs"`
}

type TestResponse struct {
	Verdict    string     `json:"verdict"`
	Passed     int        `json:"passed"`
	Total      int        `json:"total"`
	Stderr     string     `json:"stderr,omitempty"`
	FailedTest FailedTest `json:"failed_test"`
}

type FailedTest struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
}

func TestSolution(c *fiber.Ctx) error {
	log := slog.With(slog.String("op", "handler.TestSolution"))

	log.Info("request handled", slog.String("uri", "/test"))

	var body TestRequest
	if err := c.BodyParser(&body); err != nil {
		return Error(c, http.StatusBadRequest, "invalid request body")
	}

	l, ok := language.Get(body.Language)
	if !ok {
		return Error(c, http.StatusBadRequest, "unknown language")
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	filepath := fmt.Sprintf("./files/%s.%s", filebase, l.Extension)
	source, err := os.Create(filepath)
	if err != nil {
		log.Error("failed to create file", slog.Any("error", err))
		return InternalServerError(c)
	}
	defer source.Close()
	defer runner.Flush(filebase)

	_, err = source.WriteString(body.Code)
	if err != nil {
		log.Error("failed to write to file", slog.Any("error", err))
		return InternalServerError(c)
	}

	var report runner.Report

	if l.Kind == language.Compiled {
		report, err := runner.Compile(filebase, l.Name)
		if err != nil {
			log.Error("failed to compile solution", slog.Any("error", err))
			return InternalServerError(c)
		}

		if report.ExitCode != 0 {
			tr := TestResponse{
				Verdict: judge.VerdictCompilationError,
				Passed:  0,
				Total:   len(body.TCs),
				Stderr:  report.Stderr,
			}
			return c.Status(http.StatusOK).JSON(tr)
		}
	}

	tr := TestResponse{
		Passed: 0,
		Total:  len(body.TCs),
	}

	var ft *FailedTest

	for _, tc := range body.TCs {
		report, err = runner.Exec(filebase, l.Name, body.TimeLimitMS, tc.Input)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}

		// NOTE: 124 exit code returned by timeout, if timeout stoped the program
		if report.ExitCode == 124 {
			tr.Verdict = judge.VerdictTimeLimitExceeded
			tr.Stderr = report.Stderr

			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}

			break
		}

		match := judge.Match(report.Stdout, tc.Output)

		if report.ExitCode == 0 && match {
			tr.Passed++
		} else if report.ExitCode != 0 {
			tr.Verdict = judge.VerdictRuntimeError
			tr.Stderr = report.Stderr

			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}

			break
		}

		if ft == nil && !match {
			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}
		}
	}

	if ft != nil {
		tr.FailedTest = *ft
	}

	if tr.Verdict == "" {
		if tr.Passed == tr.Total {
			tr.Verdict = judge.VerdictOK
		} else {
			tr.Verdict = judge.VerdictWrongAnswer
		}
	}

	return c.Status(http.StatusOK).JSON(tr)
}
