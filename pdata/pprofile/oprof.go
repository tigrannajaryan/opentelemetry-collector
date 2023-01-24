package pprofile

type OprofState struct {
	stacks  OprofStacks
	symbols OprofSymbols
}

type OprofStackHash [16]byte

// OprofStacks is a collection of stacks. Each stack is queryable by stack id.
// This is an append-only data structure. This struct is referenced by many instances of OprofMessage
// and modifying it by appending new stacks has no impact on existing OprofMessages. This allows to
// efficiently store one instance of cumulative OprofState struct and have all received OprofMessage
// reference the same OprofState struct instance even though processing of each received OprofMessage
// may result in altering the shared OprofStacks instance (by adding new stacks from the received
// delta state).
type OprofStacks struct {
	// likely in columnar format, with slices that are easily appendable,
	// plus a map of stack hash id to allow fast lookup by id.
}

// AddStacks adds stacks into the current struct if they don't already exist. Safe to call concurrently with
// other methods. Used by OprofState.Merge().
func (s *OprofStacks) AddStacks(stacks *OprofStacks) {
	// Can be implemented efficiently, will append to existing columnar slices.
}

type OprofStack struct {
	// likely same columnar format as OprofStacks, so that it can point to sub-slices of the OprofStacks slices.
	// This will make GetStack implementation efficient.
}

// GetStack gets the stack by id. Safe to call concurrently with other methods.
func (s *OprofStacks) GetStack(id OprofStackHash) OprofStack {
}

type OprofSymbols struct {
	// TDB. Similar to OprofStacks but a map of symbols by program location.
	// Note that program location identifier must uniquely identify the source.
	// For example using just program counter/addres (for native sources) as an identifier is
	// is not sufficient since different executables will have different symbols at the
	// same address. A simple way to ensure the uniqueness is to include the connecting
	// Oprof stream id as part of the program location identifier (the downside of that is
	// that multiple instances of the same program can't share symbols in-memory of the Collector,
	// which will reduce the de-duplication factor).
}

// OprofSamples is a collection of samples. Each sample references a stack by its hash id. Each sample records
// the number of times that stack trace was hit.
type OprofSamples struct {
	// TBD
}

func (s *OprofSamples) RemoveIf(func(stateRef *OprofState, sample OprofSample) bool) {

}

// Merge a delta state with this state and create a new state. Typically used by Oprof receivers to maintain
// the cumulative state from the delta Oprof messages they receive from network.
// This is safe to call concurrently with other methods and with itself.
// Merging is fast and efficient. The resulting data structure shares the immutable portion of the data
// with the source data structures. TODO: How?
func (o *OprofState) Merge(fromDelta *OprofState) *OprofState {
	// TBD
}

// Returns a copy. Safe to call concurrently.
func (o *OprofState) GetStack(id OprofStackHash) OprofStack {

}

type OprofStreamId uint64
type OprofSequenceId uint64

// OprofMessage is the in-memory representation of the Oprof network protocol's message.
// The following operations are supported efficiently on the data structure:
//  - batching (concatenation/combining of multiple messages into one)
//  - sample filtering (dropping some samples based on any criteria, e.g. random sampling, rule-based sample dropping, etc).
type OprofMessage struct {
	// The identifier of the stream to which this message belongs.
	streamId OprofStreamId

	// The sequence number of the message within the stream. Can be used to detect
	// invalid behavior, e.g. reordering or dropping of messages by processors.
	sequenceId OprofSequenceId

	// The state accumulated so far for the particular Oprof connection/stream which this message is part of.
	// Includes the deltaState (i.e. deltaState is already merged into cumulativeState).
	// The cumulative state is kept in memory, it is not part of the payload transmitted on the wire.
	// Created by receivers and may not be modified by processors or exporters. If modification is necessary
	// a full copy must be made.
	cumulativeState *OprofState

	// The state change that this particular Oprof message is carrying. For messages that travel in the network
	// this data is part of the payload.
	// deltaState and cumulativeState can reference the same object. This is typical for OprofMessage instances
	// that are created in memory (often as a result of converting from some other format to Oprof).
	// Created by receivers and may not be modified by processors or exporters. If modification is necessary
	// a full copy must be made.
	deltaState *OprofState

	// The profiling samples. For messages that travel in the network
	// this data is part of the payload.
	samples OprofSamples
}

type OprofMessages []OprofMessage

// OprofMessages must implement the Profiles interface.
var _ Profiles = (*OprofMessages)(nil)

func (o *OprofMessages) ToOprof() OprofMessages {
	return *o
}

func (o *OprofMessages) SampleCount() uint {

}
