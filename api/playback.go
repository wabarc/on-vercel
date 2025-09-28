package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/service"
)

func Playback(w http.ResponseWriter, r *http.Request) {
	logger.Info("playback request start...")

	if err := r.ParseForm(); err != nil {
		logger.Error("parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		logger.Warn("post form value empty.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	logger.Debug("text: %s", text)

	urls := service.MatchURL(opts, text)
	if len(urls) == 0 {
		logger.Warn("url no found.")
	}
	col, err := wayback.Playback(context.Background(), opts, urls...)
	if err != nil {
		logger.Error("web: playback failed: %v", err)
		return
	}
	collector := transform(col)
	switch r.PostFormValue("data-type") {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		if data, err := json.Marshal(collector); err != nil {
			logger.Error("encode for response failed, %v", err)
		} else {
			w.Write(data) // nolint:errcheck
		}
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if html, ok := tpl.Render("layout", collector); ok {
			w.Write(html) // nolint:errcheck
		} else {
			logger.Error("render template for response failed")
		}
	}
}
