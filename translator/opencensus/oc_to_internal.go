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

package opencensus

import (
	"strings"

	occommon "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	ocresource "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	octrace "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"

	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-collector/internal"
	"github.com/open-telemetry/opentelemetry-collector/internal/data"
	"github.com/open-telemetry/opentelemetry-collector/translator/conventions"
	tracetranslator "github.com/open-telemetry/opentelemetry-collector/translator/trace"
)

// OTLP attributes to map certain OpenCensus proto fields. These fields don't have
// corresponding fields in OTLP, nor are defined in OTLP semantic conventions.
// TODO: decide if any of these must be in OTLP semantic conventions.
const (
	ocAttributeProcessStartTime  = "opencensus.starttime"
	ocAttributeProcessID         = "opencensus.pid"
	ocAttributeExporterVersion   = "opencensus.exporterversion"
	ocAttributeResourceType      = "opencensus.resourcetype"
	ocTimeEventMessageEventType  = "opencensus.timeevent.messageevent.type"
	ocTimeEventMessageEventID    = "opencensus.timeevent.messageevent.id"
	ocTimeEventMessageEventUSize = "opencensus.timeevent.messageevent.usize"
	ocTimeEventMessageEventCSize = "opencensus.timeevent.messageevent.csize"
)

func ocToInternal(td consumerdata.TraceData) data.ITraceData {

	if td.Node == nil && td.Resource == nil && len(td.Spans) == 0 {
		return data.ITraceData{}
	}

	resource := ocNodeResourceToInternal(td.Node, td.Resource)

	resourceSpans := &data.ResourceSpans{}
	resourceSpans.SetResource(resource)
	resourceSpanList := []*data.ResourceSpans{resourceSpans}

	spanCount := len(td.Spans)

	if spanCount != 0 {
		// Create slice that holds pointers and another slice that holds structs at once
		// to avoid individual allocations of structs.
		content := data.NewSpanSlice(spanCount)
		ptrs := make([]*data.Span, 0, spanCount)

		for i, ocSpan := range td.Spans {
			if ocSpan == nil {
				// Skip nil spans.
				continue
			}

			// Point one element in slice of pointers to an element in slice of structs.
			dest := &content[i]

			ocSpanToInternal(dest, ocSpan)

			if ocSpan.Resource != nil {
				// Add a separate ResourceSpans item just for this span since it
				// has a different Resource.
				separateRS := &data.ResourceSpans{}
				separateRS.SetResource(ocNodeResourceToInternal(nil, ocSpan.Resource))
				separateRS.SetSpans([]*data.Span{dest})
				resourceSpanList = append(resourceSpanList, separateRS)
			} else {
				// Otherwise add the span to the first ResourceSpans item.
				ptrs = append(ptrs, dest)
			}
		}

		resourceSpans.SetSpans(ptrs)
	}

	return data.NewITraceData(resourceSpanList)
}

func timestampToUnixnano(ts *timestamp.Timestamp) data.TimestampUnixNano {
	return data.TimestampUnixNano(uint64(internal.TimestampToTime(ts).UnixNano()))
}

func ocSpanToInternal(dest *data.Span, src *octrace.Span) {
	events, droppedEventCount := ocEventsToInternal(src.TimeEvents)
	links, droppedLinkCount := ocLinksToInternal(src.Links)

	dest.SetTraceID(data.TraceIDFromBytes(src.TraceId))
	dest.SetSpanID(data.SpanIDFromBytes(src.SpanId))
	dest.SetTraceState(ocTraceStateToInternal(src.Tracestate))
	dest.SetParentSpanID(data.SpanIDFromBytes(src.ParentSpanId))
	dest.SetName(truncableStringToStr(src.Name))
	dest.SetKind(ocSpanKindToInternal(src.Kind, src.Attributes))
	dest.SetStartTime(timestampToUnixnano(src.StartTime))
	dest.SetEndTime(timestampToUnixnano(src.EndTime))
	dest.SetAttributes(ocAttrsToInternal(src.Attributes))
	dest.SetEvents(events)
	dest.SetDroppedEventsCount(droppedEventCount)
	dest.SetLinks(links)
	dest.SetDroppedLinksCount(droppedLinkCount)
	dest.SetStatus(ocStatusToInternal(src.Status))
}

func ocStatusToInternal(ocStatus *octrace.Status) data.SpanStatus {
	if ocStatus == nil {
		return data.SpanStatus{}
	}
	return data.NewSpanStatus(otlptrace.Status_StatusCode(ocStatus.Code), ocStatus.Message)
}

// Convert tracestate to W3C format. See the https://w3c.github.io/trace-context/
func ocTraceStateToInternal(ocTracestate *octrace.Span_Tracestate) data.TraceState {
	if ocTracestate == nil {
		return ""
	}
	var sb strings.Builder
	for i, entry := range ocTracestate.Entries {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strings.Join([]string{entry.Key, entry.Value}, "="))
	}
	return data.TraceState(sb.String())
}

func ocAttrsToInternal(ocAttrs *octrace.Span_Attributes) data.Attributes {
	if ocAttrs == nil {
		return data.Attributes{}
	}

	attrCount := len(ocAttrs.AttributeMap)

	attrMap := make(data.AttributesMap, attrCount)
	values := data.NewAttributeValueSlice(attrCount)
	i := 0
	for key, ocAttr := range ocAttrs.AttributeMap {
		switch attribValue := ocAttr.Value.(type) {
		case *octrace.AttributeValue_StringValue:
			values[i].MakeString(truncableStringToStr(attribValue.StringValue))

		case *octrace.AttributeValue_IntValue:
			values[i].MakeInt(attribValue.IntValue)

		case *octrace.AttributeValue_BoolValue:
			values[i].MakeBool(attribValue.BoolValue)

		case *octrace.AttributeValue_DoubleValue:
			values[i].MakeDouble(attribValue.DoubleValue)

		default:
			str := "<Unknown OpenCensus Attribute>"
			values[i].MakeString(str)
		}
		attrMap[key] = values[i]
		i++
	}
	droppedCount := uint32(ocAttrs.DroppedAttributesCount)
	return data.NewAttributes(attrMap, droppedCount)
}

func ocSpanKindToInternal(ocKind octrace.Span_SpanKind, ocAttrs *octrace.Span_Attributes) otlptrace.Span_SpanKind {
	switch ocKind {
	case octrace.Span_SERVER:
		return otlptrace.Span_SERVER

	case octrace.Span_CLIENT:
		return otlptrace.Span_CLIENT

	case octrace.Span_SPAN_KIND_UNSPECIFIED:
		// Span kind field is unspecified, check if TagSpanKind attribute is set.
		// This can happen if span kind had no equivalent in OC, so we could represent it in
		// the SpanKind. In that case the span kind may be a special attribute TagSpanKind.
		if ocAttrs != nil {
			kindAttr := ocAttrs.AttributeMap[tracetranslator.TagSpanKind]
			if kindAttr != nil {
				strVal, ok := kindAttr.Value.(*octrace.AttributeValue_StringValue)
				if ok && strVal != nil {
					var otlpKind otlptrace.Span_SpanKind
					switch tracetranslator.OpenTracingSpanKind(truncableStringToStr(strVal.StringValue)) {
					case tracetranslator.OpenTracingSpanKindConsumer:
						otlpKind = otlptrace.Span_CONSUMER
					case tracetranslator.OpenTracingSpanKindProducer:
						otlpKind = otlptrace.Span_PRODUCER
					default:
						return otlptrace.Span_SPAN_KIND_UNSPECIFIED
					}
					delete(ocAttrs.AttributeMap, tracetranslator.TagSpanKind)
					return otlpKind
				}
			}
		}
		return otlptrace.Span_SPAN_KIND_UNSPECIFIED

	default:
		return otlptrace.Span_SPAN_KIND_UNSPECIFIED
	}
}

func ocEventsToInternal(ocEvents *octrace.Span_TimeEvents) (ptrs []*data.SpanEvent, droppedCount uint32) {
	if ocEvents == nil {
		return
	}

	droppedCount = uint32(ocEvents.DroppedMessageEventsCount + ocEvents.DroppedAnnotationsCount)

	evenCount := len(ocEvents.TimeEvent)
	if evenCount == 0 {
		return
	}

	// Create slice that holds pointers and another slice that holds structs at once
	// to avoid individual allocations of structs.
	ptrs = make([]*data.SpanEvent, 0, evenCount)
	content := data.NewSpanEventSlice(evenCount)

	for i, ocEvent := range ocEvents.TimeEvent {
		if ocEvent == nil {
			continue
		}

		// Point one element in slice of pointers to an element in slice of structs.
		event := &content[i]
		ptrs = append(ptrs, event)

		event.SetTimestamp(timestampToUnixnano(ocEvent.Time))

		switch teValue := ocEvent.Value.(type) {
		case *octrace.Span_TimeEvent_Annotation_:
			if teValue.Annotation != nil {
				event.SetName(truncableStringToStr(teValue.Annotation.Description))
				attrs := ocAttrsToInternal(teValue.Annotation.Attributes)
				event.SetAttributes(attrs)
			}

		case *octrace.Span_TimeEvent_MessageEvent_:
			event.SetAttributes(ocMessageEventToInternalAttrs(teValue.MessageEvent))

		default:
			event.SetName("An unknown OpenCensus TimeEvent type was detected when translating")
		}
	}
	return
}

func ocLinksToInternal(ocLinks *octrace.Span_Links) (ptrs []*data.SpanLink, droppedCount uint32) {
	if ocLinks == nil {
		return
	}

	droppedCount = uint32(ocLinks.DroppedLinksCount)

	linkCount := len(ocLinks.Link)
	if linkCount == 0 {
		return
	}

	// Create slice that holds pointers and another slice that holds structs at once
	// to avoid individual allocations of structs.
	ptrs = make([]*data.SpanLink, 0, linkCount)
	content := data.NewSpanLinkSlice(linkCount)

	for i, ocLink := range ocLinks.Link {
		if ocLink == nil {
			continue
		}

		// Point one element in slice of pointers to an element in slice of structs.
		link := &content[i]
		ptrs = append(ptrs, link)

		attrs := ocAttrsToInternal(ocLink.Attributes)
		link.SetTraceID(data.TraceIDFromBytes(ocLink.TraceId))
		link.SetSpanID(data.SpanIDFromBytes(ocLink.SpanId))
		link.SetTraceState(ocTraceStateToInternal(ocLink.Tracestate))
		link.SetAttributes(attrs)
	}
	return ptrs, droppedCount
}

func ocMessageEventToInternalAttrs(msgEvent *octrace.Span_TimeEvent_MessageEvent) data.Attributes {
	if msgEvent == nil {
		return data.Attributes{}
	}

	attrs := data.NewAttributeValueSlice(4)
	attrs[0].MakeString(msgEvent.Type.String())
	attrs[1].MakeInt(int64(msgEvent.Id))
	attrs[2].MakeInt(int64(msgEvent.UncompressedSize))
	attrs[3].MakeInt(int64(msgEvent.CompressedSize))

	return data.NewAttributes(
		map[string]data.AttributeValue{
			ocTimeEventMessageEventType:  attrs[0],
			ocTimeEventMessageEventID:    attrs[1],
			ocTimeEventMessageEventUSize: attrs[2],
			ocTimeEventMessageEventCSize: attrs[3],
		},
		0)
}

func truncableStringToStr(ts *octrace.TruncatableString) string {
	if ts == nil {
		return ""
	}
	return ts.Value
}

func ocNodeResourceToInternal(ocNode *occommon.Node, ocResource *ocresource.Resource) *data.Resource {
	resource := &data.Resource{}

	// Number of special fields in the Node. See the code below that deals with special fields.
	const specialNodeAttrCount = 7

	// Number of special fields in the Resource.
	const specialResourceAttrCount = 1

	// Calculate maximum total number of attributes. It is OK if we are a bit higher than
	// the exact number since this is only needed for capacity reservation.
	maxTotalAttrCount := 0
	if ocNode != nil {
		maxTotalAttrCount += len(ocNode.Attributes) + specialNodeAttrCount
	}
	if ocResource != nil {
		maxTotalAttrCount += len(ocResource.Labels) + specialResourceAttrCount
	}

	// Create a map where we will place all attributes from the Node and Resource.
	attrs := make(data.AttributesMap, maxTotalAttrCount)

	if ocNode != nil {
		// Copy all Attributes.
		for k, v := range ocNode.Attributes {
			attrs[k] = data.NewAttributeValueString(v)
		}

		// Add all special fields.
		if ocNode.ServiceInfo != nil {
			if ocNode.ServiceInfo.Name != "" {
				attrs[conventions.AttributeServiceName] = data.NewAttributeValueString(
					ocNode.ServiceInfo.Name)
			}
		}
		if ocNode.Identifier != nil {
			if ocNode.Identifier.StartTimestamp != nil {
				attrs[ocAttributeProcessStartTime] = data.NewAttributeValueString(
					ptypes.TimestampString(ocNode.Identifier.StartTimestamp))
			}
			if ocNode.Identifier.HostName != "" {
				attrs[conventions.AttributeHostHostname] = data.NewAttributeValueString(
					ocNode.Identifier.HostName)
			}
			if ocNode.Identifier.Pid != 0 {
				attrs[ocAttributeProcessID] = data.NewAttributeValueInt(int64(ocNode.Identifier.Pid))
			}
		}
		if ocNode.LibraryInfo != nil {
			if ocNode.LibraryInfo.CoreLibraryVersion != "" {
				attrs[conventions.AttributeLibraryVersion] = data.NewAttributeValueString(
					ocNode.LibraryInfo.CoreLibraryVersion)
			}
			if ocNode.LibraryInfo.ExporterVersion != "" {
				attrs[ocAttributeExporterVersion] = data.NewAttributeValueString(
					ocNode.LibraryInfo.ExporterVersion)
			}
			if ocNode.LibraryInfo.Language != occommon.LibraryInfo_LANGUAGE_UNSPECIFIED {
				attrs[conventions.AttributeLibraryLanguage] = data.NewAttributeValueString(
					ocNode.LibraryInfo.Language.String())
			}
		}
	}

	if ocResource != nil {
		// Copy resource Labels.
		for k, v := range ocResource.Labels {
			attrs[k] = data.NewAttributeValueString(v)
		}
		// Add special fields.
		if ocResource.Type != "" {
			attrs[ocAttributeResourceType] = data.NewAttributeValueString(ocResource.Type)
		}
	}

	if len(attrs) != 0 {
		resource.SetAttributes(attrs)
	}

	return resource
}
