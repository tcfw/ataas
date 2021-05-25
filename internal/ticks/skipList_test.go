package ticks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastrand"
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

	for i := int64(1); i < 1000; i++ {
		sk.insert(time.Unix(i, 0), uint64(i))
	}

	n := sk.search(time.Unix(100, 0))
	assert.Equal(t, uint64(100), n.offset)

	n = sk.search(time.Unix(200, 0))
	assert.Equal(t, uint64(200), n.offset)

	n = sk.search(time.Unix(500, 0))
	assert.Equal(t, uint64(500), n.offset)

	n = sk.search(time.Unix(900, 0))
	assert.Equal(t, uint64(900), n.offset)
}

func BenchmarkSKAdd(b *testing.B) {
	sk := &skipList{}

	for i := 0; i < b.N; i++ {
		sk.insert(time.Unix(int64(i), 0), uint64(i))
	}
}

func BenchmarkSKRandRead1000(b *testing.B)    { benchmarkSKRand(1000, b) }
func BenchmarkSKRandRead3000(b *testing.B)    { benchmarkSKRand(3000, b) }
func BenchmarkSKRandRead10000(b *testing.B)   { benchmarkSKRand(100000, b) }
func BenchmarkSKRandRead20000(b *testing.B)   { benchmarkSKRand(200000, b) }
func BenchmarkSKRandRead50000(b *testing.B)   { benchmarkSKRand(500000, b) }
func BenchmarkSKRandRead100000(b *testing.B)  { benchmarkSKRand(100000, b) }
func BenchmarkSKRandRead500000(b *testing.B)  { benchmarkSKRand(500000, b) }
func BenchmarkSKRandRead1000000(b *testing.B) { benchmarkSKRand(1000000, b) }
func BenchmarkSKRandRead3000000(b *testing.B) { benchmarkSKRand(3000000, b) }

func benchmarkSKRand(t int64, b *testing.B) {
	b.StopTimer()
	sk := &skipList{}

	for i := int64(1); i < t; i++ {
		sk.insert(time.Unix(i, 0), uint64(i))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a := int64(fastrand.Uint32n(uint32(t)))
		n := sk.insert(time.Unix(a, 0), uint64(a))
		if n.offset != uint64(a) {
			b.Fatal("unexpected offset")
		}
	}
}
