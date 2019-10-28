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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInfrastructure validates CREATE requests for infrastructures.
func (v *validator) ValidateInfrastructure(ctx context.Context, cluster *extensionscontroller.Cluster, infrastructure *extensionsv1alpha1.Infrastructure) field.ErrorList {
	infrastructureConfig, errs := v.infrastructureConfigFrom(infrastructure, cluster)
	if errs != nil {
		return errs
	}

	return validation.ValidateInfrastructureConfig(infrastructureConfig, cluster.CoreShoot.Spec.Networking.Nodes, cluster.CoreShoot.Spec.Networking.Pods, cluster.CoreShoot.Spec.Networking.Services)
}

// ValidateInfrastructureUpdate validates UPDATE requests for infrastructures.
func (v *validator) ValidateInfrastructureUpdate(ctx context.Context, cluster *extensionscontroller.Cluster, oldInfra, newInfra *extensionsv1alpha1.Infrastructure) field.ErrorList {
	oldInfrastructureConfig, errs := v.infrastructureConfigFrom(oldInfra, cluster)
	if errs != nil {
		return errs
	}
	newInfrastructureConfig, errs := v.infrastructureConfigFrom(newInfra, cluster)
	if errs != nil {
		return errs
	}

	return validation.ValidateInfrastructureConfigUpdate(oldInfrastructureConfig, newInfrastructureConfig, cluster.CoreShoot.Spec.Networking.Nodes, cluster.CoreShoot.Spec.Networking.Pods, cluster.CoreShoot.Spec.Networking.Services)
}

func (v *validator) infrastructureConfigFrom(infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) (*awsapi.InfrastructureConfig, field.ErrorList) {
	infrastructureConfig := &awsapi.InfrastructureConfig{}
	if infra.Spec.ProviderConfig == nil || infra.Spec.ProviderConfig.Raw == nil {
		return nil, field.ErrorList{field.Invalid(nil, infra.Spec.ProviderConfig, "no infrastructureConfig provided")}
	}
	if _, _, err := v.decoder.Decode(infra.Spec.ProviderConfig.Raw, nil, infrastructureConfig); err != nil {
		return nil, field.ErrorList{field.InternalError(nil, fmt.Errorf("could not decode provider config: %+v", err))}
	}

	return infrastructureConfig, nil
}
