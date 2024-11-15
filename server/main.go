package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// ToDo:
// - Add a healthcheck endpoint
// - Add a metrics endpoint (tracing, metrics, etc.)
// - Add tests
// - Add a basic auth layer
// - Add a rate limiter
// - Add a basic UI

// Logging
func loggingMiddleware(next http.Handler) http.Handler {
	logFile, err := os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger := log.New(logFile, "", log.LstdFlags)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request
		logger.Printf(
			"IP: %s | Method: %s | Path: %s | Duration: %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

func main() {
	// Create the logs directory if it doesn't exist
	os.MkdirAll("server/logs", 0755)

	// Create the files directory if it doesn't exist
	os.MkdirAll("files", 0755)

	// Setup routes
	http.Handle("/files/", loggingMiddleware(http.HandlerFunc(handleFileRequests)))

	log.Println("Server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format != "json" && format != "" {
		format = "text"
	}

	files, err := listTextFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"files": files})
	} else {
		w.Header().Set("Content-Type", "text/plain")
		for _, file := range files {
			fmt.Fprintln(w, file)
		}
	}
}

func handleFileRequests(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format != "json" && format != "" {
		format = "text"
	}

	// Strip "/files/" from the path to get filename
	filename := strings.TrimPrefix(r.URL.Path, "/files/")

	if filename == "" || filename == "files" {
		// List files
		listFiles(w, format)
		return
	}

	// Get the specific file
	getFile(w, filename, format)
}

func listFiles(w http.ResponseWriter, format string) {
	files, err := listTextFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"files": files})
	} else {
		w.Header().Set("Content-Type", "text/plain")
		// Just print the filenames, one per line
		for _, file := range files {
			fmt.Fprintln(w, file)
		}
	}
}

func getFile(w http.ResponseWriter, filename string, format string) {
	// Ensure the filename is sanitized and within the files directory
	filename = path.Base(filename)
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	content, err := os.ReadFile(filepath.Join("files", filename))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"filename": filename,
			"content":  string(content),
		})
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	}
}

func listTextFiles() ([]string, error) {
	var textFiles []string

	files, err := os.ReadDir("files")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".txt" {
			textFiles = append(textFiles, file.Name())
		}
	}

	return textFiles, nil
}
