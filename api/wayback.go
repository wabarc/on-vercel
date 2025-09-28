package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template"
)

var (
	parser  = config.NewParser()
	opts, _ = parser.ParseEnvironmentVariables()
	tpl     = template.New(mux.NewRouter())
)

func Wayback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusNotModified)
		logger.Error("httpd: request method no specific.")
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		logger.Error("httpd: parse form error")
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		logger.Error("httpd: post form value empty")
		return
	}
	logger.Debug("text: %s", text)

	urls := service.MatchURL(opts, text)
	if len(urls) == 0 {
		logger.Warn("url no found.")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		collector := transform(cols)
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
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	service.Wayback(ctx, opts, urls, do)
}

func transform(cols []wayback.Collect) template.Collector {
	collects := []template.Collect{}
	for _, col := range cols {
		collects = append(collects, template.Collect{
			Slot: col.Arc,
			Src:  col.Src,
			Dst:  col.Dst,
		})
	}
	return collects
}
