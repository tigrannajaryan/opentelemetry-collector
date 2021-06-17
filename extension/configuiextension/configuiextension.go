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

package configuiextension

import (
	"context"
	"fmt"
	"html"
	"net/http"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
)

type configUIExtension struct {
	config Config
	logger *zap.Logger
	server http.Server
	stopCh chan struct{}
}

func (cfe *configUIExtension) Start(_ context.Context, host component.Host) error {
	http.Handle("/foo", fooHandler)

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	go func() {
		if err := http.ListenAndServe(cfe.config.TCPAddr, nil); err != nil {

		}
	}()

	return nil
}

func (cfe *configUIExtension) Shutdown(context.Context) error {
	err := cfe.server.Close()
	if cfe.stopCh != nil {
		<-cfe.stopCh
	}
	return err
}

func newServer(config Config, logger *zap.Logger) *configUIExtension {
	return &configUIExtension{
		config: config,
		logger: logger,
	}
}
