package utils

import (
	"context"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestUtils struct {
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Client client.Client

	Namespace       string
	ConnectionName  string
	ProxyURL        string
	TargetHost      string
	TaggedImageName string
	TestPod         string
}
