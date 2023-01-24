package pprofile

// Profiles travel in the Collector pipeline.
type Profiles interface {
	// ToOprof converts the profile OprofMessages.
	// Profiles are either Oprof messages or they are an opaque custom format that is convertible to Oprof messages.
	// Exporters that natively support the custom format can serialize it directly. Exporters that
	// don't support the custom format (including the Oprof exporter)  will convert it first to Oprof and then
	// will convert from Oprof to their destination format.
	ToOprof() OprofMessages

	// SampleCount returns the number of the recorded samples. This is used for logging purposes
	// and also to decide if a batch should be constructed (if batching is supported and enabled).
	SampleCount() uint

	// Add any additional methods that we want to make available to processors. For example
	// we can support batching for profiles, e.g.:
	// Batch(one, two Profiles) (Profiles, bool)
}

// Here is an example of how a custom profile format (e.g. JFR) can be supported in the Collector:
type JFRProfiles struct {
	payload []byte
}

func (j *JFRProfiles) SampleCount() uint {
}

func (j *JFRProfiles) ToOprof() OprofMessages {
	// Deserialize jfrPayload and convert to OprofMessages.
}

func JFRReceiver(payload []byte) Profiles {
	return &JFRProfiles{payload: payload}
}

func JFRExporter(profiles Profiles) {
	if jfrProfiles, isJFR := profiles.(*JFRProfiles); isJFR {
		// Send jfrProfiles.payload to the network
	} else {
		oprof := profiles.ToOprof()
		// Convert from Oprof to JFr and send to the network.
	}
}
