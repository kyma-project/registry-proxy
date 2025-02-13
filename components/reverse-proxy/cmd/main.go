package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.tools.sap/kyma/image-pull-reverse-proxy/components/reverse-proxy/internal/probes"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/reverse-proxy/internal/reverseproxy"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/reverse-proxy/internal/server"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var proxyAddr string
	var probeAddr string
	var connectivityProxyAddress string
	var targetHost string

	flag.StringVar(&proxyAddr, "reverse-proxy-bind-address", ":8080", "The address the reverse proxy binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.Parse()

	// default log level to warn if user didn't specify one
	// TODO: change to warn after LOG_LEVEL is available in the CR
	logLevel := "debug"
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

	reverseProxyServer, err := reverseproxy.New(proxyAddr, connectivityProxyAddress, targetHost, logger)
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
			logger.Errorf("channel stop still contains %s.\n", strconv.FormatBool(s))
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

	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.Level = logLevel

	logger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("unable to setup logger: %s", err)
		os.Exit(1)
	}
	return logger.Sugar(), nil
}
