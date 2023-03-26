package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_memory_int(t *testing.T) {
	cases := []struct {
		key   interface{}
		value interface{}
	}{
		{"a", 1},
		{2, 3},
		{int64(1), 4},
	}

	for _, c := range cases {
		store := NewMemoryStore("")
		store.Save(c.key, c.value)

		var data int
		store.Load(c.key, &data)
		assert.Equal(t, c.value, data)
	}
}

func Test_memory_str(t *testing.T) {
	cases := []struct {
		key   interface{}
		value interface{}
	}{
		{"a", "b"},
		{1, "c"},
		{int64(1), "d"},
	}

	for _, c := range cases {
		store := NewMemoryStore("")
		store.Save(c.key, c.value)

		var data string
		store.Load(c.key, &data)
		assert.Equal(t, c.value, data)
	}
}

func Test_memory_int64(t *testing.T) {
	cases := []struct {
		key   interface{}
		value interface{}
	}{
		{"a", int64(1)},
		{1, int64(2)},
		{int64(3), int64(4)},
	}

	for _, c := range cases {
		store := NewMemoryStore("")
		store.Save(c.key, c.value)

		var data int64
		store.Load(c.key, &data)
		assert.Equal(t, c.value, data)
	}
}
