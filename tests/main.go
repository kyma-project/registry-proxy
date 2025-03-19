package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/iprp"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/iprp/pod"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/logger"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/namespace"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/utils"

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
		Namespace: fmt.Sprintf("iprp-%s", uuid.New().String()),

		ImagePullReverseProxyName: "iprp-test",
		ProxyURL:                  "http://dockerregistry.kyma-system.svc.cluster.local:5000",
		TargetHost:                "dockerregistry.kyma-system.svc.cluster.local:5000",
		ImageName:                 "alpine:3.21.3",
		TestPod:                   "iprp-test-pod",
		Ctx:                       ctx,
		Client:                    client,
		Logger:                    log,
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

	// create image pull reverse proxy
	testutil.Logger.Infof("Creating image pull reverse proxy '%s'", testutil.ImagePullReverseProxyName)
	if err := iprp.Create(testutil); err != nil {
		return err
	}

	// verify image pull reverse proxy
	testutil.Logger.Infof("Verifying iprp '%s'", testutil.ImagePullReverseProxyName)
	if err := utils.WithRetry(testutil, iprp.Verify); err != nil {
		return err
	}

	// create pod with image through iprp
	testutil.Logger.Infof("Creating pod '%s'", testutil.TestPod)
	if err := pod.Create(testutil); err != nil {
		return err
	}

	// verify pod
	testutil.Logger.Infof("Verifying pod '%s'", testutil.TestPod)
	if err := utils.WithRetry(testutil, pod.Verify); err != nil {
		return err
	}

	// delete iprp
	testutil.Logger.Infof("Deleting iprp '%s'", testutil.ImagePullReverseProxyName)
	if err := iprp.Delete(testutil); err != nil {
		return err
	}

	// verify iprp deletion
	testutil.Logger.Infof("Verifying iprp '%s' deletion", testutil.ImagePullReverseProxyName)
	if err := utils.WithRetry(testutil, iprp.VerifyDeletion); err != nil {
		return err
	}

	// cleanup
	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
