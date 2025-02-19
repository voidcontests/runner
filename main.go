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
		Code string `json:"code"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%d.c", time.Now().Unix())
	filepath := fmt.Sprintf("./files/%s", filename)
	file, err := os.Create(filepath)
	if err != nil {
		log.Printf("Failed to create file: %v\n", err)
		http.Error(w, "Can't create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.WriteString(body.Code)
	if err != nil {
		log.Printf("Failed to write to file: %v\n", err)
		http.Error(w, "Can't write to a file", http.StatusInternalServerError)
		return
	}

	res, err := runner.Execute(filename)
	if err != nil {
		log.Printf("Failed to execute solution: %v\n", err)
		http.Error(w, "Can't execute solution", http.StatusInternalServerError)
		return
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
