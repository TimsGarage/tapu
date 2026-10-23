package adminui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distribution embed.FS

type Handler struct {
	indexHTML []byte
	assets    http.Handler
}

func NewHandler() (*Handler, error) {
	indexHTML, err := distribution.ReadFile(
		"dist/index.html",
	)
	if err != nil {
		return nil, err
	}

	distributionFS, err := fs.Sub(
		distribution,
		"dist",
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		indexHTML: indexHTML,
		assets: http.StripPrefix(
			"/ui/",
			http.FileServer(
				http.FS(distributionFS),
			),
		),
	}, nil
}

func (h *Handler) Shell(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writer.Header().Set(
		"Content-Type",
		"text/html; charset=utf-8",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)
	writer.Header().Set(
		"X-Content-Type-Options",
		"nosniff",
	)

	_, _ = writer.Write(h.indexHTML)
}

func (h *Handler) Asset(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if !strings.HasPrefix(
		request.URL.Path,
		"/ui/",
	) {
		http.NotFound(writer, request)
		return
	}

	h.assets.ServeHTTP(writer, request)
}
