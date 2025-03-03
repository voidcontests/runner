package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runner/internal/runner"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Healthcheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("ok")
}

type TestRequest struct {
	Code string `json:"code"`
	TCs  []struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	} `json:"tcs"`
}

type TestResponse struct {
	Verdict    string     `json:"verdict"`
	Passed     int        `json:"passed"`
	Total      int        `json:"total"`
	FailedTest FailedTest `json:"failed_test,omitempty"`
}

type FailedTest struct {
	Input          string `json:"input,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	ActualOutput   string `json:"actual_output,omitempty"`
}

func TestSolution(c *fiber.Ctx) error {
	log := slog.With(slog.String("op", "handler.TestSolution"))

	log.Info("request handled", slog.String("uri", "/test"))

	var body TestRequest
	if err := c.BodyParser(&body); err != nil {
		return Error(c, http.StatusBadRequest, "invalid request body")
	}

	fmt.Println(body)

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	sourcePath := fmt.Sprintf("./files/%s.c", filebase)
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

	ok, err := runner.Compile(filebase)
	defer runner.Flush(filebase)
	if err != nil {
		log.Error("failed to compile solution", slog.Any("error", err))
		return InternalServerError(c)
	}

	if !ok {
		tr := TestResponse{
			Verdict: "compilation_error",
			Passed:  0,
			Total:   len(body.TCs),
		}
		return c.Status(http.StatusOK).JSON(tr)
	}

	tr := TestResponse{
		Passed: 0,
		Total:  len(body.TCs),
	}

	var ft *FailedTest

	for _, tc := range body.TCs {
		finput, err := os.Create(fmt.Sprintf("./files/%s.input.txt", filebase))
		if err != nil {
			log.Error("failed to create file", slog.Any("error", err))
			return InternalServerError(c)
		}

		_, err = finput.WriteString(tc.Input)
		finput.Close()
		if err != nil {
			log.Error("failed to write to file", slog.Any("error", err))
			return InternalServerError(c)
		}

		res, err := runner.ExecuteInteractiveCompiled(filebase)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}

		if res.ExitCode == 0 && res.Stdout == tc.Output {
			tr.Passed++
		} else if res.ExitCode != 0 {
			tr.Verdict = "runtime_error"
			break
		}

		if ft == nil && (res.ExitCode != 0 || res.Stdout != tc.Output) {
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

	if tr.Verdict != "runtime_error" {
		if tr.Passed == tr.Total {
			tr.Verdict = "ok"
		} else {
			tr.Verdict = "wrong_answer"
		}
	}

	return c.Status(http.StatusOK).JSON(tr)
}

func RunSolution(c *fiber.Ctx) error {
	log := slog.With(slog.String("op", "handler.TestSolution"))

	log.Info("request handled", slog.String("uri", "/run"))

	var body struct {
		Code  string `json:"code"`
		Input string `json:"input"`
	}

	if err := c.BodyParser(&body); err != nil {
		return Error(c, http.StatusBadRequest, "invalid request body")
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	defer runner.Flush(filebase)

	sourcePath := fmt.Sprintf("./files/%s.c", filebase)
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
		inputPath := fmt.Sprintf("./files/%s.input.txt", filebase)
		inputFile, err := os.Create(inputPath)
		if err != nil {
			log.Error("failed to create file", slog.Any("error", err))
			return InternalServerError(c)
		}
		defer inputFile.Close()

		_, err = inputFile.WriteString(body.Input)
		if err != nil {
			log.Error("failed to write to file", slog.Any("error", err))
			return InternalServerError(c)
		}

		res, err = runner.ExecuteInteractive(filebase)
		if err != nil {
			log.Error("failed to execute solution", slog.Any("error", err))
			return InternalServerError(c)
		}
	} else {
		res, err = runner.Execute(filebase)
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
