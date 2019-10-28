// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain m copy of the License at
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
	"context"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/webhook/validator/genericvalidator"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// NewValidator creates a new controlplaneexposure validator.
func NewValidator(logger logr.Logger) genericvalidator.Validator {
	return &validator{
		logger: logger.WithName("aws-validator"),
	}
}

type validator struct {
	logger  logr.Logger
	decoder runtime.Decoder
}

// InjectScheme injects the given scheme into the reconciler.
func (v *validator) InjectScheme(scheme *runtime.Scheme) error {
	v.decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	return nil
}

// Validate validates CREATE requests.
func (v *validator) Validate(ctx context.Context, cluster *extensionscontroller.Cluster, obj runtime.Object) field.ErrorList {
	switch x := obj.(type) {
	case *extensionsv1alpha1.Infrastructure:
		return v.ValidateInfrastructure(ctx, cluster, x)
	case *extensionsv1alpha1.Worker:
		return v.ValidateWorker(ctx, cluster, x)
	}

	return nil
}

// Validate validates UPDATE requests.
func (v *validator) ValidateUpdate(ctx context.Context, cluster *extensionscontroller.Cluster, oldObj, newObj runtime.Object) field.ErrorList {
	switch x := newObj.(type) {
	case *extensionsv1alpha1.Infrastructure:
		y := oldObj.(*extensionsv1alpha1.Infrastructure)
		return v.ValidateInfrastructureUpdate(ctx, cluster, y, x)
	case *extensionsv1alpha1.Worker:
		y := oldObj.(*extensionsv1alpha1.Worker)
		return v.ValidateWorkerUpdate(ctx, cluster, y, x)
	}

	return nil
}
