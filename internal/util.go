package internal

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

func NewMap(d map[string]any) *Map {
	if d == nil {
		d = map[string]any{}
	}
	return &Map{
		d: d,
	}
}

type Map struct {
	d map[string]any
}

func (m *Map) Clone() *Map {
	return &Map{
		d: maps.Clone(m.d),
	}
}

func (m *Map) Set(k string, v any) *Map {
	m.d[k] = v
	return m
}

func (m *Map) Get(k string) (any, bool) {
	v, ok := m.d[k]
	return v, ok
}

func (m *Map) Len() int {
	return len(m.d)
}

func (m *Map) SortedKeys() []string {
	keys := slices.Collect(maps.Keys(m.d))
	slices.Sort(keys)
	return keys
}

func (m *Map) SortedValues() iter.Seq2[string, any] {
	return func(yield func(k string, v any) bool) {
		for _, k := range m.SortedKeys() {
			v := m.d[k]
			if !yield(k, v) {
				return
			}
		}
	}
}

func (m *Map) SortedJoinedElements(delim string) []string {
	ss := make([]string, len(m.d))
	for i, k := range m.SortedKeys() {
		ss[i] = fmt.Sprintf("%s%s%v", k, delim, m.d[k])
	}
	return ss
}

func (m *Map) SortedJoinedString(elemDelim, joinDelim string) string {
	return strings.Join(m.SortedJoinedElements(elemDelim), joinDelim)
}
