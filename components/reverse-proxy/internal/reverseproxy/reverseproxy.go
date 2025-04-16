package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.tools.sap/kyma/registry-proxy/components/reverse-proxy/internal/server"

	"go.uber.org/zap"
)

type LogRoundTripper struct {
	log       *zap.SugaredLogger
	transport http.RoundTripper
}

func (lrt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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

func NewRT(rt http.RoundTripper, log *zap.SugaredLogger) http.RoundTripper {
	return &LogRoundTripper{log: log, transport: rt}
}

func handler(p *httputil.ReverseProxy, targetHost string, log *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	log.Infof("Registering handler to %s\n", targetHost)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Asking for %s %s %s\n", r.Proto, targetHost, r.URL)
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
	proxy.Transport = NewRT(http.DefaultTransport, log)
	proxy.ErrorLog = zap.NewStdLog(log.Desugar())

	httpServer := &http.Server{
		Addr:    reverseProxyURL,
		Handler: http.HandlerFunc(handler(proxy, targetHost, log)),
	}
	return &server.Server{HTTPServer: httpServer, Log: log}, nil
}
