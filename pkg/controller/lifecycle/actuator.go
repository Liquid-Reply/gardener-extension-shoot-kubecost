// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/go-logr/logr"

	"github.com/liquid-reply/gardener-extension-shoot-kubecost/kubecost"
	"github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/constants"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(client client.Client) extension.Actuator {
	clientGardenlet := client
	clientInterface, err := gardenclient.NewClientFromSecret(context.Background(), client, "garden", "gardenlet-kubeconfig")
	if err == nil {
		clientInterface.Start(context.Background())
		clientGardenlet = clientInterface.Client()
	}
	return &actuator{
		client:          client,
		clientGardenlet: clientGardenlet,
	}
}

type actuator struct {
	client          client.Client // seed cluster
	clientGardenlet client.Client // garden cluster
	decoder         runtime.Decoder
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

	// fetch the secret holding the per-project configuration for the shoot-kubecost installation
	kubeCostConfigMap := corev1.ConfigMap{}
	err = a.clientGardenlet.Get(ctx, types.NamespacedName{Namespace: projectNamespace, Name: "shoot-kubecost"}, &kubeCostConfigMap)
	if err != nil {
		logger.Error(err, "Unable to retrieve the KubeCost config. Make sure the configmap shoot-kubecost exists in the project namespace.")
		return err
	}

	kubeCostConfig, err := getKubeCostConfig(kubeCostConfigMap.Data)
	if err != nil {
		logger.Error(err, "Unable to retrieve the KubeCost config. Check the configmap shoot-kubecost in the garden cluster for the config field.")
		return err
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

func getKubeCostConfig(cmData map[string]string) (kubecost.KubeCostConfig, error) {
	config, ok := cmData["config"]
	if !ok {
		return kubecost.KubeCostConfig{}, errors.New("config field not found")
	}

	var out kubecost.KubeCostConfig
	err := yaml.Unmarshal([]byte(config), &out)
	return out, err
}

func createShootResourceKubeCostInstall(config kubecost.KubeCostConfig) (map[string][]byte, error) {
	manifest, err := kubecost.Render(config, true)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		"kubecost.br": manifest,
	}, nil
}
