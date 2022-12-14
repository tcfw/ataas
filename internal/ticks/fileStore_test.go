package ticks

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastrand"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
)

func TestAdd(t *testing.T) {
	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		t.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	trade := &ticks.Trade{
		Market:     "atass.io",
		Instrument: "TCFWAUD",
		TradeID:    "0",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     1234.567,
		Units:      0.001,
		Timestamp:  time.Now().Unix(),
	}

	err = fs.Add(trade)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, 0, fs.size)
	assert.NotZero(t, fs.startTime)
	assert.NotZero(t, fs.lastTime)

	if err := fs.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		t.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	market := "staging.ataas.io"
	instrument := "TCFWAUD"

	trade := &ticks.Trade{
		Market:     market,
		Instrument: instrument,
		TradeID:    "239",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     1234.567,
		Units:      0.001,
		Timestamp:  time.Now().Unix(),
	}
	trade2 := &ticks.Trade{
		Market:     market,
		Instrument: instrument,
		TradeID:    "2394237",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     765.1324,
		Units:      10.001,
		Timestamp:  time.Now().Add(1 * time.Second).Unix(),
	}

	err = fs.Add(trade)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.Add(trade2)
	if err != nil {
		t.Fatal(err)
	}

	trades, err := fs.GetAll(market, instrument, 0)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, trade, trades[0])
	assert.Equal(t, trade2, trades[1])
}

func TestFindSL(t *testing.T) {
	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		t.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	market := "ataas.io"
	instrument := "TCFW/AUD"

	trade := &ticks.Trade{
		Market:     market,
		Instrument: instrument,
		TradeID:    "0",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     1234.567,
		Units:      0.001,
		Timestamp:  time.Now().Unix(),
	}

	trade2 := &ticks.Trade{
		Market:     market,
		Instrument: instrument,
		TradeID:    "0",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     12342234.567,
		Units:      0.000001,
		Timestamp:  time.Now().Add(2 * time.Second).Unix(),
	}

	err = fs.Add(trade)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Add(trade2)
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.Close(); err != nil {
		t.Fatal(err)
	}

	fs, err = NewFileStore(dir, ts)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	assert.NotEqual(t, 0, fs.size)
	assert.Equal(t, trade.Timestamp, fs.startTime.Unix())
	assert.Equal(t, trade2.Timestamp, fs.lastTime.Unix())
}

func BenchmarkAdd(b *testing.B) {
	b.StopTimer()

	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		b.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	market := "ataas.io"
	instrument := "TCFW/AUD"

	trade := &ticks.Trade{
		Market:     market,
		Instrument: instrument,
		TradeID:    "0",
		Direction:  ticks.TradeDirection_BUY,
		Amount:     1234.567,
		Units:      0.001,
		Timestamp:  time.Now().Unix(),
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		trade.Timestamp += 1
		err := fs.Add(trade)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()

	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		b.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	market := "ataas.io"
	instrument := "TCFW/AUD"

	for i := 0; i < 10; i++ {
		trade := &ticks.Trade{
			Market:     market,
			Instrument: instrument,
			TradeID:    "0",
			Direction:  ticks.TradeDirection_BUY,
			Amount:     1234.567,
			Units:      0.001,
			Timestamp:  time.Now().Unix(),
		}

		err = fs.Add(trade)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetAll(market, instrument, 99999999999999999)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRandGet100(b *testing.B)     { benchmarkRandGet(100, b) }
func BenchmarkRandGet1000(b *testing.B)    { benchmarkRandGet(1000, b) }
func BenchmarkRandGet10000(b *testing.B)   { benchmarkRandGet(10000, b) }
func BenchmarkRandGet100000(b *testing.B)  { benchmarkRandGet(100000, b) }
func BenchmarkRandGet500000(b *testing.B)  { benchmarkRandGet(500000, b) }
func BenchmarkRandGet1000000(b *testing.B) { benchmarkRandGet(1000000, b) }

func benchmarkRandGet(t int64, b *testing.B) {
	b.StopTimer()

	dir := os.TempDir()
	ts := time.Now()

	fs, err := NewFileStore(dir, ts)
	if err != nil {
		b.Fatal(err)
	}

	fName := fs.f.Name()

	defer os.Remove(fName)

	market := "ataas.io"
	instrument := "TCFW/AUD"

	for i := int64(1); i < t; i++ {
		trade := &ticks.Trade{
			Market:     market,
			Instrument: instrument,
			TradeID:    "0",
			Direction:  ticks.TradeDirection_BUY,
			Amount:     1234.567,
			Units:      0.001,
			Timestamp:  i,
		}
		err := fs.Add(trade)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a := uint64(fastrand.Uint32n(uint32(t)))
		trades, err := fs.GetN(market, instrument, a, 1)
		if err != nil {
			b.Fatal(err)
		}
		if trades[0].Amount != 1234.567 {
			b.Fatal("unexpected amount")
		}

	}
}
