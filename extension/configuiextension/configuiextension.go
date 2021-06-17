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
	"reflect"
	"strings"

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

	for n, r := range hc.colCfg.Receivers {
		hc.renderReceiver(w, n, r)
	}
}

func (hc *configUIExtension) renderReceiver(w http.ResponseWriter, name config2.ComponentID, cfg config2.Receiver) {
	//w.Write([]byte(fmt.Sprintf("Receiver %s<br>", name.String())))

	//r := reflect.TypeOf(cfg)
	//if r.Kind() != reflect.Struct {
	//	return
	//}

	html := renderStruct(name.String(), reflect.ValueOf(cfg), 1)
	w.Write([]byte(html))
}

func renderStruct(structName string, v reflect.Value, level int) string {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	shtml := fmt.Sprintf("<h%d>%s</h%d>", level, structName, level)

	shtml = shtml + `<table border=1 style="border-collapse: collapse">`

	shtml = shtml + renderStructFields(v, level)

	shtml = shtml + "</table>"

	return shtml
}

func renderStructFields(v reflect.Value, level int) string {

	shtml := ""

	typeOfT := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := typeOfT.Field(i)

		if len(fieldType.PkgPath) != 0 {
			// skip unexported fields
			continue
		}

		field := v.Field(i)
		fieldName := fieldType.Name

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			field = field.Elem()
		}

		var html string
		kind := field.Kind()
		switch kind {
		case reflect.Bool:
			html = renderBool(fieldName, field)
		case reflect.Int:
			html = renderInt(fieldName, field)
		case reflect.String:
			html = renderString(fieldName, field)
		case reflect.Struct:
			if isSquashed(fieldType) {
				shtml = shtml + renderStructFields(field, level)
				continue
			} else {
				html = renderStruct(fieldName, field, level+1)
			}
		default:
			html = fmt.Sprintf("%d: %s %s = %v<br>", i, fieldName, field.Type(), field.Interface())
		}

		row := fmt.Sprintf(`<tr><td><label for="%s">%s</label>:</td><td>%s</td></tr>`,
			fieldName, fieldName, html)

		shtml = shtml + row
	}

	return shtml
}

func isSquashed(f reflect.StructField) bool {
	tagValue := f.Tag.Get("mapstructure")
	if tagValue == "" {

		// Ignore special types.
		switch f.Type.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer:
			// Allow the config to carry the types above, but since they are not read
			// when loading configuration, just ignore them.
			return false
		}

		// Public fields of other types should be tagged.
		chars := []byte(f.Name)
		if len(chars) > 0 && chars[0] >= 'A' && chars[0] <= 'Z' {
			return false
		}

		// Not public field, no need to have a tag.
		return false
	}

	tagParts := strings.Split(tagValue, ",")
	if tagParts[0] != "" {
		if tagParts[0] == "-" {
			// Nothing to do, as mapstructure decode skips this field.
			return false
		}
	}

	// Check if squash is specified.
	squash := false
	for _, tag := range tagParts[1:] {
		if tag == "squash" {
			squash = true
			break
		}
	}

	//if squash {
	//	// Field was squashed.
	//	if (f.Type.Kind() != reflect.Struct) && (f.Type.Kind() != reflect.Ptr || f.Type.Elem().Kind() != reflect.Struct) {
	//		return fmt.Errorf(
	//			"attempt to squash non-struct type on field %q", f.Name)
	//	}
	//}
	return squash
}

func renderBool(fieldName string, v reflect.Value) string {
	var checked string
	if v.Interface().(bool) {
		checked = "checked"
	}
	return fmt.Sprintf(
		`<input type="checkbox" id="%s" name="%s" value="" %s>`,
		fieldName, fieldName, checked)
}

func renderInt(fieldName string, v reflect.Value) string {
	val := v.Interface().(int)
	return fmt.Sprintf(
		`<input type="number" id="%s" name="%s" value="%d">`,
		fieldName, fieldName, val)
}

func renderString(fieldName string, v reflect.Value) string {
	val := v.Interface().(string)
	return fmt.Sprintf(
		`<input type="text" id="%s" name="%s" value="%s">`,
		fieldName, fieldName, val)
}
