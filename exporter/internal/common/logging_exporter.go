// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package common // import "go.opentelemetry.io/collector/exporter/internal/common"

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter/internal/otlptext"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type loggingExporter struct {
	verbosity        configtelemetry.Level
	logger           *zap.Logger
	logsMarshaler    plog.Marshaler
	metricsMarshaler pmetric.Marshaler
	tracesMarshaler  ptrace.Marshaler
}

func (s *loggingExporter) pushTraces(_ context.Context, td ptrace.Traces) error {
	s.logger.Info("TracesExporter",
		zap.Int("resource spans", td.ResourceSpans().Len()),
		zap.Int("spans", td.SpanCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.tracesMarshaler.MarshalTraces(td)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))

	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		scs := rss.At(i).ScopeSpans()
		for j := 0; j < scs.Len(); j++ {
			scope := scs.At(j).Scope()
			spans := scs.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				fmt.Printf("%s ", span.StartTimestamp().AsTime().Format("15:04:05.000000"))
				fmt.Printf("%-18s|", scope.Name()+color.HiBlackString("@")+scope.Version())

				var kind string
				switch span.Kind() {
				case ptrace.SpanKindUnspecified:
					kind = "U"
				case ptrace.SpanKindClient:
					kind = color.MagentaString("C")
				case ptrace.SpanKindServer:
					kind = color.CyanString("S")
				case ptrace.SpanKindInternal:
					kind = "I"
				case ptrace.SpanKindConsumer:
					kind = "<"
				case ptrace.SpanKindProducer:
					kind = ">"
				}
				fmt.Printf("%s|", kind)

				duration := fmt.Sprintf("%8s", span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime()).String())
				fmt.Printf("%s", duration)

				name := fmt.Sprintf("|%-20s|", span.Name())
				statusCode := span.Status().Code()
				switch statusCode {
				case ptrace.StatusCodeOk:
					name = color.HiGreenString(name)
				case ptrace.StatusCodeError:
					name = color.RedString(name)
				case ptrace.StatusCodeUnset:
					name = color.GreenString(name)
				}
				fmt.Printf("%s", name)

				attrs := span.Attributes()
				attrs.Range(
					func(k string, v pcommon.Value) bool {
						fmt.Printf("%s=%s ", k, v.AsString())
						return true
					})

				fmt.Print("\n")

				fmt.Printf("  %-32s:", color.HiBlackString(span.TraceID().String()))
				fmt.Printf("%-16s>", color.HiBlackString(span.ParentSpanID().String()))
				fmt.Printf("%-16s", color.HiBlackString(span.SpanID().String()))
				fmt.Print("\n")
			}
		}
	}

	return nil
}

func (s *loggingExporter) pushMetrics(_ context.Context, md pmetric.Metrics) error {
	s.logger.Info("MetricsExporter",
		zap.Int("resource metrics", md.ResourceMetrics().Len()),
		zap.Int("metrics", md.MetricCount()),
		zap.Int("data points", md.DataPointCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func (s *loggingExporter) pushLogs(_ context.Context, ld plog.Logs) error {
	s.logger.Info("LogsExporter",
		zap.Int("resource logs", ld.ResourceLogs().Len()),
		zap.Int("log records", ld.LogRecordCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func newLoggingExporter(logger *zap.Logger, verbosity configtelemetry.Level) *loggingExporter {
	return &loggingExporter{
		verbosity:        verbosity,
		logger:           logger,
		logsMarshaler:    otlptext.NewTextLogsMarshaler(),
		metricsMarshaler: otlptext.NewTextMetricsMarshaler(),
		tracesMarshaler:  otlptext.NewTextTracesMarshaler(),
	}
}
