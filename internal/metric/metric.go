package metric

import (
	"maps"
	"sync"
)

type Metrics struct {
	d   map[string]uint64
	mux sync.Mutex
}

func (m *Metrics) Incr(key string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.d[key]++
}

func (m *Metrics) Unwrap() map[string]uint64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	return maps.Clone(m.d)
}

func (m *Metrics) Reset() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.d = map[string]uint64{}
}

var metrics = &Metrics{
	d: map[string]uint64{},
}

// Increment the metric value.
func Incr(key string) { metrics.Incr(key) }

// Read the all metrics.
func Read() map[string]uint64 { return metrics.Unwrap() }

// Reset the all metrics.
func Reset() { metrics.Reset() }
