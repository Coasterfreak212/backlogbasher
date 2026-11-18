package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/backlogbasher/backlogbasher/internal/api"
	"github.com/backlogbasher/backlogbasher/internal/db"
	"github.com/backlogbasher/backlogbasher/web"
)

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func main() {
	loadDotEnv(".env")

	port := envOrDefault("PORT", "8080")
	dbPath := envOrDefault("DB_PATH", "./data/consoleerp.db")
	uploadDir := envOrDefault("UPLOAD_DIR", "./data/uploads")

	log.Printf("[BacklogBasher] Initializing Home Gaming ERP...")
	log.Printf("[BacklogBasher] Config -> PORT=%s | DB_PATH=%s | UPLOAD_DIR=%s", port, dbPath, uploadDir)

	store, err := db.NewStore(dbPath)
	if err != nil {
		log.Fatalf("[BacklogBasher] Failed to initialize SQLite store: %v", err)
	}
	defer store.Close()

	staticFS, err := web.StaticFS()
	if err != nil {
		log.Fatalf("[BacklogBasher] Failed to load embedded web assets: %v", err)
	}

	srvHandler, err := api.NewServer(store, uploadDir, staticFS)
	if err != nil {
		log.Fatalf("[BacklogBasher] Failed to initialize HTTP router: %v", err)
	}

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srvHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[BacklogBasher] Listening on http://0.0.0.0:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[BacklogBasher] HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("[BacklogBasher] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
