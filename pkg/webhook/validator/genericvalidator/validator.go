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

package genericvalidator

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	extensionsvalidation "github.com/gardener/gardener/pkg/apis/extensions/validation"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// Validator defines validations functions with cluster objects.
type Validator interface {
	// Validate validates CREATE requests for objects.
	Validate(ctx context.Context, cluster *extensionscontroller.Cluster, obj runtime.Object) field.ErrorList
	// ValidateUpdate validates UPDATE requests for objects.
	ValidateUpdate(ctx context.Context, cluster *extensionscontroller.Cluster, oldObj, newObj runtime.Object) field.ErrorList
}

// NewValidator creates a new validator
func NewValidator(logger logr.Logger, v Validator) extensionswebhook.Validator {
	return &validator{
		logger:    logger.WithName("validator"),
		validator: v,
	}
}

type validator struct {
	logger    logr.Logger
	client    client.Client
	validator Validator
}

// InjectClient injects the given client into the validator.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (v *validator) InjectClient(client client.Client) error {
	v.client = client
	if _, err := inject.ClientInto(client, v.validator); err != nil {
		return errors.Wrap(err, "could not inject the client into the validator")
	}
	return nil
}

// InjectScheme injects the given scheme into the validator.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (v *validator) InjectScheme(scheme *runtime.Scheme) error {
	if _, err := inject.SchemeInto(scheme, v.validator); err != nil {
		return errors.Wrap(err, "could not inject the scheme into the validator")
	}
	return nil
}

// Validate validates objects.
func (v *validator) Validate(ctx context.Context, oldObj, newObj runtime.Object) error {
	acc, err := meta.Accessor(newObj)
	if err != nil {
		return errors.Wrapf(err, "could not create accessor during webhook")
	}

	cluster, err := extensionscontroller.GetCluster(ctx, v.client, acc.GetNamespace())
	if err != nil {
		return errors.Wrapf(err, "could not get cluster for namespace '%s'", acc.GetNamespace())
	}

	var allErrs field.ErrorList

	switch x := newObj.(type) {
	case *extensionsv1alpha1.BackupBucket:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateBackupBucket(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.BackupBucket); ok {
			allErrs = extensionsvalidation.ValidateBackupBucketUpdate(x, y)
		}
	case *extensionsv1alpha1.BackupEntry:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateBackupEntry(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.BackupEntry); ok {
			allErrs = extensionsvalidation.ValidateBackupEntryUpdate(x, y)
		}
	case *extensionsv1alpha1.ControlPlane:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateControlPlane(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.ControlPlane); ok {
			allErrs = extensionsvalidation.ValidateControlPlaneUpdate(x, y)
		}
	case *extensionsv1alpha1.Extension:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateExtension(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.Extension); ok {
			allErrs = extensionsvalidation.ValidateExtensionUpdate(x, y)
		}
	case *extensionsv1alpha1.Infrastructure:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateInfrastructure(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.Infrastructure); ok {
			allErrs = extensionsvalidation.ValidateInfrastructureUpdate(x, y)
		}
	case *extensionsv1alpha1.Network:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateNetwork(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.Network); ok {
			allErrs = extensionsvalidation.ValidateNetworkUpdate(x, y)
		}
	case *extensionsv1alpha1.OperatingSystemConfig:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateOperatingSystemConfig(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.OperatingSystemConfig); ok {
			allErrs = extensionsvalidation.ValidateOperatingSystemConfigUpdate(x, y)
		}
	case *extensionsv1alpha1.Worker:
		if oldObj == nil {
			allErrs = extensionsvalidation.ValidateWorker(x)
		} else if y, ok := oldObj.(*extensionsv1alpha1.Worker); ok {
			allErrs = extensionsvalidation.ValidateWorkerUpdate(x, y)
		}
	}

	if oldObj == nil {
		allErrs = append(allErrs, v.validator.Validate(ctx, cluster, newObj)...)
	} else {
		allErrs = append(allErrs, v.validator.ValidateUpdate(ctx, cluster, oldObj, newObj)...)
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("%+v", allErrs)
	}
	return nil
}
