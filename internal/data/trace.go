// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

import (
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"
)

// This file defines in-memory data structures to represent traces (spans).

// ITraceData is the top-level struct that is propagated through the traces pipeline.
// This is the newer version of consumerdata.TraceData, but uses more efficient
// in-memory representation.
type ITraceData struct {
	resourceSpans []*ResourceSpans
}

func NewITraceData(resourceSpans []*ResourceSpans) ITraceData {
	return ITraceData{resourceSpans}
}

// SpanCount calculates the total number of spans.
func (td ITraceData) SpanCount() int {
	spanCount := 0
	for _, rsl := range td.resourceSpans {
		spanCount += len(rsl.spans)
	}
	return spanCount
}

// A collection of spans from a Resource.
type ResourceSpans struct {
	// The resource for the spans in this message.
	// If this field is not set then no resource info is known.
	resource *Resource

	// A list of Spans that originate from a resource.
	spans []*Span
}

func NewResourceSpans(resource *Resource, spans []*Span) *ResourceSpans {
	return &ResourceSpans{resource, spans}
}

func (m *ResourceSpans) Resource() *Resource {
	return m.resource
}

func (m *ResourceSpans) SetResource(r *Resource) {
	m.resource = r
}

func (m *ResourceSpans) Spans() []*Span {
	return m.spans
}

func (m *ResourceSpans) SetSpans(s []*Span) {
	m.spans = s
}

type TraceID struct {
	bytes []byte
}

func TraceIDFromBytes(bytes []byte) TraceID { return TraceID{bytes} }

type SpanID struct {
	bytes []byte
}

func SpanIDFromBytes(bytes []byte) SpanID { return SpanID{bytes} }

// TraceState in w3c-trace-context format: https://www.w3.org/TR/trace-context/#tracestate-header
type TraceState string

type SpanKind otlptrace.Span_SpanKind

func (sk SpanKind) String() string { return otlptrace.Span_SpanKind(sk).String() }

const (
	SpanKindUNSPECIFIED SpanKind = 0
	SpanKindINTERNAL    SpanKind = SpanKind(otlptrace.Span_INTERNAL)
	SpanKindSERVER      SpanKind = SpanKind(otlptrace.Span_SERVER)
	SpanKindCLIENT      SpanKind = SpanKind(otlptrace.Span_CLIENT)
	SpanKindPRODUCER    SpanKind = SpanKind(otlptrace.Span_PRODUCER)
	SpanKindCONSUMER    SpanKind = SpanKind(otlptrace.Span_CONSUMER)
)

// Span represents a single operation within a trace.
// See Span definition in OTLP: https://github.com/open-telemetry/opentelemetry-proto/blob/master/opentelemetry/proto/trace/v1/trace.proto#L37
//
// Must use NewSpan* functions to create new instances.
// Important: zero-initialized instance is not valid for use.
type Span struct {
	// Wrap OTLP Span.
	orig *otlptrace.Span

	// Override a few fields. These fields are the source of truth. Their counterparts
	// stored in corresponding fields of "orig" are ignored.
	attributes AttributesMap
	events     []*SpanEvent
	links      []*SpanLink
}

func NewSpan() *Span {
	return &Span{orig: &otlptrace.Span{}}
}

// NewSpanSlice creates a slice of Spans that are correctly initialized.
func NewSpanSlice(len int) []Span {
	origs := make([]otlptrace.Span, len)
	wrappers := make([]Span, len)
	for i := range origs {
		wrappers[i].orig = &origs[i]
	}
	return wrappers
}

func (m *Span) TraceID() TraceID {
	return TraceIDFromBytes(m.orig.TraceId)
}

func (m *Span) SpanID() SpanID {
	return SpanIDFromBytes(m.orig.SpanId)
}

func (m *Span) TraceState() TraceState {
	return TraceState(m.orig.Tracestate)
}

func (m *Span) ParentSpanID() SpanID {
	return SpanIDFromBytes(m.orig.ParentSpanId)
}

func (m *Span) Name() string {
	return m.orig.Name
}

func (m *Span) Kind() SpanKind {
	return SpanKind(m.orig.Kind)
}

func (m *Span) StartTime() TimestampUnixNano {
	return TimestampUnixNano(m.orig.StartTimeUnixnano)
}

func (m *Span) EndTime() TimestampUnixNano {
	return TimestampUnixNano(m.orig.EndTimeUnixnano)
}

func (m *Span) Attributes() AttributesMap {
	return m.attributes
}

func (m *Span) DroppedAttributesCount() uint32 {
	return m.orig.DroppedAttributesCount
}

func (m *Span) Events() []*SpanEvent {
	return m.events
}

func (m *Span) DroppedEventsCount() uint32 {
	return m.orig.DroppedEventsCount
}

func (m *Span) Links() []*SpanLink {
	return m.links
}

func (m *Span) DroppedLinksCount() uint32 {
	return m.orig.DroppedLinksCount
}

func (m *Span) Status() SpanStatus {
	return SpanStatus{orig: m.orig.Status}
}

func (m *Span) SetTraceID(v TraceID) {
	m.orig.TraceId = v.bytes
}

func (m *Span) SetSpanID(v SpanID) {
	m.orig.SpanId = v.bytes
}

func (m *Span) SetTraceState(v TraceState) {
	m.orig.Tracestate = string(v)
}

func (m *Span) SetParentSpanID(v SpanID) {
	m.orig.ParentSpanId = v.bytes
}

func (m *Span) SetName(v string) {
	m.orig.Name = v
}

func (m *Span) SetKind(v SpanKind) {
	m.orig.Kind = otlptrace.Span_SpanKind(v)
}

func (m *Span) SetStartTime(v TimestampUnixNano) {
	m.orig.StartTimeUnixnano = uint64(v)
}

func (m *Span) SetEndTime(v TimestampUnixNano) {
	m.orig.EndTimeUnixnano = uint64(v)
}

func (m *Span) SetAttributes(v Attributes) {
	m.attributes = v.attrs
	m.orig.DroppedAttributesCount = v.droppedCount
}

func (m *Span) SetEvents(v []*SpanEvent) {
	m.events = v
}

func (m *Span) SetDroppedEventsCount(v uint32) {
	m.orig.DroppedEventsCount = v
}

func (m *Span) SetLinks(v []*SpanLink) {
	m.links = v
}

func (m *Span) SetDroppedLinksCount(v uint32) {
	m.orig.DroppedLinksCount = v
}

func (m *Span) SetStatus(v SpanStatus) {
	m.orig.Status = v.orig
}

type SpanStatus struct {
	orig *otlptrace.Status
}

// StatusCode mirrors the codes defined at
// https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/api-tracing.md#statuscanonicalcode
// and is numerically equal to Standard GRPC codes https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
type StatusCode otlptrace.Status_StatusCode

func NewSpanStatus(code StatusCode, message string) SpanStatus {
	return SpanStatus{orig: &otlptrace.Status{
		Code:    otlptrace.Status_StatusCode(code),
		Message: message,
	}}
}

// SpanEvent is a time-stamped annotation of the span, consisting of user-supplied
// text description and key-value pairs. See OTLP for event definition.
//
// Must use NewSpanEvent* function to create new instances.
// Important: zero-initialized instance is not valid for use.
type SpanEvent struct {
	// Wrap OTLP Event.
	orig *otlptrace.Span_Event

	// Override attributes. This field is the source of truth for attributes.
	// The counterpart stored in corresponding field of "orig" is ignored.
	attributes AttributesMap
}

func NewSpanEvent(timestamp TimestampUnixNano, name string, attributes Attributes) *SpanEvent {
	return &SpanEvent{
		orig: &otlptrace.Span_Event{
			TimeUnixnano:           uint64(timestamp),
			Name:                   name,
			DroppedAttributesCount: attributes.droppedCount,
		},
		attributes: attributes.attrs,
	}
}

// NewSpanEventSlice creates a slice of SpanEvents that are correctly initialized.
func NewSpanEventSlice(len int) []SpanEvent {
	origs := make([]otlptrace.Span_Event, len)
	wrappers := make([]SpanEvent, len)
	for i := range origs {
		wrappers[i].orig = &origs[i]
	}
	return wrappers
}

func (m *SpanEvent) Timestamp() TimestampUnixNano {
	return TimestampUnixNano(m.orig.TimeUnixnano)
}

func (m *SpanEvent) Name() string {
	return m.orig.Name
}

func (m *SpanEvent) Attributes() AttributesMap {
	return m.attributes
}
func (m *SpanEvent) DroppedAttributesCount() uint32 {
	return m.orig.DroppedAttributesCount
}

func (m *SpanEvent) SetTimestamp(v TimestampUnixNano) {
	m.orig.TimeUnixnano = uint64(v)
}

func (m *SpanEvent) SetName(v string) {
	m.orig.Name = v
}

func (m *SpanEvent) SetAttributes(v Attributes) {
	m.attributes = v.attrs
	m.orig.DroppedAttributesCount = v.droppedCount
}

// SpanLink is a pointer from the current span to another span in the same trace or in a
// different trace. See OTLP for link definition.
//
// Must use NewSpanLink* function to create new instances.
// Important: zero-initialized instance is not valid for use.
type SpanLink struct {
	// Wrap OTLP Link.
	orig *otlptrace.Span_Link

	// Override attributes. This field is the source of truth for attributes.
	// The counterpart stored in corresponding field of "orig" is ignored.
	attributes AttributesMap
}

// NewSpanLink creates a SpanLink that is correctly initialized.
func NewSpanLink() *SpanLink {
	return &SpanLink{orig: &otlptrace.Span_Link{}}
}

// NewSpanLinkSlice creates a slice of SpanLinks that are correctly initialized.
func NewSpanLinkSlice(len int) []SpanLink {
	origs := make([]otlptrace.Span_Link, len)
	wrappers := make([]SpanLink, len)
	for i := range origs {
		wrappers[i].orig = &origs[i]
	}
	return wrappers
}

func (m *SpanLink) TraceID() TraceID {
	return TraceIDFromBytes(m.orig.TraceId)
}

func (m *SpanLink) SpanID() SpanID {
	return SpanIDFromBytes(m.orig.SpanId)
}

func (m *SpanLink) Attributes() AttributesMap {
	return m.attributes
}

func (m *SpanLink) DroppedAttributesCount() uint32 {
	return m.orig.DroppedAttributesCount
}

func (m *SpanLink) TraceState() TraceState {
	return TraceState(m.orig.Tracestate)
}

func (m *SpanLink) SetTraceID(v TraceID) {
	m.orig.TraceId = v.bytes
}

func (m *SpanLink) SetSpanID(v SpanID) {
	m.orig.SpanId = v.bytes
}

func (m *SpanLink) SetTraceState(v TraceState) {
	m.orig.Tracestate = string(v)
}

func (m *SpanLink) SetAttributes(v Attributes) {
	m.attributes = v.attrs
	m.orig.DroppedAttributesCount = v.droppedCount
}
