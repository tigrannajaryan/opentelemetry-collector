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
	parent  *LazyMessage
	bytes   []byte
	decoded bool
}

// Init attaches protobuf wire bytes to a message and records its parent.
func (m *LazyMessage) Init(bytes []byte, parent *LazyMessage) {
	m.bytes = bytes
	m.parent = parent
	m.decoded = false
}

// SetParent updates the parent message used for modification propagation.
func (m *LazyMessage) SetParent(parent *LazyMessage) {
	m.parent = parent
}

// Bytes returns the attached protobuf wire bytes.
func (m *LazyMessage) Bytes() []byte {
	return m.bytes
}

// HasBytes returns true when the message can still be marshaled from wire bytes.
func (m *LazyMessage) HasBytes() bool {
	return m.bytes != nil
}

// IsDecoded reports whether the in-memory fields have been populated.
func (m *LazyMessage) IsDecoded() bool {
	return !m.NeedsDecode()
}

// NeedsDecode reports whether the message still needs its wire bytes decoded.
// It is intentionally tiny so generated EnsureDecoded methods have an inlinable
// fast path for the common already-decoded or locally-built message.
func (m *LazyMessage) NeedsDecode() bool {
	return m.bytes != nil && !m.decoded
}

// MarkDecoded records that in-memory fields have been populated from the bytes.
func (m *LazyMessage) MarkDecoded() {
	if m.bytes != nil {
		m.decoded = true
	}
}

// MarkModified clears reusable wire bytes on this message and all parents. It
// intentionally does not decode first; mutating callers must materialize the
// fields they are about to change before discarding the raw representation.
//
// The walk intentionally continues through parents whose own bytes are already
// nil. A child can be fully materialized before it is changed while an ancestor
// still has reusable bytes, and that ancestor must stop using its raw
// representation as soon as any descendant changes.
func (m *LazyMessage) MarkModified() {
	if m.bytes != nil {
		m.bytes = nil
		m.decoded = true
	}
	for parent := m.parent; parent != nil; parent = parent.parent {
		if parent.bytes != nil {
			parent.bytes = nil
			parent.decoded = true
		}
	}
}

// Clear removes all lazy state. It is used after full materialization when the
// caller wants ordinary in-memory equality semantics instead of wire reuse.
func (m *LazyMessage) Clear() {
	*m = LazyMessage{}
}
