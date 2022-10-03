// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service // import "go.opentelemetry.io/collector/service"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/service/internal/configunmarshaler"
)

// ConfigProvider provides the service configuration.
//
// The typical usage is the following:
//
//	cfgProvider.Get(...)
//	cfgProvider.Watch() // wait for an event.
//	cfgProvider.Get(...)
//	cfgProvider.Watch() // wait for an event.
//	// repeat Get/Watch cycle until it is time to shut down the Collector process.
//	cfgProvider.Shutdown()
type ConfigProvider interface {
	// Get returns the service configuration, or error otherwise.
	//
	// Should never be called concurrently with itself, Watch or Shutdown.
	Get(ctx context.Context, factories component.Factories) (*Config, error)

	// Watch blocks until any configuration change was detected or an unrecoverable error
	// happened during monitoring the configuration changes.
	//
	// Error is nil if the configuration is changed and needs to be re-fetched. Any non-nil
	// error indicates that there was a problem with watching the config changes.
	//
	// Should never be called concurrently with itself or Get.
	Watch() <-chan error

	// Shutdown signals that the provider is no longer in use and the that should close
	// and release any resources that it may have created.
	//
	// This function must terminate the Watch channel.
	//
	// Should never be called concurrently with itself or Get.
	Shutdown(ctx context.Context) error

	// ConfigUpdateFailed is called by the Service if applying the last config
	// returned by Get() failed for any reason, including if Get() returned an error.
	ConfigUpdateFailed(err error)

	// ConfigUpdateSucceeded is called by the Service if applying the last config
	// return by Get() succeeded and the Service.Start() succeeded.
	ConfigUpdateSucceeded()
}

type configProvider struct {
	mapResolver *confmap.Resolver

	// lastResolvedConfig contains the result of the last mapResolver.Resolve() operation in YAML format.
	lastResolvedConfig []byte

	// lastKnownGoodConfig is the last config that is known to be good since the service
	// has successfully started and reported it to be good. This field is in YAML format.
	// If as a result of the last mapResolver.Resolve() the service successfully applied
	// the config and started successfully then lastKnownGoodConfig==lastResolvedConfig.
	lastKnownGoodConfig []byte
}

// ConfigProviderSettings are the settings to configure the behavior of the ConfigProvider.
type ConfigProviderSettings struct {
	// ResolverSettings are the settings to configure the behavior of the confmap.Resolver.
	ResolverSettings confmap.ResolverSettings
}

func newDefaultConfigProviderSettings(uris []string) ConfigProviderSettings {
	return ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:       uris,
			Providers:  makeMapProvidersMap(fileprovider.New(), envprovider.New(), yamlprovider.New(), httpprovider.New()),
			Converters: []confmap.Converter{expandconverter.New()},
		},
	}
}

// NewConfigProvider returns a new ConfigProvider that provides the service configuration:
// * Initially it resolves the "configuration map":
//   - Retrieve the confmap.Conf by merging all retrieved maps from the given `locations` in order.
//   - Then applies all the confmap.Converter in the given order.
//
// * Then unmarshalls the confmap.Conf into the service Config.
func NewConfigProvider(set ConfigProviderSettings) (ConfigProvider, error) {
	mr, err := confmap.NewResolver(set.ResolverSettings)
	if err != nil {
		return nil, err
	}

	return &configProvider{
		mapResolver: mr,
	}, nil
}

func (cm *configProvider) Get(ctx context.Context, factories component.Factories) (*Config, error) {
	// If resolving fails we don't want to mistakenly keep the previously resolved config.
	cm.lastResolvedConfig = nil

	retMap, err := cm.mapResolver.Resolve(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve the configuration: %w", err)
	}

	// Update lastResolvedConfig
	effectiveConfig, err := retMap.MarshalToYAML()
	if err == nil {
		cm.lastResolvedConfig = effectiveConfig
	} else {
		// TODO: log the error. We don't want to return an error just because we could not
		// compute the effective config's YAML representation which is only needed for
		// reporting, but not needed for service operation.
	}

	var cfg *Config
	if cfg, err = configunmarshaler.New().Unmarshal(retMap, factories); err != nil {
		return nil, fmt.Errorf("cannot unmarshal the configuration: %w", err)
	}

	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (cm *configProvider) ConfigUpdateFailed(err error) {
	cm.mapResolver.ConfigUpdateFailed(cm.lastKnownGoodConfig, cm.lastResolvedConfig, err)
}

func (cm *configProvider) ConfigUpdateSucceeded() {
	cm.lastKnownGoodConfig = cm.lastResolvedConfig
	cm.mapResolver.ConfigUpdateSucceeded(cm.lastKnownGoodConfig)
}

func (cm *configProvider) Watch() <-chan error {
	return cm.mapResolver.Watch()
}

func (cm *configProvider) Shutdown(ctx context.Context) error {
	return cm.mapResolver.Shutdown(ctx)
}

func makeMapProvidersMap(providers ...confmap.Provider) map[string]confmap.Provider {
	ret := make(map[string]confmap.Provider, len(providers))
	for _, provider := range providers {
		ret[provider.Scheme()] = provider
	}
	return ret
}
