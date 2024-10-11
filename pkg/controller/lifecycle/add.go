// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DefaultAddOptions contains configuration for the shoot-kubecost controller
var DefaultAddOptions = AddOptions{}

// AddOptions are options to apply when adding the shoot-kubecost controller to the manager.
type AddOptions struct {
	// ControllerOptions contains options for the controller.
	ControllerOptions controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManager adds a shoot-kubecost Lifecycle controller to the given Controller Manager.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	decoder := serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder()

	return extension.Add(ctx, mgr, extension.AddArgs{
		Actuator:          NewActuator(mgr.GetClient(), decoder),
		ControllerOptions: DefaultAddOptions.ControllerOptions,
		Name:              "shoot_kubecost_lifecycle_controller",
		FinalizerSuffix:   constants.ExtensionType,
		Resync:            1 * time.Minute,
		Predicates:        extension.DefaultPredicates(ctx, mgr, DefaultAddOptions.IgnoreOperationAnnotation),
		Type:              constants.ExtensionType,
	})
}
