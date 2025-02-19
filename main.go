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

	log.Printf("Starting server on %s\n", PORT)
	log.Fatal(http.ListenAndServe(PORT, mux))
}

func run(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request")
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
