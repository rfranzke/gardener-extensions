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

package csimigration_test

import (
	"encoding/json"

	. "github.com/gardener/gardener-extensions/pkg/controller/csimigration"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("predicate", func() {
	const (
		extensionType = "extension-type"
		version       = "1.18"
	)

	Describe("#ClusterShootProviderType", func() {
		It("should match the type", func() {
			var (
				predicate                                           = ClusterShootProviderType(extensionType)
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents(extensionType, version, false)
			)

			Expect(predicate.Create(createEvent)).To(BeTrue())
			Expect(predicate.Update(updateEvent)).To(BeTrue())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeTrue())
		})

		It("should not match the type", func() {
			var (
				predicate                                           = ClusterShootProviderType(extensionType)
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents("other-extension-type", version, false)
			)

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeFalse())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})

	Describe("#ClusterShootKubernetesVersionAtLeast", func() {
		It("should match the minimum kubernetes version", func() {
			var (
				predicate                                           = ClusterShootKubernetesVersionAtLeast(version)
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents(extensionType, version, false)
			)

			Expect(predicate.Create(createEvent)).To(BeTrue())
			Expect(predicate.Update(updateEvent)).To(BeTrue())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeTrue())
		})

		It("should not match the minimum kubernetes version", func() {
			var (
				predicate                                           = ClusterShootKubernetesVersionAtLeast(version)
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents(extensionType, "1.17", false)
			)

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeFalse())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})

	Describe("#ClusterCSIMigrationControllerFinished", func() {
		It("should return true because controller not finished", func() {
			var (
				predicate                                           = ClusterCSIMigrationControllerFinished()
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents(extensionType, version, false)
			)

			Expect(predicate.Create(createEvent)).To(BeTrue())
			Expect(predicate.Update(updateEvent)).To(BeTrue())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeTrue())
		})

		It("should return false because controller is finished", func() {
			var (
				predicate                                           = ClusterCSIMigrationControllerFinished()
				createEvent, updateEvent, deleteEvent, genericEvent = computeEvents(extensionType, version, true)
			)

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeFalse())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})
})

func computeEvents(extensionType, kubernetesVersion string, controllerFinished bool) (event.CreateEvent, event.UpdateEvent, event.DeleteEvent, event.GenericEvent) {
	shoot := &gardencorev1beta1.Shoot{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gardencorev1beta1.SchemeGroupVersion.String(),
			Kind:       "Shoot",
		},
		Spec: gardencorev1beta1.ShootSpec{
			Provider: gardencorev1beta1.Provider{
				Type: extensionType,
			},
			Kubernetes: gardencorev1beta1.Kubernetes{
				Version: kubernetesVersion,
			},
		},
	}

	shootJSON, err := json.Marshal(shoot)
	Expect(err).To(Succeed())

	cluster := &extensionsv1alpha1.Cluster{
		Spec: extensionsv1alpha1.ClusterSpec{
			Shoot: runtime.RawExtension{Raw: shootJSON},
		},
	}

	if controllerFinished {
		cluster.ObjectMeta.Annotations = map[string]string{AnnotationKeyControllerFinished: "true"}
	}

	return event.CreateEvent{Object: cluster},
		event.UpdateEvent{ObjectOld: cluster, ObjectNew: cluster},
		event.DeleteEvent{Object: cluster},
		event.GenericEvent{Object: cluster}
}
