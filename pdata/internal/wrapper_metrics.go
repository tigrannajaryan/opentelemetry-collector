// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/pdata/internal"

// MetricsToProto internal helper to convert Metrics to protobuf representation.
func MetricsToProto(l MetricsWrapper) MetricsData {
	l.orig.EnsureDecoded()
	return MetricsData{
		ResourceMetrics: l.orig.ResourceMetrics,
	}
}

// MetricsFromProto internal helper to convert protobuf representation to Metrics.
// This function set exclusive state assuming that it's called only once per Metrics.
func MetricsFromProto(orig MetricsData) MetricsWrapper {
	orig.EnsureDecoded()
	return NewMetricsWrapper(&ExportMetricsServiceRequest{
		ResourceMetrics: orig.ResourceMetrics,
	}, NewState())
}
