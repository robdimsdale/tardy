package home

import (
	"html/template"
	"net/http"

	"github.com/pivotal-golang/lager"
)

type Handler interface {
	Home(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger    lager.Logger
	templates *template.Template
}

func NewHandler(
	logger lager.Logger,
	templates *template.Template,
) Handler {
	return &handler{
		logger:    logger.Session("handler-home"),
		templates: templates,
	}
}

func (h handler) Home(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("received request")
	h.templates.ExecuteTemplate(w, "homepage", nil)
}
