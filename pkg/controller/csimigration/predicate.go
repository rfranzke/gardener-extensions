// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package csimigration

import (
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ClusterShootProviderType is a predicate for the provider type of the shoot in the cluster resource.
func ClusterShootProviderType(providerType string) predicate.Predicate {
	f := func(obj runtime.Object) bool {
		if obj == nil {
			return false
		}

		cluster, ok := obj.(*extensionsv1alpha1.Cluster)
		if !ok {
			return false
		}

		decoder, err := extensionscontroller.NewGardenDecoder()
		if err != nil {
			return false
		}

		shoot, err := extensionscontroller.ShootFromCluster(decoder, cluster)
		if err != nil {
			return false
		}

		return shoot.Spec.Provider.Type == providerType
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return f(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return f(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return f(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return f(event.Object)
		},
	}
}

// ClusterShootKubernetesVersionAtLeast is a predicate for the kubernetes version of the shoot in the cluster resource.
func ClusterShootKubernetesVersionAtLeast(kubernetesVersion string) predicate.Predicate {
	f := func(obj runtime.Object) bool {
		if obj == nil {
			return false
		}

		cluster, ok := obj.(*extensionsv1alpha1.Cluster)
		if !ok {
			return false
		}

		decoder, err := extensionscontroller.NewGardenDecoder()
		if err != nil {
			return false
		}

		shoot, err := extensionscontroller.ShootFromCluster(decoder, cluster)
		if err != nil {
			return false
		}

		constraint, err := version.CompareVersions(shoot.Spec.Kubernetes.Version, ">=", kubernetesVersion)
		if err != nil {
			return false
		}

		return constraint
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return f(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return f(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return f(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return f(event.Object)
		},
	}
}

// ClusterCSIMigrationControllerFinished is a predicate for an annotation on the cluster.
func ClusterCSIMigrationControllerFinished() predicate.Predicate {
	f := func(obj runtime.Object) bool {
		if obj == nil {
			return false
		}

		cluster, ok := obj.(*extensionsv1alpha1.Cluster)
		if !ok {
			return false
		}

		return !metav1.HasAnnotation(cluster.ObjectMeta, AnnotationKeyControllerFinished)
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return f(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return f(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return f(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return f(event.Object)
		},
	}
}
