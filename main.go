package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const PORT = ":2111"
const TIMEOUT = "5s"

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func execute(filename string) (*Result, error) {
	extension := path.Ext(filename)
	base := strings.Replace(filename, extension, "", 1)
	command := fmt.Sprintf(
		`gcc -o %s.out /sandbox/%s.c ; timeout %s /sandbox/%s.out ; EXIT_CODE=$? ; find /sandbox -type f -name "%s.*" -delete ; exit $EXIT_CODE`,
		base, base, TIMEOUT, base, base,
	)

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

func run(w http.ResponseWriter, r *http.Request) {
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

	res, err := execute(filename)
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

func main() {
	err := os.MkdirAll("files", 0755)
	if err != nil {
		log.Fatalf("Failed to create `./files/` directory: %v\n", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /run", run)

	log.Printf("Starting server on %s\n", PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))
}
