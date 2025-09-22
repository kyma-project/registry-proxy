package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/registry-proxy/tests/common/logger"
	"github.com/kyma-project/registry-proxy/tests/common/namespace"
	"github.com/kyma-project/registry-proxy/tests/common/utils"

	"github.com/google/uuid"
	"github.com/kyma-project/registry-proxy/tests/common/connection"
	"github.com/kyma-project/registry-proxy/tests/common/connection/pod"
)

var (
	testTimeout = time.Minute * 5
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	log, err := logger.New()
	if err != nil {
		fmt.Printf("%s: %s\n", "unable to setup logger", err)
		os.Exit(1)
	}

	log.Info("Configuring test essentials")
	client, err := utils.GetKuberentesClient()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("Start scenario")
	err = runScenario(&utils.TestUtils{
		Namespace: fmt.Sprintf("rp-%s", uuid.New().String()),

		ConnectionName:  "connection-test",
		ProxyURL:        "http://dockerregistry.kyma-system.svc.cluster.local:5000",
		TargetHost:      "dockerregistry.kyma-system.svc.cluster.local:5000",
		TaggedImageName: "alpine:3.21.3",
		TestPod:         "connection-test-pod",
		Ctx:             ctx,
		Client:          client,
		Logger:          log,
		AuthToken:       true,
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runScenario(testutil *utils.TestUtils) error {
	testutil.Logger.Infof("Creating namespace '%s'", testutil.Namespace)
	if err := namespace.Create(testutil); err != nil {
		return err
	}

	testutil.Logger.Infof("Creating registry proxy's connection '%s'", testutil.ConnectionName)
	if err := connection.Create(testutil); err != nil {
		return err
	}

	testutil.Logger.Infof("Verifying connection '%s'", testutil.ConnectionName)
	if err := utils.WithRetry(testutil, connection.Verify); err != nil {
		return err
	}

	testutil.Logger.Infof("Creating pod '%s'", testutil.TestPod)
	if err := pod.Create(testutil); err != nil {
		return err
	}

	testutil.Logger.Infof("Verifying pod '%s'", testutil.TestPod)
	if err := utils.WithRetry(testutil, pod.Verify); err != nil {
		return err
	}

	testutil.Logger.Infof("Deleting connection '%s'", testutil.ConnectionName)
	if err := connection.Delete(testutil); err != nil {
		return err
	}

	testutil.Logger.Infof("Verifying connection '%s' deletion", testutil.ConnectionName)
	if err := utils.WithRetry(testutil, connection.VerifyDeletion); err != nil {
		return err
	}

	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
