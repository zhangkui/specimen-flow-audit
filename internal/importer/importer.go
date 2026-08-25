package importer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var ErrUnsupportedFormat = errors.New("unsupported waveform format")

type Format string

const (
	JSON   Format = "json"
	Binary Format = "binary"
)

type Request struct {
	ID          string
	Format      Format
	Station     string
	Start       time.Time
	Body        io.Reader
	SubmittedAt time.Time
}
type Result struct {
	ID           string
	Trace        waveform.Trace
	ImportedAt   time.Time
	SourceFormat Format
}

func Read(ctx context.Context, request Request) (Result, error) {
	if request.ID == "" || request.Station == "" || request.Body == nil {
		return Result{}, errors.New("invalid import request")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	var trace waveform.Trace
	var err error
	switch request.Format {
	case JSON:
		trace, err = waveform.Decode(request.Body)
	case Binary:
		trace, err = waveform.DecodeBinary(request.Body, request.Station, request.Start.Unix())
	default:
		return Result{}, ErrUnsupportedFormat
	}
	if err != nil {
		return Result{}, fmt.Errorf("decode import: %w", err)
	}
	if trace.Station != request.Station {
		return Result{}, errors.New("station does not match request")
	}
	return Result{ID: request.ID, Trace: trace, ImportedAt: request.SubmittedAt.UTC(), SourceFormat: request.Format}, nil
}
