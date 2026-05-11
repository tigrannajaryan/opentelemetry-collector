// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/pdata/internal"
)

func TestDeprecatedScopeMetrics(t *testing.T) {
	sm := new(internal.ScopeMetrics)
	rms := []*internal.ResourceMetrics{
		{
			ScopeMetrics:           []*internal.ScopeMetrics{sm},
			DeprecatedScopeMetrics: []*internal.ScopeMetrics{sm},
		},
		{
			ScopeMetrics:           []*internal.ScopeMetrics{},
			DeprecatedScopeMetrics: []*internal.ScopeMetrics{sm},
		},
	}

	MigrateMetrics(rms)
	assert.Same(t, sm, rms[0].ScopeMetrics[0])
	assert.Same(t, sm, rms[1].ScopeMetrics[0])
	assert.Nil(t, rms[0].DeprecatedScopeMetrics)
	assert.Nil(t, rms[0].DeprecatedScopeMetrics)
}

func TestDeprecatedScopeMetricsMigrationKeepsNestedMessagesLazy(t *testing.T) {
	src := &internal.ResourceMetrics{
		DeprecatedScopeMetrics: []*internal.ScopeMetrics{internal.GenTestScopeMetrics()},
	}
	buf := make([]byte, src.SizeProto())
	src.MarshalProto(buf)

	rm := internal.NewResourceMetrics()
	assert.NoError(t, rm.UnmarshalProto(buf))

	MigrateMetrics([]*internal.ResourceMetrics{rm})

	assert.Len(t, rm.ScopeMetrics, 1)
	assert.Nil(t, rm.DeprecatedScopeMetrics)
	assert.True(t, rm.ScopeMetrics[0].LazyMessage().NeedsDecode())
}
