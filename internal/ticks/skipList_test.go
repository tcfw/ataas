package ticks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSKSimpleSearch(t *testing.T) {
	sk := &skipList{}

	ts := time.Unix(1621250971, 0)

	assert.Nil(t, sk.search(ts))

	sk.insert(ts, 123)

	assert.Equal(t, uint64(123), sk.search(ts).offset)
}

func TestSKSearch(t *testing.T) {
	sk := &skipList{}

	sk.insert(time.Unix(1, 0), 1)
	sk.insert(time.Unix(2, 0), 2)
	sk.insert(time.Unix(3, 0), 3)
	sk.insert(time.Unix(4, 0), 4)
	sk.insert(time.Unix(6, 0), 6)
	sk.insert(time.Unix(7, 0), 7)
	sk.insert(time.Unix(8, 0), 8)

	n := sk.search(time.Unix(3, 0))
	assert.Equal(t, uint64(3), n.offset)

	n = sk.search(time.Unix(2, 0))
	assert.Equal(t, uint64(2), n.offset)

	n = sk.search(time.Unix(5, 0))
	assert.Equal(t, uint64(4), n.offset)
}

func BenchmarkSKAdd(b *testing.B) {
	sk := &skipList{}

	for i := 0; i < b.N; i++ {
		sk.insert(time.Unix(int64(i), 0), uint64(i))
	}
}
