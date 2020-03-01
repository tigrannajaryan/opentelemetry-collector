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

// This file contains data structures that are common for all telemetry types,
// such as timestamps, attributes, etc.

import (
	otlp "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
)

// TimestampUnixNano is a time specified as UNIX Epoch time in nanoseconds since
// 00:00:00 UTC on 1 January 1970.
type TimestampUnixNano uint64

// AttributeValue represents a value of an attribute. Typically used in an Attributes map.
// Must use one of NewAttributeValue* functions below to create new instances.
// Important: zero-initialized instance is not valid for use.
type AttributeValue struct {
	orig *otlp.AttributeKeyValue
}

func NewAttributeValueString(v string) AttributeValue {
	return AttributeValue{orig: &otlp.AttributeKeyValue{Type: otlp.AttributeKeyValue_STRING, StringValue: v}}
}

func NewAttributeValueInt(v int64) AttributeValue {
	return AttributeValue{orig: &otlp.AttributeKeyValue{Type: otlp.AttributeKeyValue_INT, IntValue: v}}
}

func NewAttributeValueDouble(v float64) AttributeValue {
	return AttributeValue{orig: &otlp.AttributeKeyValue{Type: otlp.AttributeKeyValue_DOUBLE, DoubleValue: v}}
}

func NewAttributeValueBool(v bool) AttributeValue {
	return AttributeValue{orig: &otlp.AttributeKeyValue{Type: otlp.AttributeKeyValue_BOOL, BoolValue: v}}
}

// NewAttributeValueSlice creates a slice of attributes values that are correctly initialized.
func NewAttributeValueSlice(len int) []AttributeValue {
	// Allocate 2 slices, one for AttributeValues, another for underlying OTLP structs.
	// TODO: make one allocation for both slices.
	origs := make([]otlp.AttributeKeyValue, len)
	wrappers := make([]AttributeValue, len)
	for i := range origs {
		wrappers[i].orig = &origs[i]
	}
	return wrappers
}

// All AttributeValue functions bellow must be called only on instances that are created
// via NewAttributeValue* functions. Calling these functions on zero-initialized
// AttributeValue struct will cause a panic.

func (a AttributeValue) Type() otlp.AttributeKeyValue_ValueType {
	return a.orig.Type
}

func (a AttributeValue) StringVal() string {
	return a.orig.StringValue
}

func (a AttributeValue) IntVal() int64 {
	return a.orig.IntValue
}

func (a AttributeValue) DoubleVal() float64 {
	return a.orig.DoubleValue
}

func (a AttributeValue) BoolVal() bool {
	return a.orig.BoolValue
}

func (a AttributeValue) MakeString(v string) {
	a.orig.Type = otlp.AttributeKeyValue_STRING
	a.orig.StringValue = v
}

func (a AttributeValue) MakeInt(v int64) {
	a.orig.Type = otlp.AttributeKeyValue_INT
	a.orig.IntValue = v
}

func (a AttributeValue) MakeDouble(v float64) {
	a.orig.Type = otlp.AttributeKeyValue_DOUBLE
	a.orig.DoubleValue = v
}

func (a AttributeValue) MakeBool(v bool) {
	a.orig.Type = otlp.AttributeKeyValue_BOOL
	a.orig.BoolValue = v
}

// AttributesMap stores a map of attribute keys to values.
type AttributesMap map[string]AttributeValue

// Attributes stores the map of attributes and a number of dropped attributes.
// Typically used by translator functions to easily pass the pair.
type Attributes struct {
	attrs        AttributesMap
	droppedCount uint32
}

func NewAttributes(m AttributesMap, droppedCount uint32) Attributes {
	return Attributes{m, droppedCount}
}
