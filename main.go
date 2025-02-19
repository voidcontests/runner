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

type RequestBody struct {
	Code string `json:"code"`
}

func main() {
	http.HandleFunc("/run", run)
	log.Println("server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func run(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	log.Println("got request")

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%d.c", time.Now().Unix())
	filepath := fmt.Sprintf("./files/%s", filename)
	file, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Can't create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.WriteString(body.Code)
	if err != nil {
		http.Error(w, "Can't write to a file", http.StatusInternalServerError)
		return
	}

	res, err := runner.Execute(filename)
	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
