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
	"net/http"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	config2 "go.opentelemetry.io/collector/config"
)

type configUIExtension struct {
	config Config
	logger *zap.Logger
	server http.Server
	stopCh chan struct{}
	colCfg *config2.Config
}

func (cfe *configUIExtension) Start(_ context.Context, host component.Host) error {
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	//})

	//go func() {
	//	if err := http.ListenAndServe(cfe.config.TCPAddr, nil); err != nil {
	//
	//	}
	//}()

	cfe.colCfg = host.GetConfig()

	ln, err := cfe.config.TCPAddr.Listen()
	if err != nil {
		return err
	}

	cfe.logger.Info("Starting config UI extension", zap.Any("config", cfe.config))
	mux := http.NewServeMux()
	mux.HandleFunc("/", cfe.Handler)

	cfe.server = http.Server{Handler: mux}

	cfe.stopCh = make(chan struct{})
	go func() {
		defer close(cfe.stopCh)

		if err := cfe.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			host.ReportFatalError(err)
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

// Handler creates a new HTTP handler.
func (hc *configUIExtension) Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("hello world"))

	for n, _ := range hc.colCfg.Pipelines {
		w.Write([]byte(fmt.Sprintf("Pipeline %s", n)))

	}
}
