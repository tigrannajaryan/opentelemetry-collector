// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlp // import "go.opentelemetry.io/collector/pdata/internal/otlp"

import (
	"go.opentelemetry.io/collector/pdata/internal"
)

// MigrateProfiles implements any translation needed due to deprecation in OTLP profiles protocol.
// It is currently a no-op, so proto receive paths do not need to materialize
// profile data just to call it.
func MigrateProfiles(_ []*internal.ResourceProfiles) {}
