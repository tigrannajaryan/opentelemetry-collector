// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package proto // import "go.opentelemetry.io/collector/pdata/internal/proto"

// LazyMessage stores the protobuf wire bytes attached to an unmarshaled message.
//
// A non-nil bytes slice means the message still has a valid wire representation
// that can be copied directly during protobuf marshaling. MarkModified clears the
// bytes and walks up the parent chain so a changed child also forces all parents
// to reassemble their protobuf output.
type LazyMessage struct {
	state *lazyMessageState
}

type lazyMessageState struct {
	parent   *LazyMessage
	bytes    []byte
	decoded  bool
	modified bool
}

// Init attaches protobuf wire bytes to a message and records its parent.
func (m *LazyMessage) Init(bytes []byte, parent *LazyMessage) {
	m.state = &lazyMessageState{
		parent: parent,
		bytes:  bytes,
	}
}

// SetParent updates the parent message used for modification propagation.
func (m *LazyMessage) SetParent(parent *LazyMessage) {
	if parent == nil {
		if m.state != nil {
			m.state.parent = nil
		}
		return
	}
	if m.state == nil {
		m.state = &lazyMessageState{parent: parent}
		return
	}
	if m.state.parent != parent {
		// A previous modification only proves that the old parent chain was
		// invalidated. Re-parenting needs a fresh propagation on the next
		// mutation so the new ancestors cannot keep stale wire bytes.
		m.state.modified = false
	}
	m.state.parent = parent
}

// Bytes returns the attached protobuf wire bytes.
func (m *LazyMessage) Bytes() []byte {
	state := m.state
	if state == nil {
		return nil
	}
	return state.bytes
}

// HasBytes returns true when the message can still be marshaled from wire bytes.
func (m *LazyMessage) HasBytes() bool {
	state := m.state
	return state != nil && state.bytes != nil
}

// MutationMarker returns m when mutations to a wrapper that points inside this
// message need to invalidate existing wire bytes. Freshly constructed messages
// have neither bytes nor a parent, so returning nil keeps ordinary in-memory
// pdata wrappers small for equality and avoids unnecessary MarkModified calls.
//
// A message with no bytes but with a parent still needs to be returned. A child
// can be materialized independently while an ancestor keeps reusable bytes, and
// mutating the child must still bubble to that ancestor.
func (m *LazyMessage) MutationMarker() *LazyMessage {
	state := m.state
	if state == nil || state.modified || (state.bytes == nil && state.parent == nil) {
		return nil
	}
	return m
}

// IsDecoded reports whether the in-memory fields have been populated.
func (m *LazyMessage) IsDecoded() bool {
	return !m.NeedsDecode()
}

// NeedsDecode reports whether the message still needs its wire bytes decoded.
// It is intentionally tiny so generated EnsureDecoded methods have an inlinable
// fast path for the common already-decoded or locally-built message.
func (m *LazyMessage) NeedsDecode() bool {
	state := m.state
	return state != nil && state.bytes != nil && !state.decoded
}

// MarkDecoded records that in-memory fields have been populated from the bytes.
func (m *LazyMessage) MarkDecoded() {
	state := m.state
	if state != nil && state.bytes != nil {
		state.decoded = true
	}
}

// MarkModified clears reusable wire bytes on this message and all parents. It
// intentionally does not decode first; mutating callers must materialize the
// fields they are about to change before discarding the raw representation.
//
// The walk continues through unmodified parents whose own bytes are already nil.
// A child can be fully materialized before it is changed while an ancestor still
// has reusable bytes, and that ancestor must stop using its raw representation
// as soon as any descendant changes. Once an already-modified parent is reached,
// the walk stops because that parent previously invalidated all of its
// ancestors.
func (m *LazyMessage) MarkModified() {
	// Keep the common already-modified and no-lazy-state paths in this tiny
	// method so callers can inline the no-op case.
	state := m.state
	if state == nil || state.modified || (state.bytes == nil && state.parent == nil) {
		return
	}
	m.markModified(state)
}

func (m *LazyMessage) markModified(state *lazyMessageState) {
	state.modified = true
	if state.bytes != nil {
		state.bytes = nil
		state.decoded = true
	}
	for parent := state.parent; parent != nil; {
		parentState := parent.state
		if parentState == nil || parentState.modified || (parentState.bytes == nil && parentState.parent == nil) {
			return
		}
		parentState.modified = true
		if parentState.bytes != nil {
			parentState.bytes = nil
			parentState.decoded = true
		}
		parent = parentState.parent
	}
}

// Clear removes all lazy state. It is used after full materialization when the
// caller wants ordinary in-memory equality semantics instead of wire reuse.
func (m *LazyMessage) Clear() {
	*m = LazyMessage{}
}
