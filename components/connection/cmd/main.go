package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/kyma-project/registry-proxy/components/connection/internal/probes"
	"github.com/kyma-project/registry-proxy/components/connection/internal/reverseproxy"
	"github.com/kyma-project/registry-proxy/components/connection/internal/server"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
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

	logger, err := newLogger(logLevel)
	if err != nil {
		fmt.Printf("unable to setup logger: %s", err)
		os.Exit(1)
	}

	logger.Info("Beginning setup")

	if os.Getenv("PROXY_URL") == "" {
		logger.Panic("PROXY_URL env was not set")
	}
	connectivityProxyAddress = os.Getenv("PROXY_URL")

	if os.Getenv("TARGET_HOST") == "" {
		logger.Panic("TARGET_HOST env was not set")
	}
	targetHost = os.Getenv("TARGET_HOST")

	locationID := os.Getenv("LOCATION_ID")

	authPort := os.Getenv("AUTHORIZATION_NODE_PORT")

	// read /secrets/authorization file is it exists and store it in authorizationHeader
	authorizationHeaderData, err := os.ReadFile("/secrets/authorization/authorizationHeader")
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("No authorization header file found, proceeding without it")
		} else {
			logger.Panicf("Error reading authorization header file: %s", err)
		}
	}
	authorizationHeader := string(authorizationHeaderData)

	logger.Infof("Registering reverse proxy on %s through %s", proxyAddr, connectivityProxyAddress)
	reverseProxyServer, err := reverseproxy.New(proxyAddr, connectivityProxyAddress, targetHost, locationID, authPort, authorizationHeader, logger)
	if err != nil {
		log.Panicf("unable to setup reverse proxy: %s", err)
	}

	probesServer := probes.New(probeAddr, reverseProxyServer.HTTPServer.Addr, logger)

	stop := make(chan bool)

	logger.Info("Starting reverse proxy server")
	go reverseProxyServer.Serve(stop)

	logger.Info("Starting probes server")
	go probesServer.Serve(stop)

	// wait for channel to return anything
	shouldStop := <-stop
	if shouldStop {
		logger.Info("one or more servers have closed, stopping all servers")
		err = shutdownServer(reverseProxyServer)
		if err != nil {
			logger.Errorf("error while shutting down reverse proxy server: %v", err)
		}

		err = shutdownServer(probesServer)
		if err != nil {
			logger.Errorf("error while shutting down probes server: %v", err)
		}
	}
	logger.Info("all servers stopped")

	// sanity check
	select {
	case s, ok := <-stop:
		if ok {
			logger.Errorf("channel stop still contains %s", strconv.FormatBool(s))
		} else {
			logger.Info("channel stop is closed")
		}
	default:
		logger.Info("channel stop is empty")
	}
}

func shutdownServer(s *server.Server) error {
	return s.HTTPServer.Shutdown(context.Background())
}

func newLogger(level string) (*zap.SugaredLogger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.Level = logLevel

	logger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("unable to setup logger: %s", err)
		os.Exit(1)
	}
	return logger.Sugar(), nil
}
