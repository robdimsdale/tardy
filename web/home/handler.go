package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/tardy-tmp"
	"github.com/robdimsdale/wl"
	"github.com/robdimsdale/wl/logger"
	"github.com/robdimsdale/wl/oauth"
)

type Handler interface {
	Home(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger   lager.Logger
	clientID string
	store    *sessions.CookieStore
}

func NewHandler(
	logger lager.Logger,
	clientID string,
	store *sessions.CookieStore,
) Handler {
	return &handler{
		logger:   logger,
		clientID: clientID,
		store:    store,
	}
}

func (h handler) Home(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "session-name")
	if err != nil {
		h.logger.Error("", err)
		http.Error(w, err.Error(), 500)
		return
	}

	accessTokenInterface := session.Values["accessToken"]
	if accessTokenInterface == nil {
		h.logger.Info("accessToken not found in session - redirecting")
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	accessToken, ok := accessTokenInterface.(string)
	if !ok {
		err := fmt.Errorf("failed to convert %v into string", accessTokenInterface)
		h.logger.Error("", err)
		http.Error(w, err.Error(), 500)
		return
	}

	if accessToken == "" {
		h.logger.Info("accessToken empty in session - redirecting")
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	client := oauth.NewClient(
		accessToken,
		h.clientID,
		wl.APIURL,
		logger.NewLogger(logger.INFO),
	)

	completed := true
	completedTasks, err := client.CompletedTasks(completed)
	if err != nil {
		fmt.Printf("err getting tasks: %s\n", err.Error())
	}

	tasks, err := tardyTasks(completedTasks)
	if err != nil {
		fmt.Printf("err converting tasks: %s\n", err.Error())
	}

	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		fmt.Printf("err serializing completed: %s\n", err.Error())
	}
}

func tardyTasks(wlTasks []wl.Task) ([]tardy.Task, error) {
	tasks := []tardy.Task{}
	for _, t := range wlTasks {
		if (t.DueDate != time.Time{}) {
			days := int(t.CompletedAt.Sub(t.DueDate).Hours() / 24)

			tardyTask := tardy.Task{
				ID:          t.ID,
				Title:       t.Title,
				DueDate:     t.DueDate,
				CompletedAt: t.CompletedAt,
				Days:        days,
			}
			tasks = append(tasks, tardyTask)
		}
	}
	return tasks, nil
}
