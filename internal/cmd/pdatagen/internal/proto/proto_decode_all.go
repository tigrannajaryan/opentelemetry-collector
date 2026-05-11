// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package proto // import "go.opentelemetry.io/collector/internal/cmd/pdatagen/internal/proto"

import "go.opentelemetry.io/collector/internal/cmd/pdatagen/internal/tmplutil"

// DecodeAll is generated separately from the lazy getter path because some
// internal callers, especially migration and tests, need ordinary fully
// materialized structs. Only generated message fields participate; custom ID
// fields already decode eagerly through their hand-written protobuf support.
const decodeAllProtoMessage = `{{ if .lazyMessage -}}
{{- if .repeated -}}
	for i := range orig.{{ .fieldName }} {
		{{ if .nullable -}}
		if orig.{{ .fieldName }}[i] != nil {
			orig.{{ .fieldName }}[i].DecodeAll()
		}
		{{- else -}}
		orig.{{ .fieldName }}[i].DecodeAll()
		{{- end }}
	}
{{- else if ne .oneOfGroup "" -}}
	if ov, ok := orig.{{ .oneOfGroup }}.(*{{ .oneOfMessageName }}); ok && ov.{{ .fieldName }} != nil {
		ov.{{ .fieldName }}.DecodeAll()
	}
{{- else if .nullable -}}
	if orig.{{ .fieldName }} != nil {
		orig.{{ .fieldName }}.DecodeAll()
	}
{{- else -}}
	orig.{{ .fieldName }}.DecodeAll()
{{- end }}
{{- end }}`

func (pf *Field) GenDecodeAllProto() string {
	if pf.Type != TypeMessage {
		return ""
	}
	return tmplutil.Execute(tmplutil.Parse("decodeAllProtoMessage", []byte(decodeAllProtoMessage)), pf.getTemplateFields())
}
