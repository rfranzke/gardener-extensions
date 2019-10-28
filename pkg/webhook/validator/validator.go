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

package validator

import (
	"fmt"

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
	WebhookName = "validator"
)

var logger = log.Log.WithName("validator-webhook")

// AddArgs are arguments for adding a validator webhook to a manager.
type AddArgs struct {
	// Kind is the kind of this webhook.
	Kind string
	// Provider is the provider of this webhook.
	Provider string
	// Types is a list of resource types.
	Types []runtime.Object
	// Validator is a validator to be used by the admission handler.
	Validator extensionswebhook.Validator
}

// Add creates a new validator webhook and adds it to the given Manager.
func Add(mgr manager.Manager, args AddArgs) (*extensionswebhook.Webhook, error) {
	logger := logger.WithValues("validator", args.Kind, "provider", args.Provider)

	// Create handler
	handler, err := extensionswebhook.NewHandler(mgr, args.Types, nil, args.Validator, logger)
	if err != nil {
		return nil, err
	}

	// Build namespace selector from the webhook kind and provider
	namespaceSelector, err := buildSelector(args.Kind, args.Provider)
	if err != nil {
		return nil, err
	}

	// Create webhook
	var (
		name = WebhookName
		path = WebhookName
	)

	logger.Info("Creating validator webhook", "name", name)
	return &extensionswebhook.Webhook{
		Name:     name,
		Kind:     args.Kind,
		Provider: args.Provider,
		Types:    args.Types,
		Target:   extensionswebhook.TargetSeed,
		Mode:     extensionswebhook.ModeValidating,
		Path:     path,
		Webhook:  &admission.Webhook{Handler: handler},
		Selector: namespaceSelector,
	}, nil
}

// buildSelector creates and returns a LabelSelector for the given webhook kind and provider.
func buildSelector(kind, provider string) (*metav1.LabelSelector, error) {
	// Determine label selector key from the kind
	var key string
	switch kind {
	case extensionswebhook.KindSeed:
		key = v1alpha1constants.LabelSeedProvider
	case extensionswebhook.KindShoot:
		key = v1alpha1constants.LabelShootProvider
	case extensionswebhook.KindBackup:
		key = v1alpha1constants.LabelBackupProvider
	default:
		return nil, fmt.Errorf("invalid webhook kind '%s'", kind)
	}

	// Create and return LabelSelector
	return &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: key, Operator: metav1.LabelSelectorOpIn, Values: []string{provider}},
		},
	}, nil
}
