package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
	"pdfgithub.com/internal/blobpath"
	"pdfgithub.com/internal/httpx"
)

type App struct {
	log *slog.Logger
	ac  *httpx.Client
	hc  *httpx.Client
}

func New(log *slog.Logger, ac *httpx.Client) *App {
	return &App{
		log: log,
		ac:  ac,
		hc:  httpx.NewClient(),
	}
}

func (a *App) HandleGet(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		a.log.Error("could not parse url", "url", r.URL.String())
		return
	}
	if u.Path == "/" {
		serveIndex(w, r)
		return
	}

	if len(u.Path) < 4 || u.Path[len(u.Path)-4:] != ".pdf" {
		a.log.Debug("missing .pdf extension", "url", r.URL.String())
		redirectIndex(w, r)
		return
	}

	bp, err := blobpath.Parse(u.Path)
	if err != nil {
		a.log.Error("could not parse blob path", "path", u.Path, "error", err)
		redirectIndex(w, r)
		return
	}

	ccs := bp.ContentCandidates(11)
	if len(ccs) == 0 {
		a.log.Error("could not find any candidates for the given path", "path", u.Path, "error", err)
		redirectIndex(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	reqs := make([]*http.Request, len(ccs))
	for i, cc := range ccs {
		reqs[i], err = a.ac.NewRequest(ctx, http.MethodGet, cc.String(), nil)
		if err != nil {
			a.log.Error("could not create request", "path", cc.String(), "error", err)
			continue
		}
	}

	resultCh := make(chan string, 1)
	errCh := make(chan error, len(reqs))

	var eg errgroup.Group
	eg.SetLimit(3)

	for _, req := range reqs {
		// we're using errgroup to limit the number of concurrent requests.
		eg.Go(func() error {
			rsp, rerr := a.ac.Do(req)
			if rerr != nil {
				errCh <- rerr
				return nil
			}
			defer rsp.Body.Close()

			if rsp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
				return nil
			}

			var body struct {
				Type        string `json:"type"`
				DownloadURL string `json:"download_url"`
			}
			if err = json.NewDecoder(rsp.Body).Decode(&body); err != nil {
				errCh <- fmt.Errorf("could not decode response: %w", err)
				return nil
			}

			if body.Type != "file" || body.DownloadURL == "" {
				errCh <- fmt.Errorf("type is not file or download url is empty")
				return nil
			}

			select {
			case resultCh <- body.DownloadURL:
				cancel()
			default:
			}

			return nil
		})
	}
	go func() {
		_ = eg.Wait()
		close(resultCh)
		close(errCh)
	}()

	dlu, ok := <-resultCh
	if !ok {
		// TODO: redirect to an error page? maybe?
		redirectIndex(w, r)
		return
	}

	dlCtx, dlCancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer dlCancel()

	rsp, err := a.hc.Get(dlCtx, dlu)
	if err != nil {
		a.log.Debug("could not make request", "url", dlu, "error", err)
		// TODO: redirect to an error page? or maybe back to github?
		redirectIndex(w, r)
		return
	}
	defer rsp.Body.Close()

	// for some reason some PDFs on GitHub are text/plain :/
	// therefore, we're forcing PDF content-type.
	w.Header().Set("Content-Type", httpx.MIMEPDF)

	if _, err = io.Copy(w, rsp.Body); err != nil {
		a.log.Debug("could not copy body", "error", err)
		// TODO: redirect to an error page? maybe?
		redirectIndex(w, r)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "static/index.html")
}

func redirectIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
