package events

import (
	"google.golang.org/protobuf/encoding/protojson"
)

// MarshalEvent encodes an Event to JSON for transport and storage.
func MarshalEvent(evt *Event) ([]byte, error) {
	if evt == nil {
		return nil, nil
	}
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: false,
		UseEnumNumbers:  false,
	}
	return marshaler.Marshal(evt)
}

// UnmarshalEvent decodes a JSON payload into an Event, discarding unknown fields.
func UnmarshalEvent(data []byte) (*Event, error) {
	if len(data) == 0 {
		return &Event{}, nil
	}
	var evt Event
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
