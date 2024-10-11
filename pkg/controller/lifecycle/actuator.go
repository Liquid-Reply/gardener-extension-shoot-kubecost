// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/liquid-reply/gardener-extension-shoot-kubecost/kubecost"
	api "github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/apis/config"
	"github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/constants"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/extensions"
	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(c client.Client, decoder runtime.Decoder) extension.Actuator {
	return &actuator{
		client:  c,
		decoder: decoder,
	}
}

type actuator struct {
	client  client.Client // seed cluster
	decoder runtime.Decoder
}

// Reconcile the Extension resource
func (a *actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// get the shoot and the project namespace
	extensionNamespace := ex.GetNamespace()
	shoot, err := extensions.GetShoot(ctx, a.client, extensionNamespace)
	if err != nil {
		return err
	}
	projectNamespace := shoot.GetNamespace()
	logger = logger.WithValues("project", projectNamespace)
	logger.Info("Reconciling")

	kubeCostConfig := &api.Configuration{}
	if _, _, err := a.decoder.Decode(ex.Spec.ProviderConfig.Raw, nil, kubeCostConfig); err != nil {
		return fmt.Errorf("failed to decode provider config: %w", err)
	}

	// Create the resource for the kubecost installation
	shootResourceKubeCostInstall, err := createShootResourceKubeCostInstall(kubeCostConfig)
	if err != nil {
		return err
	}
	// deploy the managed resource for the kubecost installatation
	logger.Info("Creating ManagedResource with KubeCost manifest", "name", constants.ManagedResourceNameKubeCostConfig)
	err = managedresources.CreateForShoot(ctx, a.client, extensionNamespace, constants.ManagedResourceNameKubeCostConfig, "shoot-kubecost", true, shootResourceKubeCostInstall)
	if err != nil {
		return err
	}

	return nil
}

// Delete the Extension resource
func (a *actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	twoMinutes := 2 * time.Minute

	timeoutShootCtx, cancelShootCtx := context.WithTimeout(ctx, twoMinutes)
	defer cancelShootCtx()

	if err := managedresources.SetKeepObjects(ctx, a.client, namespace, constants.ManagedResourceNameKubeCostConfig, false); err != nil {
		return err
	}

	logger.Info("Deleting ManagedResource", "name", constants.ManagedResourceNameKubeCostConfig)
	if err := managedresources.DeleteForShoot(ctx, a.client, namespace, constants.ManagedResourceNameKubeCostConfig); err != nil {
		return err
	}

	logger.Info("Waiting until ManagedResource is deleted", "name", constants.ManagedResourceNameKubeCostConfig)
	if err := managedresources.WaitUntilDeleted(timeoutShootCtx, a.client, namespace, constants.ManagedResourceNameKubeCostConfig); err != nil {
		return err
	}

	return nil
}

// ForceDelete the Extension resource
func (a *actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, logger, ex)
}

// Restore the Extension resource
func (a *actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}

// Migrate the Extension resource
func (a *actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, logger, ex)
}

func getKubeCostConfig(config []byte) (kubecost.KubeCostConfig, error) {
	var out kubecost.KubeCostConfig
	err := yaml.Unmarshal(config, &out)
	return out, err
}

func createShootResourceKubeCostInstall(config *api.Configuration) (map[string][]byte, error) {
	manifest, err := kubecost.Render(config, true)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		"kubecost.br": manifest,
	}, nil
}
