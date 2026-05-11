// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"bytes"
	"testing"
)

func TestLazyMessageMarkModifiedIsNoopAfterModified(t *testing.T) {
	var msg LazyMessage
	msg.Init([]byte("message"), nil)

	msg.MarkModified()
	if !msg.modified {
		t.Fatal("expected message to be marked modified")
	}

	parent := msg.parent
	decoded := msg.decoded
	modified := msg.modified
	wire := append([]byte(nil), msg.bytes...)

	msg.MarkModified()

	if msg.parent != parent {
		t.Fatal("second MarkModified changed parent")
	}
	if msg.decoded != decoded {
		t.Fatal("second MarkModified changed decoded state")
	}
	if msg.modified != modified {
		t.Fatal("second MarkModified changed modified state")
	}
	if !bytes.Equal(msg.bytes, wire) {
		t.Fatal("second MarkModified changed bytes")
	}
}

func TestLazyMessageMarkModifiedClearsChildWhenParentAlreadyModified(t *testing.T) {
	var parent, firstChild, secondChild LazyMessage
	parent.Init([]byte("parent"), nil)
	firstChild.Init([]byte("first-child"), &parent)
	secondChild.Init([]byte("second-child"), &parent)

	firstChild.MarkModified()
	if !parent.modified {
		t.Fatal("expected parent to be marked by first child")
	}

	secondChild.MarkModified()
	if secondChild.HasBytes() {
		t.Fatal("expected second child bytes to be cleared even when parent was already modified")
	}
	if !parent.modified {
		t.Fatal("expected parent to remain marked")
	}
}

func TestLazyMessageSetParentAllowsFuturePropagation(t *testing.T) {
	var oldParent, newParent, child LazyMessage
	oldParent.Init([]byte("old-parent"), nil)
	newParent.Init([]byte("new-parent"), nil)
	child.Init([]byte("child"), &oldParent)

	child.MarkModified()
	child.SetParent(&newParent)
	child.MarkModified()

	if newParent.HasBytes() {
		t.Fatal("expected mutation after re-parenting to clear new parent bytes")
	}
}
