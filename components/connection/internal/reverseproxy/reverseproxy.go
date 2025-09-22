package reverseproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kyma-project/registry-proxy/components/connection/internal/server"

	"go.uber.org/zap"
)

type logRoundTripper struct {
	log       *zap.SugaredLogger
	transport http.RoundTripper
}

func (lrt *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	filteredReqHeaders := getRedactedHeaders(req.Header)
	lrt.log.Debugf("Request: %s %s %s %v\n", req.Proto, req.Host, req.URL, filteredReqHeaders)
	res, err := lrt.transport.RoundTrip(req)
	if err != nil {
		lrt.log.Errorf("Error: %v\n", err)
		return res, err
	}
	lrt.log.Debugf("Response status: %s\n", res.Status)
	filteredRespHeaders := getRedactedHeaders(res.Header)
	lrt.log.Debugf("Response headers: %v", filteredRespHeaders)
	return res, err
}

// getRedactedHeaders returns a copy of the provided headers with sensitive information redacted
func getRedactedHeaders(headers http.Header) http.Header {
	filteredHeaders := make(map[string][]string)
	for key, values := range headers {
		if filteredHeaders[key] == nil {
			filteredHeaders[key] = make([]string, 0)
		}
		for _, v := range values {
			filteredValue := v
			if key == "Authorization" || key == "Proxy-Authorization" {
				filteredValue = "***"
			}
			filteredHeaders[key] = append(filteredHeaders[key], filteredValue)
		}
	}
	return filteredHeaders
}

// getModifyResponseFunc replaces host of the WWW-Authenticate header with localhost:authPort
// example header: Www-Authenticate: Bearer realm="http://gitlab.kyma/jwt/auth",service="container_registry"\r\n
func getModifyResponseFunc(authPort string) func(*http.Response) error {
	return func(resp *http.Response) error {
		authenticateHeader := resp.Header.Get("WWW-Authenticate")
		if authenticateHeader != "" {
			wwwAuthHeader := ParseAuthSettings(authenticateHeader)
			originalDestination := wwwAuthHeader.Params["realm"]
			originalDestinationURL, err := url.Parse(originalDestination)
			if err != nil {
				return err
			}
			newDestination := fmt.Sprintf("localhost:%s", authPort)
			authenticateHeader = strings.Replace(authenticateHeader, originalDestinationURL.Host, newDestination, 1)
			resp.Header.Set("WWW-Authenticate", authenticateHeader)
		}
		return nil
	}
}

func handler(p *httputil.ReverseProxy, targetHost, locationID, authHeader string, log *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	log.Infof("Registering handler to %s\n", targetHost)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Asking for %s %s %s\n", r.Proto, targetHost, r.URL)
		r.Host = targetHost
		r.Header.Set("X-Forwarded-Host", targetHost)

		if locationID != "" {
			r.Header.Set("SAP-Connectivity-SCC-Location_ID", locationID)
		}

		if authHeader != "" {
			r.Header.Set("Authorization", authHeader)
		}

		p.ServeHTTP(w, r)
	}
}

// New creates a new reverse proxy server
func New(reverseProxyURL, connectivityProxyURL, targetHost, locationID, authPort, authorizationHeader string, log *zap.SugaredLogger) (*server.Server, error) {
	remote, err := url.Parse(connectivityProxyURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &logRoundTripper{log: log, transport: http.DefaultTransport}
	proxy.ErrorLog = zap.NewStdLog(log.Desugar())

	if authPort != "" {
		log.Infof("Setting up authorization host to localhost:%s", authPort)
		proxy.ModifyResponse = getModifyResponseFunc(authPort)
	}

	httpServer := &http.Server{
		Addr:    reverseProxyURL,
		Handler: http.HandlerFunc(handler(proxy, targetHost, locationID, authorizationHeader, log)),
	}
	return &server.Server{HTTPServer: httpServer, Log: log}, nil
}
