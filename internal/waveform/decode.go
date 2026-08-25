package waveform

import (
	"encoding/json"
	"io"
)

func Decode(reader io.Reader) (Trace, error) {
	var trace Trace
	err := json.NewDecoder(reader).Decode(&trace)
	if err != nil {
		return Trace{}, err
	}
	return trace, Validate(trace)
}
