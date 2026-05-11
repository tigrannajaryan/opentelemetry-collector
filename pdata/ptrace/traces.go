// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ptrace // import "go.opentelemetry.io/collector/pdata/ptrace"

// MarkReadOnly marks the Traces as shared so that no further modifications can be done on it.
func (ms Traces) MarkReadOnly() {
	ms.getState().MarkReadOnly()
}

// IsReadOnly returns true if this Traces instance is read-only.
func (ms Traces) IsReadOnly() bool {
	return ms.getState().IsReadOnly()
}

// SpanCount calculates the total number of spans.
func (ms Traces) SpanCount() int {
	spanCount := 0
	orig := ms.getOrig()
	orig.EnsureDecoded()
	for _, rs := range orig.ResourceSpans {
		if rs == nil {
			continue
		}
		rs.EnsureDecoded()
		for _, ss := range rs.ScopeSpans {
			if ss == nil {
				continue
			}
			ss.EnsureDecoded()
			spanCount += len(ss.Spans)
		}
	}
	return spanCount
}
