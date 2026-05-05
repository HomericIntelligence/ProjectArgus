package mnemosyne

import (
	"bytes"

	"github.com/yuin/goldmark"
)

// RenderMarkdown converts markdown to HTML with unsafe HTML disabled.
// IMPORTANT: do NOT call goldmark.WithUnsafe() — it enables raw HTML passthrough (XSS risk).
func RenderMarkdown(md string) (string, error) {
	g := goldmark.New() // default: Unsafe=false, raw HTML stripped
	var buf bytes.Buffer
	if err := g.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
