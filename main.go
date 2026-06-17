package main

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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
	logger := slog.New(handler)

	err := godotenv.Load()
	if err != nil {
		logger.Info("could not load .env file", "error", err)
	}

	if os.Getenv("APP_ENV") != "local" {
		logLevel.Set(slog.LevelInfo)
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
		logger.Error("invalid port", "port", port)
		os.Exit(1)
	}

	hc := &http.Client{}

	mux := http.NewServeMux()

	serveHome := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "static/index.html")
	}

	redirectHome := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			logger.Error("could not parse url", "url", r.URL.String())
			return
		}
		if u.Path == "/" {
			serveHome(w, r)
			return
		}

		u.Path = strings.Replace(u.Path, "blob/", "", 1)
		u.Scheme = "https"
		u.Host = "raw.githubusercontent.com"
		u.Query().Add("raw", "true")

		if u.Path[len(u.Path)-4:] != ".pdf" {
			logger.Error("missing .pdf extension", "url", r.URL.String())
			redirectHome(w, r)
			return
		}

		logger.Info("redirecting", "url", u.String())

		rsp, err := hc.Get(u.String())
		if err != nil {
			logger.Error("could not make request", "url", u.String(), "error", err)
			// TODO: redirect to an error page? maybe?
			redirectHome(w, r)
			return
		}
		defer rsp.Body.Close()

		if _, err = io.Copy(w, rsp.Body); err != nil {
			logger.Error("could not copy body", "error", err)
			// TODO: redirect to an error page? maybe?
			redirectHome(w, r)
		}
	})

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static/"))))

	server := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: mux,
	}

	if err = server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
