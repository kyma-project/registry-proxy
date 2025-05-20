package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.tools.sap/kyma/registry-proxy/tests/common/connection"
	"github.tools.sap/kyma/registry-proxy/tests/common/logger"
	"github.tools.sap/kyma/registry-proxy/tests/common/namespace"
	"github.tools.sap/kyma/registry-proxy/tests/common/utils"
	"github.tools.sap/kyma/registry-proxy/tests/operator/manager"
)

var (
	testTimeout = time.Minute * 10
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
		Namespace:               fmt.Sprintf("rp-%s", uuid.New().String()),
		RegistryProxyName:       "test-registry-proxy",
		SecondRegistryProxyName: "test-registry-proxy-two",
		ConnectionName:          "connection-test",
		ProxyURL:                "http://dockerregistry.kyma-system.svc.cluster.local:5000",
		TargetHost:              "dockerregistry.kyma-system.svc.cluster.local:5000",
		Ctx:                     ctx,
		Client:                  client,
		Logger:                  log,
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

	// create connection
	// **** verify - pod wstał/Cr conenction ready
	// try delete RP CR
	// verify it's stuck in deleting
	// delete connection
	// verify RP is gone

	testutil.Logger.Infof("Creating registry proxy '%s'", testutil.RegistryProxyName)
	if err := manager.Create(testutil); err != nil {
		return err
	}
	testutil.Logger.Infof("Verifying registry proxy '%s'", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, manager.Verify); err != nil {
		return err
	}

	testutil.Logger.Infof("Creating second registry proxy '%s'", testutil.SecondRegistryProxyName)
	if err := manager.CreateSecond(testutil); err != nil {
		return err
	}
	testutil.Logger.Infof("Verifying second registry proxy won't create '%s'", testutil.SecondRegistryProxyName)
	if err := utils.WithRetry(testutil, manager.VerifyStuck); err != nil {
		return err
	}
	testutil.Logger.Infof("Deleting second registry proxy '%s'", testutil.SecondRegistryProxyName)
	if err := manager.DeleteSecond(testutil); err != nil {
		return err
	}

	testutil.Logger.Infof("Creating connection '%s'", testutil.RegistryProxyName)
	if err := connection.Create(testutil); err != nil {
		return err
	}
	testutil.Logger.Infof("Verifying connection '%s'", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, connection.Verify); err != nil {
		return err
	}

	testutil.Logger.Infof("Deleting registry proxy '%s'", testutil.RegistryProxyName)
	if err := manager.Delete(testutil); err != nil {
		// TODO: check if it's stuck in deleting
		return err
	}
	testutil.Logger.Infof("Verifying registry proxy '%s' deletion is stuck", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, manager.VerifyDeletionStuck); err != nil {
		return err
	}

	testutil.Logger.Infof("Deleting connection '%s'", testutil.RegistryProxyName)
	if err := connection.Delete(testutil); err != nil {
		return err
	}
	testutil.Logger.Infof("Verifying connection '%s' deletion", testutil.ConnectionName)
	if err := utils.WithRetry(testutil, connection.VerifyDeletion); err != nil {
		return err
	}

	testutil.Logger.Infof("Verifying registry proxy '%s' deletion", testutil.RegistryProxyName)
	if err := utils.WithRetry(testutil, manager.VerifyDeletion); err != nil {
		return err
	}

	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
