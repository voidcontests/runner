package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"runner/internal/judge"
	"runner/internal/runner"
	"runner/internal/runner/language"
)

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling /healthcheck")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
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

func TestSolution(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling /test")

	var body TestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		InvalidRequestBody(w)
		return
	}

	l, ok := language.Get(body.Language)
	if !ok {
		WriteError(w, http.StatusBadRequest, "unknown language")
		return
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	filepath := fmt.Sprintf("./files/%s.%s", filebase, l.Extension)
	source, err := os.Create(filepath)
	if err != nil {
		log.Printf("failed to create file: %v", err)
		InternalServerError(w)
		return
	}
	defer source.Close()
	defer runner.Flush(filebase)

	if _, err := source.WriteString(body.Code); err != nil {
		log.Printf("failed to write to file: %v", err)
		InternalServerError(w)
		return
	}

	var report runner.Report
	if l.Kind == language.Compiled {
		report, err = runner.Compile(filebase, l.Name)
		if err != nil {
			log.Printf("failed to compile solution: %v", err)
			InternalServerError(w)
			return
		}

		if report.ExitCode != 0 {
			tr := TestResponse{
				Verdict: judge.VerdictCompilationError,
				Passed:  0,
				Total:   len(body.TCs),
				Stderr:  report.Stderr,
			}
			WriteJSON(w, http.StatusOK, tr)
			return
		}
	}

	tr := TestResponse{
		Passed: 0,
		Total:  len(body.TCs),
	}

	var ft FailedTest
	failed := false

	for _, tc := range body.TCs {
		report, err = runner.Exec(filebase, l.Name, body.TimeLimitMS, tc.Input)
		if err != nil {
			log.Printf("failed to execute solution: %v", err)
			InternalServerError(w)
			return
		}

		// if exit code 124, assume timeout
		if report.ExitCode == 124 {
			tr.Verdict = judge.VerdictTimeLimitExceeded
			tr.Stderr = report.Stderr
			ft = FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}
			failed = true
			break
		}

		match := judge.Match(report.Stdout, tc.Output)
		if report.ExitCode == 0 && match {
			tr.Passed++
		} else if report.ExitCode != 0 {
			tr.Verdict = judge.VerdictRuntimeError
			tr.Stderr = report.Stderr
			ft = FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}
			failed = true
			break
		} else if failed == false && !match {
			ft = FailedTest{
				Input:          tc.Input,
				ExpectedOutput: tc.Output,
				ActualOutput:   report.Stdout,
			}
			failed = true
		}
	}

	if failed {
		tr.FailedTest = ft
	}
	if tr.Verdict == "" {
		if tr.Passed == tr.Total {
			tr.Verdict = judge.VerdictOK
		} else {
			tr.Verdict = judge.VerdictWrongAnswer
		}
	}

	WriteJSON(w, http.StatusOK, tr)
}
