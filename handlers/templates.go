package handlers

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

//go:embed templates
var templatesFS embed.FS

var (
	pageTmpls     map[string]*template.Template
	fragmentTmpls map[string]*template.Template
)

func init() {
	pageTmpls = map[string]*template.Template{
		"index": mustParsePage("index"),
		"event": template.Must(template.ParseFS(templatesFS,
			"templates/base.html",
			"templates/event.html",
			"templates/results.html",
			"templates/join.html",
		)),
	}
	fragmentTmpls = map[string]*template.Template{
		"results": mustParseFragment("results"),
		"join":    mustParseFragment("join"),
	}
}

func mustParsePage(name string) *template.Template {
	return template.Must(template.ParseFS(templatesFS,
		"templates/base.html",
		"templates/"+name+".html",
	))
}

func mustParseFragment(name string) *template.Template {
	return template.Must(template.ParseFS(templatesFS, "templates/"+name+".html"))
}

func renderPage(e *core.RequestEvent, name string, data any) error {
	tmpl, ok := pageTmpls[name]
	if !ok {
		return e.HTML(http.StatusNotFound, "page not found")
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	return e.HTML(http.StatusOK, buf.String())
}

func renderFragment(e *core.RequestEvent, name string, data any) error {
	tmpl, ok := fragmentTmpls[name]
	if !ok {
		return e.HTML(http.StatusNotFound, "fragment not found")
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	return e.HTML(http.StatusOK, buf.String())
}
