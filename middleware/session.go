package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pivotal-golang/lager"
)

type auth struct {
	logger        lager.Logger
	cookieHandler *securecookie.SecureCookie
	store         *sessions.CookieStore
}

func NewAuth(
	logger lager.Logger,
	cookieHandler *securecookie.SecureCookie,
	store *sessions.CookieStore,
) Middleware {
	return auth{
		logger:        logger,
		cookieHandler: cookieHandler,
		store:         store,
	}
}

func (s auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if s.unauthenticatedAccessAllowedForURL(req.URL.Path) ||
			s.validSession(rw, req) {
			next.ServeHTTP(rw, req)
		} else {
			s.logger.Debug("not logged in - redirecting")
			http.Redirect(rw, req, "/login", http.StatusFound)
		}
	})
}

func (s auth) unauthenticatedAccessAllowedForURL(url string) bool {
	openURLs := []string{"/login", "/static"}

	for _, u := range openURLs {
		if strings.HasPrefix(url, u) {
			s.logger.Debug("unauthenticated access allowed for URL", lager.Data{"url": url})
			return true
		}
	}
	s.logger.Debug("authenticated access required for URL", lager.Data{"url": url})
	return false
}

func (s auth) validSession(w http.ResponseWriter, r *http.Request) bool {
	var accessToken string

	session, err := s.store.Get(r, "session-name")
	if err != nil {
		s.logger.Error("", err)
		http.Error(w, err.Error(), 500)
		return false
	}

	fmt.Printf("session %+v\n", session.Values)

	accessTokenInterface := session.Values["accessToken"]
	if accessTokenInterface == nil {
		s.logger.Info("accessToken nil in session - redirecting")
		return false
	} else {
		accessToken, ok := accessTokenInterface.(string)
		if !ok {
			err := fmt.Errorf("failed to convert %v into string", accessTokenInterface)
			s.logger.Error("", err)
			http.Error(w, err.Error(), 500)
			return false
		}

		if accessToken == "" {
			s.logger.Info("accessToken empty in session - redirecting")
			return false
		} else {
			s.logger.Debug("accessToken found in session")
			return true
		}
	}

	cookie, err := r.Cookie("session")
	if err == nil {
		cookieValue := make(map[string]string)
		err = s.cookieHandler.Decode("session", cookie.Value, &cookieValue)
		if err == nil {
			accessToken = cookieValue["accessToken"]
			if accessToken != "" {
				session, err := s.store.Get(r, "session-name")
				if err != nil {
					s.logger.Error("", err)
					http.Error(w, err.Error(), 500)
					return false
				}

				session.Values["accessToken"] = accessToken
				session.Save(r, w)
				s.logger.Debug("successfully validated via session")
				return true
			}
		}
	}

	s.logger.Debug("failed validation via session")
	return false
}
