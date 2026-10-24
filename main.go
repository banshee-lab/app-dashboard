package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed template.html
var tpl string

type App struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	ImageURL string `yaml:"image-url"`
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func render(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var apps []App
	if err := yaml.Unmarshal(raw, &apps); err != nil {
		return nil, fmt.Errorf("yaml inválido: %w", err)
	}
	t, err := template.New("index").Parse(tpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]any{"Title": env("TITLE", "Apps"), "Apps": apps}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	page, err := render(env("APPS_FILE", "apps.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(page))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Write(page)
	})

	addr := ":" + env("PORT", "8080")
	log.Printf("sirviendo en %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
