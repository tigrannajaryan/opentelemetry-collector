// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/pdata/internal"

type MapWrapper struct {
	orig   *[]KeyValue
	state  *State
	marker *LazyMessage
}

func GetMapOrig(ms MapWrapper) *[]KeyValue {
	return ms.orig
}

func GetMapState(ms MapWrapper) *State {
	return ms.state
}

func GetMapLazyMessage(ms MapWrapper) *LazyMessage {
	return ms.marker
}

func NewMapWrapper(orig *[]KeyValue, state *State) MapWrapper {
	return NewMapWrapperWithLazyMessage(orig, state, nil)
}

func NewMapWrapperWithLazyMessage(orig *[]KeyValue, state *State, marker *LazyMessage) MapWrapper {
	return MapWrapper{orig: orig, state: state, marker: marker}
}

func GenTestMapWrapper() MapWrapper {
	orig := GenTestKeyValueSlice()
	return NewMapWrapper(&orig, NewState())
}
