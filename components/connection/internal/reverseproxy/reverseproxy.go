package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.tools.sap/kyma/registry-proxy/components/connection/internal/server"

	"go.uber.org/zap"
)

type logRoundTripper struct {
	log       *zap.SugaredLogger
	transport http.RoundTripper
}

func (lrt *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	lrt.log.Debugf("Request: %s %s %s %v\n", req.Proto, req.Host, req.URL, req.Header)
	res, err := lrt.transport.RoundTrip(req)
	if err != nil {
		lrt.log.Errorf("Error: %v\n", err)
		return res, err
	}
	lrt.log.Debugf("Response status: %s\n", res.Status)
	lrt.log.Debugf("Response headers: %v", res.Header)
	return res, err
}

func handler(p *httputil.ReverseProxy, targetHost, locationID string, log *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	log.Infof("Registering handler to %s\n", targetHost)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Asking for %s %s %s\n", r.Proto, targetHost, r.URL)
		r.Host = targetHost
		r.Header.Set("X-Forwarded-Host", targetHost)

		if locationID != "" {
			r.Header.Set("SAP-Connectivity-SCC-Location_ID", locationID)
		}

		p.ServeHTTP(w, r)
	}
}

// New creates a new reverse proxy server
func New(reverseProxyURL, connectivityProxyURL, targetHost, locationID string, log *zap.SugaredLogger) (*server.Server, error) {
	remote, err := url.Parse(connectivityProxyURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &logRoundTripper{log: log, transport: http.DefaultTransport}
	proxy.ErrorLog = zap.NewStdLog(log.Desugar())

	httpServer := &http.Server{
		Addr:    reverseProxyURL,
		Handler: http.HandlerFunc(handler(proxy, targetHost, locationID, log)),
	}
	return &server.Server{HTTPServer: httpServer, Log: log}, nil
}
