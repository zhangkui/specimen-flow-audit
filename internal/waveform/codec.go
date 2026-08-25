package waveform

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

var ErrInvalidBinary = errors.New("invalid binary waveform")

type BinaryHeader struct {
	SampleRate  uint32
	SampleCount uint32
}

func DecodeBinary(reader io.Reader, station string, startUnix int64) (Trace, error) {
	var header BinaryHeader
	if err := binary.Read(reader, binary.BigEndian, &header); err != nil {
		return Trace{}, err
	}
	if header.SampleRate == 0 || header.SampleCount == 0 || header.SampleCount > 10_000_000 {
		return Trace{}, ErrInvalidBinary
	}
	samples := make([]float64, header.SampleCount)
	for index := range samples {
		var value int32
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return Trace{}, err
		}
		samples[index] = float64(value) / 1000
	}
	return Trace{Station: station, Start: time.Unix(startUnix, 0).UTC(), SampleRate: float64(header.SampleRate), Samples: samples}, nil
}
