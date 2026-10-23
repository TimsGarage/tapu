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
	shell  []byte
	assets http.Handler
}

func NewHandler() (*Handler, error) {
	shell, err := distribution.ReadFile(
		"dist/index.html",
	)
	if err != nil {
		return nil, err
	}

	distFS, err := fs.Sub(
		distribution,
		"dist",
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		shell: shell,
		assets: http.StripPrefix(
			"/_app/",
			http.FileServer(
				http.FS(
					mustSubFS(distFS, "_app"),
				),
			),
		),
	}, nil
}

func mustSubFS(
	root fs.FS,
	directory string,
) fs.FS {
	subdirectory, err := fs.Sub(
		root,
		directory,
	)
	if err != nil {
		panic(err)
	}

	return subdirectory
}

func (h *Handler) Shell(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	// SvelteKit's static output contains its own hash-based CSP
	// in a meta element. A generic script-src policy from the Go
	// middleware would additionally block its bootstrap script.
	writer.Header().Del(
		"Content-Security-Policy",
	)

	// frame-ancestors is not reliably enforced from a CSP meta
	// element, so retain it as an HTTP response header.
	writer.Header().Set(
		"Content-Security-Policy",
		"frame-ancestors 'none'",
	)
	writer.Header().Set(
		"X-Frame-Options",
		"DENY",
	)
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

	_, _ = writer.Write(h.shell)
}

func (h *Handler) Asset(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if !strings.HasPrefix(
		request.URL.Path,
		"/_app/",
	) {
		http.NotFound(writer, request)
		return
	}

	h.assets.ServeHTTP(writer, request)
}
