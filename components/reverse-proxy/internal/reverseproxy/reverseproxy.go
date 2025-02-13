package reverseproxy

import (
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/reverse-proxy/internal/server"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

func handler(p *httputil.ReverseProxy, targetHost string, log *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	log.Infof("Registering handler to %s\n", targetHost)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Asking for %s %s -> %s\n", r.Proto, r.URL, targetHost)
		r.Host = targetHost
		w.Header().Set("X-Forwarded-Host", targetHost)
		p.ServeHTTP(w, r)
	}
}

// New creates a new reverse proxy server
func New(reverseProxyURL, connectivityProxyURL, targetHost string, log *zap.SugaredLogger) (*server.Server, error) {
	remote, err := url.Parse(connectivityProxyURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	httpServer := &http.Server{
		Addr:    reverseProxyURL,
		Handler: http.HandlerFunc(handler(proxy, targetHost, log)),
	}
	return &server.Server{HTTPServer: httpServer, Log: log}, nil
}
