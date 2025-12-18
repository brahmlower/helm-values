package pkg

import (
	"encoding/json"
	"iter"
	"maps"

	om "github.com/elliotchance/orderedmap/v3"
)

// 😮‍💨 this is stupid, but orderedmap doesn't implement json marshalling
// github discussion here: https://github.com/elliotchance/orderedmap/issues/12

func NewEncodableOrderedMap[K comparable, V any]() *EncodableOrderedMap[K, V] {
	m := om.NewOrderedMap[K, V]()
	return (*EncodableOrderedMap[K, V])(m)
}

type EncodableOrderedMap[K comparable, V any] om.OrderedMap[K, V]

func (m *EncodableOrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	items := maps.Collect(m.ToOrderedMap().AllFromFront())
	return json.Marshal(items)
}

func (m *EncodableOrderedMap[K, V]) ToOrderedMap() *om.OrderedMap[K, V] {
	return (*om.OrderedMap[K, V])(m)
}

func (m *EncodableOrderedMap[K, V]) AllFromFront() iter.Seq2[K, V] {
	inner := m.ToOrderedMap()
	return inner.AllFromFront()
}

func (m *EncodableOrderedMap[K, V]) Keys() iter.Seq[K] {
	inner := m.ToOrderedMap()
	return inner.Keys()
}

func (m *EncodableOrderedMap[K, V]) Get(key K) (V, bool) {
	inner := m.ToOrderedMap()
	return inner.Get(key)
}

func (m *EncodableOrderedMap[K, V]) Set(key K, value V) {
	inner := m.ToOrderedMap()
	inner.Set(key, value)
}
