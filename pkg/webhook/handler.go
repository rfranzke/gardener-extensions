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

package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NewHandler creates a new handler for the given types, using the given mutator, and logger.
func NewHandler(mgr manager.Manager, types []runtime.Object, mutator Mutator, validator Validator, logger logr.Logger) (*handler, error) {
	// Build a map of the given types keyed by their GVKs
	typesMap, err := buildTypesMap(mgr, types)
	if err != nil {
		return nil, err
	}

	// Create and return a handler
	return &handler{
		logger:    logger.WithName("handler"),
		typesMap:  typesMap,
		mutator:   mutator,
		validator: validator,
	}, nil
}

type handler struct {
	logger    logr.Logger
	decoder   *admission.Decoder
	typesMap  map[metav1.GroupVersionKind]runtime.Object
	mutator   Mutator
	validator Validator
}

// InjectDecoder injects the given decoder into the handler.
func (h *handler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the given client into the mutator.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (h *handler) InjectClient(client client.Client) error {
	if h.mutator != nil {
		if _, err := inject.ClientInto(client, h.mutator); err != nil {
			return errors.Wrap(err, "could not inject the client into the mutator")
		}
	}
	if h.validator != nil {
		if _, err := inject.ClientInto(client, h.validator); err != nil {
			return errors.Wrap(err, "could not inject the client into the validator")
		}
	}
	return nil
}

// InjectScheme injects the given scheme into the mutator.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (h *handler) InjectScheme(scheme *runtime.Scheme) error {
	if h.mutator != nil {
		if _, err := inject.SchemeInto(scheme, h.mutator); err != nil {
			return errors.Wrap(err, "could not inject the scheme into the mutator")
		}
	}
	if h.validator != nil {
		if _, err := inject.SchemeInto(scheme, h.validator); err != nil {
			return errors.Wrap(err, "could not inject the scheme into the validator")
		}
	}
	return nil
}

type handleFunc func(context.Context, runtime.Object, runtime.Object, *http.Request) error

// Handle handles the given admission request.
func (h *handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	var (
		f    handleFunc
		mode string
	)

	switch {
	case h.mutator != nil:
		mode = ModeMutating
		f = func(ctx context.Context, oldObj, newObj runtime.Object, r *http.Request) error {
			return h.mutator.Mutate(ctx, newObj)
		}
	case h.validator != nil:
		mode = ModeValidating

		f = func(ctx context.Context, oldObj, newObj runtime.Object, r *http.Request) error {
			return h.validator.Validate(ctx, oldObj, newObj)
		}
	default:
		panic("neither mutator nor validator is set")
	}

	return handle(ctx, req, nil, mode, f, h.typesMap, h.decoder, h.logger)
}

func handle(ctx context.Context, req admission.Request, r *http.Request, mode string, f handleFunc, typesMap map[metav1.GroupVersionKind]runtime.Object, decoder *admission.Decoder, logger logr.Logger) admission.Response {
	ar := req.AdmissionRequest

	// Decode object
	t, ok := typesMap[ar.Kind]
	if !ok {
		return admission.Errored(http.StatusBadRequest, errors.Errorf("unexpected request kind %s", ar.Kind.String()))
	}

	var (
		objOld runtime.Object
		oldObj runtime.Object

		objNew = t.DeepCopyObject()
		newObj runtime.Object
	)

	if req.OldObject.Raw != nil {
		objOld = t.DeepCopyObject()
		if err := decoder.DecodeRaw(req.OldObject, objOld); err != nil {
			return admission.Errored(http.StatusBadRequest, errors.Wrapf(err, "could not decode old obj %v", ar))
		}
		oldObj = objOld.DeepCopyObject()
	}

	if err := decoder.DecodeRaw(req.Object, objNew); err != nil {
		return admission.Errored(http.StatusBadRequest, errors.Wrapf(err, "could not decode new obj %v", ar))
	}
	newObj = objNew.DeepCopyObject()

	// Get object accessor
	accessor, err := meta.Accessor(objNew)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, errors.Wrapf(err, "could not get accessor for %v", objNew))
	}

	// Handle the resource
	if err := f(ctx, oldObj, newObj, r); err != nil {
		return admission.Errored(http.StatusBadRequest, errors.Wrapf(err, "could not %s %s %s/%s", mode, ar.Kind.Kind, accessor.GetNamespace(), accessor.GetName()))
	}

	if mode == ModeMutating {
		// Return a patch response if the resource should be changed
		if !equality.Semantic.DeepEqual(objNew, newObj) {
			oldObjMarshaled, err := json.Marshal(objNew)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
			newObjMarshaled, err := json.Marshal(newObj)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
			return admission.PatchResponseFromRaw(oldObjMarshaled, newObjMarshaled)
		}
	}

	// Return a validation response if the resource should not be changed
	return admission.ValidationResponse(true, "")
}
