package publisher

// CloudEvent represents a CloudEvents message.
type CloudEvent struct {
	// Required
	Type        string `json:"type"`
	Source      string `json:"source"`
	SpecVersion string `json:"specversion"`
	ID          string `json:"id"`
	Time        string `json:"time"`

	// Data
	Data any `json:"data"`
}
