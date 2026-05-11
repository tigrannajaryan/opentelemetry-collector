// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/pdata/internal"

func validateTraceIDProto(buf []byte) error {
	var traceID TraceID
	return traceID.UnmarshalProto(buf)
}

func validateSpanIDProto(buf []byte) error {
	var spanID SpanID
	return spanID.UnmarshalProto(buf)
}

func validateProfileIDProto(buf []byte) error {
	var profileID ProfileID
	return profileID.UnmarshalProto(buf)
}
