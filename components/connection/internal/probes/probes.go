package probes

import (
	"fmt"
	"io"
	"net/http"

	"github.tools.sap/kyma/registry-proxy/components/connection/internal/server"

	"go.uber.org/zap"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// getReadyz checks if we can access the registry proxy connection URL and returns its status code (or 503 if it's unreachable)
func getReadyz(registryProxyConnection string, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(registryProxyConnection)

		if err != nil {
			log.Warnf("couldn't reach registry proxy connection at %s: %v", registryProxyConnection, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(fmt.Sprintf("couldn't reach registry proxy connection at %s: %v", registryProxyConnection, err)))
			return
		}

		defer func() {
			errClose := resp.Body.Close()
			if errClose != nil {
				log.Infof("error closing body of response from to %s: %v", registryProxyConnection, errClose)
			}
		}()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("couldn't read response body from registry proxy connection at %s: %v", registryProxyConnection, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(fmt.Sprintf("couldn't read response body from registry proxy connection at %s: %v", registryProxyConnection, err)))
			return
		}

		log.Debugf("registry proxy connection at %s returned status code %d", registryProxyConnection, resp.StatusCode)
		if string(respBody) != "" && resp.StatusCode != http.StatusOK {
			log.Debugf("registry proxy connection at %s returned body: %s", registryProxyConnection, string(respBody))
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
