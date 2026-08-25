package packet

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var (
	ErrInvalid = errors.New("invalid packet")
	ErrExpired = errors.New("packet series expired")
)

type Packet struct {
	Station    string
	Channel    string
	Start      time.Time
	SampleRate float64
	Sequence   int
	Final      bool
	Samples    []float64
	ReceivedAt time.Time
}
type Series struct {
	Station    string
	Channel    string
	Start      time.Time
	SampleRate float64
	Packets    []Packet
	UpdatedAt  time.Time
}
type Buffer struct {
	mu      sync.Mutex
	pending map[string]Series
	ttl     time.Duration
}

func New(ttl time.Duration) *Buffer { return &Buffer{pending: make(map[string]Series), ttl: ttl} }
func key(packet Packet) string {
	return packet.Station + "/" + packet.Channel + "/" + packet.Start.UTC().Format(time.RFC3339Nano)
}
func (b *Buffer) Add(packet Packet) (waveform.Trace, bool, error) {
	if packet.Station == "" || packet.Channel == "" || packet.Start.IsZero() || packet.SampleRate <= 0 || packet.Sequence < 0 || len(packet.Samples) == 0 {
		return waveform.Trace{}, false, ErrInvalid
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	id := key(packet)
	series := b.pending[id]
	if !series.UpdatedAt.IsZero() && packet.ReceivedAt.Sub(series.UpdatedAt) > b.ttl {
		delete(b.pending, id)
		return waveform.Trace{}, false, ErrExpired
	}
	if series.Station == "" {
		series = Series{Station: packet.Station, Channel: packet.Channel, Start: packet.Start, SampleRate: packet.SampleRate}
	}
	if series.SampleRate != packet.SampleRate {
		return waveform.Trace{}, false, ErrInvalid
	}
	series.Packets = append(series.Packets, packet)
	series.UpdatedAt = packet.ReceivedAt
	if !packet.Final {
		b.pending[id] = series
		return waveform.Trace{}, false, nil
	}
	sort.Slice(series.Packets, func(i, j int) bool { return series.Packets[i].Sequence < series.Packets[j].Sequence })
	samples := []float64{}
	for expected, index := 0, 0; index < len(series.Packets); index++ {
		part := series.Packets[index]
		if part.Sequence != expected {
			return waveform.Trace{}, false, ErrInvalid
		}
		samples = append(samples, part.Samples...)
		expected++
	}
	delete(b.pending, id)
	return waveform.Trace{Station: packet.Station, Start: packet.Start, SampleRate: packet.SampleRate, Samples: samples}, true, nil
}
func (b *Buffer) Pending() []Series {
	b.mu.Lock()
	defer b.mu.Unlock()
	items := make([]Series, 0, len(b.pending))
	for _, series := range b.pending {
		series.Packets = append([]Packet(nil), series.Packets...)
		items = append(items, series)
	}
	return items
}
