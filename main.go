package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runner/runner"
	"time"
)

const PORT = ":2111"

func main() {
	err := os.MkdirAll("files", 0755)
	if err != nil {
		log.Fatalf("Failed to create `./files/` directory: %v\n", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /run", run)
	mux.HandleFunc("POST /test", testSubmission)

	log.Printf("Starting server on %s\n", PORT)
	log.Fatal(http.ListenAndServe(PORT, mux))
}

func run(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request to /run")
	var body struct {
		Code  string `json:"code"`
		Input string `json:"input"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	defer runner.Clean(filebase)

	source_path := fmt.Sprintf("./files/%s.c", filebase)
	source, err := os.Create(source_path)
	if err != nil {
		log.Printf("Failed to create file: %v\n", err)
		http.Error(w, "Can't create file", http.StatusInternalServerError)
		return
	}
	defer source.Close()

	_, err = source.WriteString(body.Code)
	if err != nil {
		log.Printf("Failed to write to file: %v\n", err)
		http.Error(w, "Can't write to a file", http.StatusInternalServerError)
		return
	}

	var res *runner.Result
	if body.Input != "" {
		input_path := fmt.Sprintf("./files/%s.input.txt", filebase)
		inputtxt, err := os.Create(input_path)
		if err != nil {
			log.Printf("Failed to create file: %v\n", err)
			http.Error(w, "Can't create file", http.StatusInternalServerError)
			return
		}
		defer inputtxt.Close()

		_, err = inputtxt.WriteString(body.Input)
		if err != nil {
			log.Printf("Failed to write to file: %v\n", err)
			http.Error(w, "Can't write to a file", http.StatusInternalServerError)
			return
		}

		res, err = runner.ExecuteInteractive(filebase)
		if err != nil {
			log.Printf("Failed to execute solution: %v\n", err)
			http.Error(w, "Can't execute solution", http.StatusInternalServerError)
			return
		}
	} else {
		res, err = runner.Execute(filebase)
		if err != nil {
			log.Printf("Failed to execute solution: %v\n", err)
			http.Error(w, "Can't execute solution", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]any{
		"status": res.ExitCode,
		"stdout": res.Stdout,
		"stderr": res.Stderr,
	}

	json, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshall json response: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

type TestResponse struct {
	Verdict string `json:"verdict"` // "ok" | "wrong_answer" | "compilation_error" | "runtime_error"
	Passed  int    `json:"passed"`
	Total   int    `json:"total"`
}

func testSubmission(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request to /test")
	var body struct {
		Code string `json:"code"`
		TCs  []struct {
			Input  string `json:"input"`
			Output string `json:"output"`
		} `json:"tcs"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filebase := fmt.Sprintf("%d", time.Now().Unix())
	defer runner.Clean(filebase)

	source_path := fmt.Sprintf("./files/%s.c", filebase)
	source, err := os.Create(source_path)
	if err != nil {
		log.Printf("Failed to create file: %v\n", err)
		http.Error(w, "Can't create file", http.StatusInternalServerError)
		return
	}
	defer source.Close()

	_, err = source.WriteString(body.Code)
	if err != nil {
		log.Printf("Failed to write to file: %v\n", err)
		http.Error(w, "Can't write to a file", http.StatusInternalServerError)
		return
	}

	ok, err := runner.Compile(filebase)
	if err != nil {
		log.Printf("Failed to compile: %v\n", err)
		http.Error(w, "can't compile", http.StatusInternalServerError)
		return
	}

	if !ok {
		response := TestResponse{
			Verdict: "compilation_error",
			Passed:  0,
			Total:   len(body.TCs),
		}

		json, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshall json response: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	var res *runner.Result
	tr := TestResponse{
		Passed: 0,
		Total:  len(body.TCs),
	}
	for _, tc := range body.TCs {
		input_path := fmt.Sprintf("./files/%s.input.txt", filebase)
		inputtxt, err := os.Create(input_path)
		if err != nil {
			log.Printf("Failed to create file: %v\n", err)
			http.Error(w, "Can't create file", http.StatusInternalServerError)
			return
		}
		defer inputtxt.Close()

		_, err = inputtxt.WriteString(tc.Input)
		if err != nil {
			log.Printf("Failed to write to file: %v\n", err)
			http.Error(w, "Can't write to a file", http.StatusInternalServerError)
			return
		}

		res, err = runner.ExecuteInteractive(filebase)
		if err != nil {
			log.Printf("Failed to execute solution: %v\n", err)
			http.Error(w, "Can't execute solution", http.StatusInternalServerError)
			return
		}

		if res.ExitCode == 0 && res.Stdout == tc.Output {
			tr.Passed++
		} else if res.ExitCode != 0 {
			tr.Verdict = "runtime_error"
		}
	}

	if tr.Verdict != "runtime_error" {
		if tr.Passed == tr.Total {
			tr.Verdict = "ok"
		} else {
			tr.Verdict = "wrong_answer"
		}
	}

	json, err := json.Marshal(tr)
	if err != nil {
		log.Printf("Failed to marshall json response: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
