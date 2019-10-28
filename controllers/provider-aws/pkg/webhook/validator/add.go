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
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	validatorwebhook "github.com/gardener/gardener-extensions/pkg/webhook/validator"
	"github.com/gardener/gardener-extensions/pkg/webhook/validator/genericvalidator"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the AWS validation webhook to the manager.
type AddOptions struct{}

var logger = log.Log.WithName("aws-validation-webhook")

// AddToManagerWithOptions creates a webhook with the given options and adds it to the manager.
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) (*extensionswebhook.Webhook, error) {
	logger.Info("Adding webhook to manager")
	return validatorwebhook.Add(mgr, validatorwebhook.AddArgs{
		Kind:     extensionswebhook.KindShoot,
		Provider: aws.Type,
		Types: []runtime.Object{
			&extensionsv1alpha1.Infrastructure{},
			&extensionsv1alpha1.Worker{},
		},
		Validator: genericvalidator.NewValidator(logger, NewValidator(logger)),
	})
}

// AddToManager creates a webhook with the default options and adds it to the manager.
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}
