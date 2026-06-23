package main

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"pdfgithub.com/internal/app"
	"pdfgithub.com/internal/httpx"
)

const (
	DefaultHost = "0.0.0.0"
	DefaultPort = "80"
)

func main() {
	var logLevel slog.LevelVar
	logLevel.Set(slog.LevelInfo)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     &logLevel,
	})
	log := slog.New(handler)

	err := godotenv.Load()
	if err != nil {
		log.Info("could not load .env file", "error", err)
	}

	if os.Getenv("APP_ENV") == "local" {
		logLevel.Set(slog.LevelDebug)
	}

	host := os.Getenv("HTTP_BIND_HOST")
	if host == "" {
		host = DefaultHost
	}

	port := os.Getenv("HTTP_BIND_PORT")
	if port == "" {
		port = DefaultPort
	}
	if _, err = strconv.Atoi(port); err != nil {
		log.Error("invalid port", "port", port)
		os.Exit(1)
	}

	ght := os.Getenv("GH_TOKEN")
	if ght == "" {
		log.Error("missing github token")
		os.Exit(1)
	}

	ac := httpx.NewClient(
		httpx.WithBaseURL("https://api.github.com"),
		httpx.WithHeader("Authorization", "Bearer "+ght),
		httpx.WithHeader("X-GitHub-Api-Version", "2026-03-10"),
	)

	a := app.New(log, ac)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static/"))))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	mux.HandleFunc("GET /", a.HandleGet)

	server := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: mux,
	}

	if err = server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
