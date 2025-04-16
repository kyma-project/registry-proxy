package probes

import (
	"fmt"
	"io"
	"net/http"

	"github.tools.sap/kyma/registry-proxy/components/reverse-proxy/internal/server"

	"go.uber.org/zap"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// getReadyz checks if we can access the reverse proxy URL and returns its status code (or 503 if it's unreachable)
func getReadyz(reverseProxyURL string, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(reverseProxyURL)

		if err != nil {
			log.Warnf("couldn't reach reverse proxy at %s: %v", reverseProxyURL, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(fmt.Sprintf("couldn't reach reverse proxy at %s: %v", reverseProxyURL, err)))
			return
		}

		defer func() {
			errClose := resp.Body.Close()
			if errClose != nil {
				log.Infof("error closing body of response from to %s: %v", reverseProxyURL, errClose)
			}
		}()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("couldn't read response body from reverse proxy at %s: %v", reverseProxyURL, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(fmt.Sprintf("couldn't read response body from reverse proxy at %s: %v", reverseProxyURL, err)))
			return
		}

		log.Debugf("reverse proxy at %s returned status code %d", reverseProxyURL, resp.StatusCode)
		if string(respBody) != "" && resp.StatusCode != http.StatusOK {
			log.Debugf("reverse proxy at %s returned body: %s", reverseProxyURL, string(respBody))
		}

		w.WriteHeader(resp.StatusCode)
	}
}

func newProbesMuxer(reverseProxyURL string, log *zap.SugaredLogger) *http.ServeMux {
	muxer := http.NewServeMux()

	muxer.HandleFunc("/healthz", healthz)
	muxer.HandleFunc("/readyz", getReadyz(fmt.Sprintf("%s%s", "http://localhost", reverseProxyURL), log))
	return muxer
}

// New creates a new probes server
func New(probesURL, reverseProxyURL string, log *zap.SugaredLogger) *server.Server {
	muxer := newProbesMuxer(reverseProxyURL, log)
	httpServer := http.Server{
		Addr:    probesURL,
		Handler: muxer,
	}
	return &server.Server{HTTPServer: &httpServer, Log: log}
}
