// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package network

import (
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// WebhookName is the webhook name.
	WebhookName = "network"
)

var logger = log.Log.WithName("network-webhook")

// AddArgs are arguments for adding a network webhook to a manager.
type AddArgs struct {
	// NetworkProvider is the network provider for this webhook
	NetworkProvider string
	// CloudProvider is the cloud provider of this webhook.
	CloudProvider string
	// Types is a list of resource types.
	Types []runtime.Object
	// Mutator is a mutator to be used by the admission handler.
	Mutator extensionswebhook.Mutator
}

// Add creates a new network webhook and adds it to the given Manager.
func Add(mgr manager.Manager, args AddArgs) (*extensionswebhook.Webhook, error) {
	logger := logger.WithValues("network-provider", args.NetworkProvider, "cloud-provider", args.CloudProvider)

	// Create handler
	handler, err := extensionswebhook.NewHandler(mgr, args.Types, args.Mutator, logger)
	if err != nil {
		return nil, err
	}

	// Build namespace selector from the webhook kind and provider
	namespaceSelector, err := buildSelector(args.NetworkProvider, args.CloudProvider)
	if err != nil {
		return nil, err
	}

	// Create webhook
	var (
		name = WebhookName
		path = WebhookName
	)

	logger.Info("Creating network webhook", "name", name)
	return &extensionswebhook.Webhook{
		Name:     name,
		Provider: args.NetworkProvider,
		Types:    args.Types,
		Target:   extensionswebhook.TargetSeed,
		Mode:     extensionswebhook.ModeMutating,
		Path:     path,
		Webhook:  &admission.Webhook{Handler: handler},
		Selector: namespaceSelector,
	}, nil
}

// buildSelector creates and returns a LabelSelector for the given webhook kind and provider.
func buildSelector(networkProvider, cloudProvider string) (*metav1.LabelSelector, error) {
	// Create and return LabelSelector
	return &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: v1alpha1constants.LabelShootProvider, Operator: metav1.LabelSelectorOpIn, Values: []string{cloudProvider}},
			{Key: v1alpha1constants.LabelNetworkingProvider, Operator: metav1.LabelSelectorOpIn, Values: []string{networkProvider}},
		},
	}, nil
}
