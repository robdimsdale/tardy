package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . Handler

type Handler interface {
	LoginGET(w http.ResponseWriter, r *http.Request)
	LoginResponse(w http.ResponseWriter, r *http.Request)
	LogoutPOST(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger        lager.Logger
	cookieHandler *securecookie.SecureCookie
	cookieMaxAge  int
	store         *sessions.CookieStore
	clientID      string
	clientSecret  string
	redirectURI   string
	secretState   string
}

func NewHandler(
	logger lager.Logger,
	cookieHandler *securecookie.SecureCookie,
	cookieMaxAge int,
	store *sessions.CookieStore,
	clientID string,
	clientSecret string,
	redirectURI string,
	secretState string,
) Handler {
	return &handler{
		logger:        logger.Session("handler-oauth"),
		cookieHandler: cookieHandler,
		cookieMaxAge:  cookieMaxAge,
		store:         store,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   redirectURI,
		secretState:   secretState,
	}
}

func (h handler) LoginGET(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received request at")

	redirectQueryString := fmt.Sprintf(
		"client_id=%s&redirect_uri=%s&state=%s",
		h.clientID,
		h.redirectURI,
		url.QueryEscape(h.secretState),
	)

	http.Redirect(
		w,
		r,
		fmt.Sprintf(
			"https://www.wunderlist.com/oauth/authorize?%s",
			redirectQueryString,
		),
		http.StatusFound,
	)
}

func (h handler) LoginResponse(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received login response", lager.Data{"url": r.URL})

	values := r.URL.Query()
	returnedState := values["state"][0]
	if returnedState != h.secretState {
		// No need to leak any info if we are being impersonated
		err := fmt.Errorf("returned state %s did not match expected state %s - returning 404", returnedState, h.secretState)
		h.logger.Error("", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	code := values["code"][0]

	bodyString := fmt.Sprintf(
		`{"client_id":"%s","client_secret":"%s","code":"%s"}`,
		h.clientID,
		h.clientSecret,
		code,
	)

	h.logger.Debug("Exchanging code for accessToken with Wunderlist", lager.Data{"code": code})

	resp, err := http.Post(
		"https://www.wunderlist.com/oauth/access_token",
		"application/json",
		bytes.NewBuffer([]byte(bodyString)),
	)

	if err != nil {
		fmt.Printf("error communicating with Wunderlist: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var accessTokenResp accessTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&accessTokenResp)
	if err != nil {
		fmt.Printf("error unmarshalling accessToken response: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Debug("completed code exchange: received access_token", lager.Data{"accessToken": accessTokenResp.AccessToken})

	h.setSession(accessTokenResp.AccessToken, r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h handler) LogoutPOST(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h handler) setSession(
	accessToken string,
	r *http.Request,
	w http.ResponseWriter,
) {
	value := map[string]string{
		"accessToken": accessToken,
	}
	encoded, err := h.cookieHandler.Encode("session", value)
	if err == nil {
		cookie := &http.Cookie{
			Name:   "session",
			Value:  encoded,
			Path:   "/",
			MaxAge: h.cookieMaxAge,
		}
		http.SetCookie(w, cookie)

		h.logger.Debug("Getting session")

		session, err := h.store.Get(r, "session-name")
		if err != nil {
			h.logger.Error("", err)
			http.Error(w, err.Error(), 500)
			return
		}

		h.logger.Debug(
			"Setting access token in session",
			lager.Data{
				"accessToken": accessToken,
			},
		)

		session.Values["accessToken"] = accessToken
		session.Save(r, w)

		h.logger.Debug(
			"Successfuly saved session",
			lager.Data{
				"sessionName": "session-name",
			},
		)

		session, err = h.store.Get(r, "session-name")
		if err != nil {
			h.logger.Error("", err)
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Printf("session %+v\n", session.Values)

		accessTokenInterface := session.Values["accessToken"]
		if accessTokenInterface == nil {
			h.logger.Info("accessToken nil in session - redirecting")
			return
		} else {
			accessToken, ok := accessTokenInterface.(string)
			if !ok {
				err := fmt.Errorf("failed to convert %v into string", accessTokenInterface)
				h.logger.Error("", err)
				http.Error(w, err.Error(), 500)
				return
			}

			if accessToken == "" {
				h.logger.Info("accessToken empty in session - redirecting")
				return
			} else {
				h.logger.Debug("accessToken found in session", lager.Data{"accessToken": accessToken})
				return
			}
		}
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}
