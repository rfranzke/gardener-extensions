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

package cmd

import (
	"context"
	"fmt"

	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// CertDirFlag is the name of the command line flag to specify the directory that contains the webhook server key and certificate.
	CertDirFlag = "webhook-config-cert-dir"
	// ModeFlag is the name of the command line flag to specify the webhook config mode.
	ModeFlag = "webhook-config-mode"
	// URLFlag is the name of the command line flag to specify the URL that is used to register the webhooks in Kubernetes.
	URLFlag = "webhook-config-url"
	// NamespaceFlag is the name of the command line flag to specify the webhook config namespace for 'service' mode.
	NamespaceFlag = "webhook-config-namespace"
)

// ServerOptions are command line options that can be set for ServerConfig.
type ServerOptions struct {
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string
	// Mode is the URl that is used to register the webhooks in Kubernetes.
	Mode string
	// URL is the URl that is used to register the webhooks in Kubernetes.
	URL string
	// Namespace is the webhook config namespace for 'service' mode.
	Namespace string

	config *ServerConfig
}

// ServerConfig is a completed webhook server configuration.
type ServerConfig struct {
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string
	// Mode is the URl that is used to register the webhooks in Kubernetes.
	Mode string
	// URL is the URl that is used to register the webhooks in Kubernetes.
	URL string
	// Namespace is the webhook config namespace for 'service' mode.
	Namespace string
}

// Complete implements Completer.Complete.
func (w *ServerOptions) Complete() error {
	w.config = &ServerConfig{
		CertDir:   w.CertDir,
		Mode:      w.Mode,
		URL:       w.URL,
		Namespace: w.Namespace,
	}

	if len(w.Mode) == 0 {
		w.config.Mode = extensionswebhook.ModeService
	}

	return nil
}

// Completed returns the completed ServerConfig. Only call this if `Complete` was successful.
func (w *ServerOptions) Completed() *ServerConfig {
	return w.config
}

// AddFlags implements Flagger.AddFlags.
func (w *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&w.CertDir, CertDirFlag, w.CertDir, "The directory that contains the webhook server key and certificate.")
	fs.StringVar(&w.Mode, ModeFlag, w.Mode, "The webhook mode - either 'url' (when running outside the cluster) or 'service' (when running inside the cluster).")
	fs.StringVar(&w.URL, URLFlag, w.URL, "The directory that contains the webhook URL when running outside of the cluster it is serving.")
	fs.StringVar(&w.Namespace, NamespaceFlag, w.Namespace, "The webhook config namespace for 'service' mode.")
}

// DisableFlag is the name of the command line flag to disable individual webhooks.
const DisableFlag = "disable-webhooks"

// NameToFactory binds a specific name to a webhook's factory function.
type NameToFactory struct {
	Name string
	Func func(manager.Manager) (*extensionswebhook.Webhook, error)
}

// SwitchOptions are options to build an AddToManager function that filters the disabled webhooks.
type SwitchOptions struct {
	Disabled []string

	nameToWebhookFactory     map[string]func(manager.Manager) (*extensionswebhook.Webhook, error)
	webhookFactoryAggregator extensionswebhook.FactoryAggregator
}

// Register registers the given NameToWebhookFuncs in the options.
func (w *SwitchOptions) Register(pairs ...NameToFactory) {
	for _, pair := range pairs {
		w.nameToWebhookFactory[pair.Name] = pair.Func
	}
}

// AddFlags implements Option.
func (w *SwitchOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&w.Disabled, DisableFlag, w.Disabled, "List of webhooks to disable")
}

// Complete implements Option.
func (w *SwitchOptions) Complete() error {
	disabled := sets.NewString()
	for _, disabledName := range w.Disabled {
		if _, ok := w.nameToWebhookFactory[disabledName]; !ok {
			return fmt.Errorf("cannot disable unknown webhook %q", disabledName)
		}
		disabled.Insert(disabledName)
	}

	for name, addToManager := range w.nameToWebhookFactory {
		if !disabled.Has(name) {
			w.webhookFactoryAggregator.Register(addToManager)
		}
	}
	return nil
}

// Completed returns the completed SwitchConfig. Call this only after successfully calling `Completed`.
func (w *SwitchOptions) Completed() *SwitchConfig {
	return &SwitchConfig{WebhooksFactory: w.webhookFactoryAggregator.Webhooks}
}

// SwitchConfig is the completed configuration of SwitchOptions.
type SwitchConfig struct {
	WebhooksFactory func(manager.Manager) ([]*extensionswebhook.Webhook, error)
}

// Switch binds the given name to the given AddToManager function.
func Switch(name string, f func(manager.Manager) (*extensionswebhook.Webhook, error)) NameToFactory {
	return NameToFactory{
		Name: name,
		Func: f,
	}
}

// NewSwitchOptions creates new SwitchOptions with the given initial pairs.
func NewSwitchOptions(pairs ...NameToFactory) *SwitchOptions {
	opts := SwitchOptions{nameToWebhookFactory: map[string]func(manager.Manager) (*extensionswebhook.Webhook, error){}, webhookFactoryAggregator: extensionswebhook.FactoryAggregator{}}
	opts.Register(pairs...)
	return &opts
}

// AddToManagerOptions are options to create an `AddToManager` function from ServerOptions and SwitchOptions.
type AddToManagerOptions struct {
	serverName string
	Server     ServerOptions
	Switch     SwitchOptions
}

// NewAddToManagerOptions creates new AddToManagerOptions with the given server name, server, and switch options.
func NewAddToManagerOptions(serverName string, serverOpts *ServerOptions, switchOpts *SwitchOptions) *AddToManagerOptions {
	return &AddToManagerOptions{
		serverName: serverName,
		Server:     *serverOpts,
		Switch:     *switchOpts,
	}
}

// AddFlags implements Option.
func (c *AddToManagerOptions) AddFlags(fs *pflag.FlagSet) {
	c.Switch.AddFlags(fs)
	c.Server.AddFlags(fs)
}

// Complete implements Option.
func (c *AddToManagerOptions) Complete() error {
	if err := c.Switch.Complete(); err != nil {
		return err
	}

	return c.Server.Complete()
}

// Compoleted returns the completed AddToManagerConfig. Only call this if a previous call to `Complete` succeeded.
func (c *AddToManagerOptions) Completed() *AddToManagerConfig {
	return &AddToManagerConfig{
		serverName: c.serverName,
		Server:     *c.Server.Completed(),
		Switch:     *c.Switch.Completed(),
	}
}

// AddToManagerConfig is a completed AddToManager configuration.
type AddToManagerConfig struct {
	serverName string
	Server     ServerConfig
	Switch     SwitchConfig
}

// AddToManager instantiates all webhooks of this configuration. If there are any webhooks, it creates a
// webhook server, registers the webhooks and adds the server to the manager. Otherwise, it is a no-op.
func (c *AddToManagerConfig) AddToManager(mgr manager.Manager) (
	[]admissionregistrationv1beta1.Webhook,
	[]admissionregistrationv1beta1.Webhook,
	[]admissionregistrationv1beta1.Webhook,
	[]admissionregistrationv1beta1.Webhook,
	error,
) {

	ctx := context.Background()

	webhooks, err := c.Switch.WebhooksFactory(mgr)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "could not create webhooks")
	}

	webhookServer := mgr.GetWebhookServer()
	webhookServer.CertDir = c.Server.CertDir

	for _, wh := range webhooks {
		if wh.Handler != nil {
			webhookServer.Register("/"+wh.Name, wh.Handler)
		} else {
			webhookServer.Register("/"+wh.Name, wh.Webhook)
		}
	}

	caBundle, err := extensionswebhook.GenerateCertificates(ctx, mgr, c.Server.CertDir, c.Server.Namespace, c.serverName, c.Server.Mode, c.Server.URL)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "could not generate certificates")
	}

	mutatingSeedWebhooks, mutatingShootWebhooks, validatingSeedWebhooks, validatingShootWebhooks, err := extensionswebhook.RegisterWebhooks(ctx, mgr, c.Server.Namespace, c.serverName, webhookServer.Port, c.Server.Mode, c.Server.URL, caBundle, webhooks)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "could not create webhooks")
	}

	return mutatingSeedWebhooks, mutatingShootWebhooks, validatingSeedWebhooks, validatingShootWebhooks, nil
}
