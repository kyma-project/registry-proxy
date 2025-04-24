package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.tools.sap/kyma/registry-proxy/tests/logger"
	"github.tools.sap/kyma/registry-proxy/tests/namespace"
	"github.tools.sap/kyma/registry-proxy/tests/rp"
	"github.tools.sap/kyma/registry-proxy/tests/rp/pod"
	"github.tools.sap/kyma/registry-proxy/tests/utils"

	"github.com/google/uuid"
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

		RegistryProxyName: "rp-test",
		ProxyURL:          "http://dockerregistry.kyma-system.svc.cluster.local:5000",
		TargetHost:        "dockerregistry.kyma-system.svc.cluster.local:5000",
		ImageName:         "alpine:3.21.3",
		TestPod:           "rp-test-pod",
		Ctx:               ctx,
		Client:            client,
		Logger:            log,
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runScenario(testutil *utils.TestUtils) error {
	// create test namespace
	testutil.Logger.Infof("Creating namespace '%s'", testutil.Namespace)
	if err := namespace.Create(testutil); err != nil {
		return err
	}

	// create registry proxy
	testutil.Logger.Infof("Creating registry proxy '%s'", testutil.RegistryProxyName)
	if err := rp.Create(testutil); err != nil {
		return err
	}

	// verify registry proxy
	testutil.Logger.Infof("Verifying rp '%s'", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, rp.Verify); err != nil {
		return err
	}

	// create pod with image through rp
	testutil.Logger.Infof("Creating pod '%s'", testutil.TestPod)
	if err := pod.Create(testutil); err != nil {
		return err
	}

	// verify pod
	testutil.Logger.Infof("Verifying pod '%s'", testutil.TestPod)
	if err := utils.WithRetry(testutil, pod.Verify); err != nil {
		return err
	}

	// delete rp
	testutil.Logger.Infof("Deleting rp '%s'", testutil.RegistryProxyName)
	if err := rp.Delete(testutil); err != nil {
		return err
	}

	// verify rp deletion
	testutil.Logger.Infof("Verifying rp '%s' deletion", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, rp.VerifyDeletion); err != nil {
		return err
	}

	// cleanup
	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
