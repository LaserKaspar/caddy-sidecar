package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"main/internal/caddyproc"
)

type Config struct {
	APIKey  string
	RootDir string
}

var cfg Config

func loadConfig() {
	cfg.APIKey = os.Getenv("API_KEY")
	cfg.RootDir = os.Getenv("ROOT_DIR")
	if cfg.RootDir == "" {
		cfg.RootDir = "/etc/caddy"
	}
	if cfg.APIKey == "" {
		log.Fatal("API_KEY must be set")
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != cfg.APIKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// find caddy process by name
	pid, err := caddyproc.FindPID()
	if err != nil {
		if errors.Is(err, caddyproc.ErrNotFound) {
			http.Error(w, "Caddy process not found", http.StatusInternalServerError)
		} else {
			http.Error(w, "Process lookup failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// send SIGUSR1 to reload
	if err := caddyproc.SignalReload(pid); err != nil {
		http.Error(w, "Failed to send signal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "reload signal sent",
		"pid":    pid,
	})
}

func handleFile(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(r.URL.Path, "/file/")
	if relPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	// sanitize path
	fullPath := filepath.Join(cfg.RootDir, filepath.Clean("/"+relPath))
	if !strings.HasPrefix(fullPath, cfg.RootDir) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		fileGet(w, fullPath)
	case http.MethodPut:
		filePut(w, r, fullPath)
	case http.MethodPost:
		filePost(w, r, fullPath)
	case http.MethodDelete:
		fileDelete(w, fullPath)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func fileGet(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Read error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func filePut(w http.ResponseWriter, r *http.Request, path string) {
	if _, err := os.Stat(path); err == nil {
		http.Error(w, "File already exists", http.StatusConflict)
		return
	}
	writeFile(w, r, path)
}

func filePost(w http.ResponseWriter, r *http.Request, path string) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "File does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Stat error", http.StatusInternalServerError)
		}
		return
	}
	writeFile(w, r, path)
}

func writeFile(w http.ResponseWriter, r *http.Request, path string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		http.Error(w, "Cannot create directory", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body error", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(path, body, 0644); err != nil {
		http.Error(w, "Write error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "written"})
}

func fileDelete(w http.ResponseWriter, path string) {
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Delete error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func main() {
	loadConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/reload", handleReload)
	mux.HandleFunc("/file/", handleFile)

	handler := authMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("caddy-sidecar listening on :%s, root dir: %s", port, cfg.RootDir)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
