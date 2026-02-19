package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/kyma-project/registry-proxy/components/common/fips"
	"github.com/kyma-project/registry-proxy/components/connection/internal/probes"
	"github.com/kyma-project/registry-proxy/components/connection/internal/reverseproxy"
	"github.com/kyma-project/registry-proxy/components/connection/internal/server"

	"github.com/kyma-project/manager-toolkit/logging/logger"
	"go.uber.org/zap"
)

func main() {
	if !fips.IsFIPS140Only() {
		log.Panic("FIPS 140 exclusive mode is not enabled. Check GODEBUG flags.")
	}
	var proxyAddr string
	var probeAddr string
	var connectivityProxyAddress string
	var targetHost string

	flag.StringVar(&proxyAddr, "connection-bind-address", ":8080", "The address the registry proxy connection binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.Parse()

	// default log level to info if user didn't specify one
	logLevel := "info"
	if os.Getenv("LOG_LEVEL") != "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}

	zapLogger, err := newLogger(logLevel)
	if err != nil {
		fmt.Printf("unable to setup logger: %s", err)
		os.Exit(1)
	}

	zapLogger.Info("Beginning setup")

	if os.Getenv("PROXY_URL") == "" {
		zapLogger.Panic("PROXY_URL env was not set")
	}
	connectivityProxyAddress = os.Getenv("PROXY_URL")

	if os.Getenv("TARGET_HOST") == "" {
		zapLogger.Panic("TARGET_HOST env was not set")
	}
	targetHost = os.Getenv("TARGET_HOST")

	locationID := os.Getenv("LOCATION_ID")

	authPort := os.Getenv("AUTHORIZATION_NODE_PORT")

	// read /secrets/authorization file is it exists and store it in authorizationHeader
	authorizationHeaderData, err := os.ReadFile("/secrets/authorization/authorizationHeader")
	if err != nil {
		if os.IsNotExist(err) {
			zapLogger.Info("No authorization header file found, proceeding without it")
		} else {
			zapLogger.Panicf("Error reading authorization header file: %s", err)
		}
	}
	authorizationHeader := string(authorizationHeaderData)

	zapLogger.Infof("Registering reverse proxy on %s through %s", proxyAddr, connectivityProxyAddress)
	reverseProxyServer, err := reverseproxy.New(proxyAddr, connectivityProxyAddress, targetHost, locationID, authPort, authorizationHeader, zapLogger)
	if err != nil {
		log.Panicf("unable to setup reverse proxy: %s", err)
	}

	probesServer := probes.New(probeAddr, reverseProxyServer.HTTPServer.Addr, zapLogger)

	stop := make(chan bool)

	zapLogger.Info("Starting reverse proxy server")
	go reverseProxyServer.Serve(stop)

	zapLogger.Info("Starting probes server")
	go probesServer.Serve(stop)

	// wait for channel to return anything
	shouldStop := <-stop
	if shouldStop {
		zapLogger.Info("one or more servers have closed, stopping all servers")
		err = shutdownServer(reverseProxyServer)
		if err != nil {
			zapLogger.Errorf("error while shutting down reverse proxy server: %v", err)
		}

		err = shutdownServer(probesServer)
		if err != nil {
			zapLogger.Errorf("error while shutting down probes server: %v", err)
		}
	}
	zapLogger.Info("all servers stopped")

	// sanity check
	select {
	case s, ok := <-stop:
		if ok {
			zapLogger.Errorf("channel stop still contains %s", strconv.FormatBool(s))
		} else {
			zapLogger.Info("channel stop is closed")
		}
	default:
		zapLogger.Info("channel stop is empty")
	}
}

func shutdownServer(s *server.Server) error {
	return s.HTTPServer.Shutdown(context.Background())
}

func newLogger(level string) (*zap.SugaredLogger, error) {
	logLevel, err := logger.MapLevel(level)
	if err != nil {
		return nil, err
	}

	zapLog, err := logger.New(logger.JSON, logLevel)
	if err != nil {
		return nil, err
	}
	return zapLog.WithContext(), nil
}
