// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/pdata/internal"

import "go.opentelemetry.io/collector/pdata/internal/proto"

// LazyMessage is the per-message protobuf lazy state used by generated pdata
// wrappers to invalidate reusable wire bytes when a nested collection mutates.
type LazyMessage = proto.LazyMessage
