// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporterhelper

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector/component"
	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-collector/exporter"
	"github.com/open-telemetry/opentelemetry-collector/internal/data"
	"github.com/open-telemetry/opentelemetry-collector/obsreport"
)

// traceDataPusher is a helper function that is similar to ConsumeTraceData but also
// returns the number of dropped spans.
type traceDataPusher func(ctx context.Context, td consumerdata.TraceData) (droppedSpans int, err error)

// iTraceDataPusher is a helper function that is similar to ConsumeTraceData but also
// returns the number of dropped spans.
type iTraceDataPusher func(ctx context.Context, td data.ITraceData) (droppedSpans int, err error)

// traceExporter implements the exporter with additional helper options.
type traceExporter struct {
	exporterFullName string
	dataPusher       traceDataPusher
	shutdown         Shutdown
}

var _ exporter.TraceExporter = (*traceExporter)(nil)

func (te *traceExporter) Start(host component.Host) error {
	return nil
}

func (te *traceExporter) ConsumeTraceData(ctx context.Context, td consumerdata.TraceData) error {
	exporterCtx := obsreport.ExporterContext(ctx, te.exporterFullName)
	_, err := te.dataPusher(exporterCtx, td)
	return err
}

// Shutdown stops the exporter and is invoked during shutdown.
func (te *traceExporter) Shutdown() error {
	return te.shutdown()
}

// NewTraceExporter creates an TraceExporter that can record metrics and can wrap every
// request with a Span. If no options are passed it just adds the exporter format as a
// tag in the Context.
func NewTraceExporter(
	config configmodels.Exporter,
	dataPusher traceDataPusher,
	options ...ExporterOption,
) (exporter.TraceExporter, error) {

	if config == nil {
		return nil, errNilConfig
	}

	if dataPusher == nil {
		return nil, errNilPushTraceData
	}

	opts := newExporterOptions(options...)

	dataPusher = dataPusher.withObservability(config.Name())

	// The default shutdown function does nothing.
	if opts.shutdown == nil {
		opts.shutdown = func() error {
			return nil
		}
	}

	return &traceExporter{
		exporterFullName: config.Name(),
		dataPusher:       dataPusher,
		shutdown:         opts.shutdown,
	}, nil
}

// withObservability wraps the current pusher into a function that records
// the observability signals during the pusher execution.
func (p traceDataPusher) withObservability(exporterName string) traceDataPusher {
	return func(ctx context.Context, td consumerdata.TraceData) (int, error) {
		exporterCtx, span := obsreport.StartTraceDataExportOp(ctx, exporterName)
		// Forward the data to the next consumer (this pusher is the next).
		droppedSpans, err := p(exporterCtx, td)

		// TODO: this is not ideal: it should come from the next function itself.
		// 	temporarily loading it from internal format. Once full switch is done
		// 	to new metrics will remove this.
		numSpans := len(td.Spans)
		obsreport.EndTraceDataExportOp(exporterCtx, span, numSpans, droppedSpans, err)
		return droppedSpans, err
	}
}

type iTraceExporter struct {
	exporterFullName string
	dataPusher       iTraceDataPusher
	shutdown         Shutdown
}

var _ exporter.ITraceExporter = (*iTraceExporter)(nil)

func (te *iTraceExporter) Start(host component.Host) error {
	return nil
}

func (te *iTraceExporter) ConsumeITrace(
	ctx context.Context,
	td data.ITraceData,
) error {
	exporterCtx := obsreport.ExporterContext(ctx, te.exporterFullName)
	_, err := te.dataPusher(exporterCtx, td)
	return err
}

// Shutdown stops the exporter and is invoked during shutdown.
func (te *iTraceExporter) Shutdown() error {
	return te.shutdown()
}

// NewITraceExporter creates an ITraceExporter that can record metrics and can wrap
// every request with a Span.
func NewITraceExporter(
	config configmodels.Exporter,
	dataPusher iTraceDataPusher,
	options ...ExporterOption,
) (exporter.ITraceExporter, error) {

	if config == nil {
		return nil, errNilConfig
	}

	if dataPusher == nil {
		return nil, errNilPushTraceData
	}

	opts := newExporterOptions(options...)

	dataPusher = dataPusher.withObservability(config.Name())

	// The default shutdown function does nothing.
	if opts.shutdown == nil {
		opts.shutdown = func() error {
			return nil
		}
	}

	return &iTraceExporter{
		exporterFullName: config.Name(),
		dataPusher:       dataPusher,
		shutdown:         opts.shutdown,
	}, nil
}

// withObservability wraps the current pusher into a function that records
// the observability signals during the pusher execution.
func (p iTraceDataPusher) withObservability(exporterName string) iTraceDataPusher {
	return func(ctx context.Context, td data.ITraceData) (int, error) {
		exporterCtx, span := obsreport.StartTraceDataExportOp(ctx, exporterName)
		// Forward the data to the next consumer (this pusher is the next).
		droppedSpans, err := p(exporterCtx, td)

		// TODO: this is not ideal: it should come from the next function itself.
		// 	temporarily loading it from internal format. Once full switch is done
		// 	to new metrics will remove this.
		numSpans := td.SpanCount()
		obsreport.EndTraceDataExportOp(exporterCtx, span, numSpans, droppedSpans, err)
		return droppedSpans, err
	}
}
