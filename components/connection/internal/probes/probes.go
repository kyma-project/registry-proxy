package probes

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/registry-proxy/components/connection/internal/server"

	"go.uber.org/zap"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// getReadyz checks if we can access the registry proxy connection URL and returns its status code (or 503 if it's unreachable)
func getReadyz(registryProxyConnection string, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// don't follow redirects
		c := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := c.Get(registryProxyConnection)

		if err != nil {
			log.Warnf("couldn't reach registry proxy connection at %s: %v", registryProxyConnection, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, "couldn't reach registry proxy connection at %s: %v", registryProxyConnection, err)
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
			_, _ = fmt.Fprintf(w, "couldn't read response body from registry proxy connection at %s: %v", registryProxyConnection, err)
			return
		}

		log.Debugf("registry proxy connection at %s returned status code %d", registryProxyConnection, resp.StatusCode)
		if string(respBody) != "" && resp.StatusCode != http.StatusOK {
			log.Debugf("registry proxy connection at %s returned body: %s", registryProxyConnection, string(respBody))
		}

		// we only want to check if the target system is responsice, we don't care if the response is non 200, as long as server works
		if resp.StatusCode < http.StatusInternalServerError {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(resp.StatusCode)
		}
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
