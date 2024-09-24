// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	_ "embed"
	"errors"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	"github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/constants"

	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator() extension.Actuator {
	return &actuator{
		logger: log.Log.WithName("FirstLogger"),
	}
}

type actuator struct {
	logger          logr.Logger   // logger
	client          client.Client // seed cluster
	clientGardenlet client.Client // garden cluster
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// get the shoot and the project namespace
	extensionNamespace := ex.GetNamespace()
	shoot, err := extensions.GetShoot(ctx, a.client, extensionNamespace)
	if err != nil {
		return err
	}
	projectNamespace := shoot.GetNamespace()

	// fetch the secret holding the per-project configuration for the shoot-kubecost installation
	kubeCostSecret := corev1.Secret{}
	err = a.clientGardenlet.Get(ctx, types.NamespacedName{Namespace: projectNamespace, Name: "shoot-kubecost"}, &kubeCostSecret)
	if err != nil {
		return err
	}

	kubeCostApiKey, err := getKubeCostApiKey(kubeCostSecret.Data)
	if err != nil {
		a.logger.Error(err, "Unable to retrieve the KubeCost api key. Check the secret in the garden cluster for the apiKey field.")
		return err
	}

	// Create the resource for the flux installation
	shootResourceKubeCostInstall, err := createShootResourceKubeCostInstall(kubeCostApiKey)
	if err != nil {
		return err
	}
	// deploy the managed resource for the flux installatation
	err = managedresources.CreateForShoot(ctx, a.client, extensionNamespace, constants.ManagedResourceNameKubeCostConfig, "shoot-kubecost", true, shootResourceKubeCostInstall)
	if err != nil {
		return err
	}

	return nil
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	a.logger.Info("Hello World, I just entered the Delete method")
	return nil
}

func (a *actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	a.logger.Info("Hello World, I just entered the ForceDelete method")
	return nil
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, logger, ex)
}

func getKubeCostApiKey(secretData map[string][]byte) (string, error) {
	if apiKey, ok := secretData["apiKey"]; ok {
		return string(apiKey), nil
	}

	return "", errors.New("apiKey field not found")
}

func createShootResourceKubeCostInstall(apiKey string) (map[string][]byte, error) {
	return nil, nil
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	clientInterface, err := gardenclient.NewClientFromSecret(context.Background(), a.client, "garden", "gardenlet-kubeconfig")
	if err != nil {
		return err
	}
	clientInterface.Start(context.Background())
	a.clientGardenlet = clientInterface.Client()
	return nil
}
