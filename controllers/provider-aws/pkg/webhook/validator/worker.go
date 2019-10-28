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
	"fmt"

	awsapi "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws/validation"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateWorker validates CREATE requests for workers.
func (v *validator) ValidateWorker(ctx context.Context, cluster *extensionscontroller.Cluster, obj runtime.Object) field.ErrorList {
	worker, ok := obj.(*extensionsv1alpha1.Worker)
	if !ok {
		return field.ErrorList{field.InternalError(nil, fmt.Errorf("expected extensionsv1alpha1.Worker but got something else"))}
	}

	var allErrs field.ErrorList

	for i, pool := range worker.Spec.Pools {
		if pool.ProviderConfig == nil || pool.ProviderConfig.Raw == nil {
			continue
		}

		volumePath := field.NewPath("pools").Index(i).Child("volume")
		if pool.Volume == nil {
			allErrs = append(allErrs, field.Required(volumePath, "volume is required"))
			return allErrs
		}
		if pool.Volume.Type == nil {
			allErrs = append(allErrs, field.Required(volumePath.Child("type"), "volume type is required"))
			return allErrs
		}

		workerConfig := &awsapi.WorkerConfig{}
		if _, _, err := v.decoder.Decode(pool.ProviderConfig.Raw, nil, workerConfig); err != nil {
			return field.ErrorList{field.InternalError(nil, fmt.Errorf("could not decode provider config: %+v", err))}
		}

		allErrs = append(allErrs, validation.ValidateWorkerConfig(workerConfig, pool.Volume.Type)...)
	}

	return allErrs
}

// ValidateWorker validates UPDATE requests for workers.
func (v *validator) ValidateWorkerUpdate(ctx context.Context, cluster *extensionscontroller.Cluster, oldObj, newObj runtime.Object) field.ErrorList {
	return v.Validate(ctx, cluster, newObj)
}
