package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runner/internal/judge"
	"runner/internal/runner"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Healthcheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("ok")
}

type TestRequest struct {
	Code        string `json:"code"`
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

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	filepath := fmt.Sprintf("./files/%s.c", filebase)
	source, err := os.Create(filepath)
	if err != nil {
		log.Error("failed to create file", slog.Any("error", err))
		return InternalServerError(c)
	}
	defer source.Close()

	_, err = source.WriteString(body.Code)
	if err != nil {
		log.Error("failed to write to file", slog.Any("error", err))
		return InternalServerError(c)
	}

	res, err := runner.Compile(filebase)
	defer runner.Flush(filebase)
	if err != nil {
		log.Error("failed to compile solution", slog.Any("error", err))
		return InternalServerError(c)
	}

	if res.ExitCode != 0 {
		tr := TestResponse{
			Verdict: judge.VerdictCompilationError,
			Passed:  0,
			Total:   len(body.TCs),
			Stderr:  res.Stderr,
		}
		return c.Status(http.StatusOK).JSON(tr)
	}

	tr := TestResponse{
		Passed: 0,
		Total:  len(body.TCs),
	}

	var ft *FailedTest

	for _, tc := range body.TCs {
		res, err = runner.ExecuteInteractiveCompiled(filebase, tc.Input, body.TimeLimitMS)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}

		// NOTE: 124 exit code returned by timeout = timeout stoped the program
		if res.ExitCode == 124 {
			tr.Verdict = judge.VerdictTimeLimitExceeded
			tr.Stderr = res.Stderr

			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   res.Stdout,
			}

			break
		}

		match := judge.Match(res.Stdout, tc.Output)

		if res.ExitCode == 0 && match {
			tr.Passed++
		} else if res.ExitCode != 0 {
			tr.Verdict = judge.VerdictRuntimeError
			tr.Stderr = res.Stderr

			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   res.Stdout,
			}

			break
		}

		if ft == nil && !match {
			ft = &FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   res.Stdout,
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

func RunSolution(c *fiber.Ctx) error {
	log := slog.With(slog.String("op", "handler.RunSolution"))

	log.Info("request handled", slog.String("uri", "/run"))

	var body struct {
		Code        string `json:"code"`
		Input       string `json:"input"`
		TimeLimitMS int    `json:"time_limit_ms"`
	}

	if err := c.BodyParser(&body); err != nil {
		return Error(c, http.StatusBadRequest, "invalid request body")
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	defer runner.Flush(filebase)

	sourcePath := fmt.Sprintf("./files/%s.py", filebase)
	source, err := os.Create(sourcePath)
	if err != nil {
		log.Error("failed to create file", slog.Any("error", err))
		return InternalServerError(c)
	}
	defer source.Close()

	_, err = source.WriteString(body.Code)
	if err != nil {
		log.Error("failed to write to file", slog.Any("error", err))
		return InternalServerError(c)
	}

	var res *runner.Result
	if body.Input != "" {
		res, err = runner.ExecuteInteractive(filebase, body.Input, body.TimeLimitMS)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}
	} else {
		res, err = runner.RunPython(filebase, body.TimeLimitMS)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}
	}

	response := fiber.Map{
		"status": res.ExitCode,
		"stdout": res.Stdout,
		"stderr": res.Stderr,
	}

	return c.JSON(response)
}
