package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/robdimsdale/tardy/api/tasks"
	"github.com/robdimsdale/tardy/filesystem"
	"github.com/robdimsdale/tardy/logger"
	"github.com/robdimsdale/tardy/middleware"
	"github.com/robdimsdale/tardy/web/generated/static"
	"github.com/robdimsdale/tardy/web/home"
	"github.com/robdimsdale/tardy/web/oauth"
)

var (
	state string

	accessToken string
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "12345"
	}

	redirectHost := os.Getenv("REDIRECT_HOST")
	if redirectHost == "" {
		redirectHost = fmt.Sprintf("http://localhost:%s", port)
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Printf("clientID and clientSecret must be provided and non-empty\n")
		os.Exit(2)
	}

	logger, _, err := logger.InitializeLogger(logger.LogLevel("debug"))
	if err != nil {
		fmt.Printf("Failed to initialize logger\n")
		panic(err)
	}

	fmt.Printf("clientID: %s\n", clientID)
	fmt.Printf("clientSecret: %s\n", clientID)

	fmt.Printf("port: %s\n", port)

	state, err = createState()
	if err != nil {
		panic(err)
	}
	fmt.Printf("state: %s\n", state)

	oauthRedirectURI := fmt.Sprintf("%s/login-resp", redirectHost)

	cookieHandler := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	cookieStore := sessions.NewCookieStore([]byte("something-very-secret"))

	templates, err := filesystem.LoadTemplates()
	if err != nil {
		logger.Fatal("exiting", err)
	}

	homeHandler := home.NewHandler(logger, templates)
	tasksHandler := tasks.NewHandler(logger, clientID, cookieStore)

	cookieMaxAge := 3600
	loginHandler := oauth.NewHandler(
		logger,
		cookieHandler,
		cookieMaxAge,
		cookieStore,
		clientID,
		clientSecret,
		oauthRedirectURI,
		state,
	)

	staticFileServer := http.FileServer(static.FS(false))

	rtr := mux.NewRouter()

	rtr.PathPrefix("/static/").Handler(staticFileServer)

	rtr.HandleFunc("/", homeHandler.Home).Methods("GET")

	rtr.HandleFunc("/login", loginHandler.LoginGET).Methods("GET")
	rtr.HandleFunc("/login-resp", loginHandler.LoginResponse).Methods("GET")
	rtr.HandleFunc("/logout", loginHandler.LogoutPOST).Methods("POST")

	a := rtr.PathPrefix("/api/v1").Subrouter()
	a.HandleFunc("/tasks", tasksHandler.Tasks).Methods("GET")

	m := middleware.Chain{
		middleware.NewAuth(logger, cookieHandler, cookieStore),
	}

	handler := m.Wrap(rtr)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
	panic(err)
}

func createState() (string, error) {
	stateBytes := make([]byte, 64)
	for i := range stateBytes {
		val, err := rand.Int(rand.Reader, big.NewInt('~'-'!'))
		if err != nil {
			return "", err
		}
		stateBytes[i] = byte(val.Int64()) + '!'
	}

	return string(stateBytes), nil
}
