package pkg

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"

	om "github.com/elliotchance/orderedmap/v3"
)

// 😮‍💨 this is stupid, but orderedmap doesn't implement json marshalling
// github discussion here: https://github.com/elliotchance/orderedmap/issues/12

// NewEncodableOrderedMap creates an empty EncodableOrderedMap.
func NewEncodableOrderedMap[K comparable, V any]() *EncodableOrderedMap[K, V] {
	m := om.NewOrderedMap[K, V]()

	return (*EncodableOrderedMap[K, V])(m)
}

// EncodableOrderedMap is an om.OrderedMap that also supports JSON marshalling
// while preserving insertion order.
type EncodableOrderedMap[K comparable, V any] om.OrderedMap[K, V]

// MarshalJSON encodes the map as a JSON object, preserving insertion order.
func (m *EncodableOrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	items := maps.Collect(m.ToOrderedMap().AllFromFront())

	b, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal ordered map: %w", err)
	}

	return b, nil
}

// ToOrderedMap returns the underlying om.OrderedMap.
func (m *EncodableOrderedMap[K, V]) ToOrderedMap() *om.OrderedMap[K, V] {
	return (*om.OrderedMap[K, V])(m)
}

// AllFromFront returns an iterator over the map's key/value pairs in
// insertion order, starting from the front.
func (m *EncodableOrderedMap[K, V]) AllFromFront() iter.Seq2[K, V] {
	inner := m.ToOrderedMap()

	return inner.AllFromFront()
}

// Keys returns an iterator over the map's keys in insertion order.
func (m *EncodableOrderedMap[K, V]) Keys() iter.Seq[K] {
	inner := m.ToOrderedMap()

	return inner.Keys()
}

// Get returns the value stored for key and whether it was present.
//
//nolint:ireturn // mirrors om.OrderedMap.Get's generic (V, bool) signature
func (m *EncodableOrderedMap[K, V]) Get(key K) (V, bool) {
	inner := m.ToOrderedMap()

	return inner.Get(key)
}

// Set stores value under key, appending key to the insertion order if it is
// not already present.
func (m *EncodableOrderedMap[K, V]) Set(key K, value V) {
	inner := m.ToOrderedMap()
	inner.Set(key, value)
}
