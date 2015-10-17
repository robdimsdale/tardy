package middleware

import (
	"net/http"
	"net/url"

	"github.com/pivotal-golang/lager"
)

type httpsEnforcer struct {
	logger lager.Logger
}

func NewHTTPSEnforcer(logger lager.Logger) Middleware {
	return httpsEnforcer{
		logger: logger.Session("middleware-force-https"),
	}
}

func (h httpsEnforcer) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURL, err := url.Parse(req.URL.String())
		if err != nil {
			h.logger.Error("failed to parse URL", err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		protoHeader := req.Header.Get("X-Forwarded-Proto")
		if reqURL.Scheme != "https" && protoHeader != "https" {
			reqURL.Scheme = "https"
			reqURL.Host = req.Host
			h.logger.Debug("redirecting", lager.Data{"url": reqURL})
			http.Redirect(rw, req, reqURL.String(), http.StatusFound)
		}

		h.logger.Debug("continuing to next handler")

		next.ServeHTTP(rw, req)
	})
}
