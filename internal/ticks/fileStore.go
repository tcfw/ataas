package ticks

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
)

const (
	fileStorePrefix = "trades_"
)

func NewFileStore(dir string, ts time.Time) (*FileStore, error) {
	//Round down to closest hour
	ts = ts.Truncate(time.Minute)

	abs := fmt.Sprintf("%s/%s", dir, fileName(ts))

	return NewFileStoreFromFile(abs)
}

func NewFileStoreFromFile(file string) (*FileStore, error) {
	// f, err := mmap.OpenFile(file, 0644, maxFileStoreSize)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	fstat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fStore := &FileStore{
		f:         f,
		startTime: time.Time{},
		lastTime:  time.Time{},
		size:      fstat.Size(),
		sk:        &skipList{},
		encBuff:   make([]byte, encTradeSize),
	}

	if fStore.size != 0 {
		fStore.buildSK()

		s, l, err := findSLTimestamps(fStore)
		if err != nil {
			return nil, err
		}
		fStore.startTime = s
		fStore.lastTime = l
	}

	return fStore, nil
}

func findSLTimestamps(fs *FileStore) (time.Time, time.Time, error) {

	var startTime, lastTime time.Time

	r := fs.newReader(time.Time{})
	for {
		ts, err := r.nextTs(0)
		if err == io.EOF {
			break
		}
		if err != nil {
			return startTime, lastTime, err
		}

		pTs := time.Unix(int64(ts), 0)

		if startTime.IsZero() || pTs.Before(startTime) {
			startTime = pTs
		}
		if lastTime.IsZero() || pTs.After(lastTime) {
			lastTime = pTs
		}
	}

	return startTime, lastTime, nil
}

type fileStoreFile interface {
	io.Closer
	io.ReaderAt
	io.Writer

	Sync() error
	Name() string
}

type FileStore struct {
	f         fileStoreFile
	startTime time.Time
	lastTime  time.Time
	size      int64

	mu sync.RWMutex
	sk *skipList

	encBuff []byte
}

func (fs *FileStore) buildSK() {
	r := fs.newReader(time.Time{})

	buf := make([]byte, 1024)
	var off uint64

	for {
		rowN := 0
		n, err := io.ReadFull(r.r, buf[:10])
		if err == io.EOF {
			return
		} else if err != nil {
			return
		}
		rowN += n

		rl := binary.LittleEndian.Uint16(buf[:2])
		ts := binary.LittleEndian.Uint64(buf[2:10])

		no, err := r.r.Discard(int(rl))
		if err != nil {
			return
		}
		rowN += int(no)

		if fs.sk.coinFlip() {
			fs.sk.insert(time.Unix(int64(ts), 0), off)
		}

		off += uint64(rowN)
	}
}

func (fs *FileStore) Add(trade *ticks.Trade) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	valBuff := bytes.NewBuffer(nil)
	err := fs.encode(valBuff, trade)
	if err != nil {
		return err
	}

	tradeTs := trade.Timestamp
	if tradeTs > 9999999999 {
		tradeTs = tradeTs / 1000
	}

	fs.sk.insert(time.Unix(tradeTs, 0), uint64(fs.size))

	buff := make([]byte, 10+valBuff.Len())
	binary.LittleEndian.PutUint16(buff[:2], uint16(valBuff.Len()))
	binary.LittleEndian.PutUint64(buff[2:10], uint64(tradeTs))

	copy(buff[10:], valBuff.Bytes())

	n, err := fs.f.Write(buff)
	if err != nil {
		return err
	}

	fs.size += int64(n)

	ts := time.Unix(tradeTs, 0)
	if fs.startTime.IsZero() {
		fs.startTime = ts
	}
	fs.lastTime = ts

	return nil
}

func (fs *FileStore) Close() error {
	fs.mu.Lock()
	// defer fs.mu.Unlock()

	fs.f.Sync()

	err := fs.f.Close()
	if err != nil {
		return err
	}

	fs.f = nil

	return nil
}

func (fs *FileStore) Open() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f, err := os.OpenFile(fs.f.Name(), os.O_RDWR|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	fs.f = f

	return nil
}

func (fs *FileStore) GetAll(market, instrument string, after uint64) ([]*ticks.Trade, error) {
	trades := make([]*ticks.Trade, 0, 200)
	r := fs.newReader(time.Unix(int64(after), 0))

	for {
		trade, err := r.next(after)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if trade.Market == market && trade.Instrument == instrument {
			trades = append(trades, trade)
		}
	}

	return trades, nil
}

func (fs *FileStore) GetN(market, instrument string, after uint64, n int) ([]*ticks.Trade, error) {
	trades := make([]*ticks.Trade, 0, n)

	r := fs.newReader(time.Unix(int64(after), 0))

	for {
		trade, err := r.next(after)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if trade.Market == market && trade.Instrument == instrument {
			trades = append(trades, trade)
			if len(trades) == n {
				break
			}
		}
	}
	return trades, nil
}

func (fs *FileStore) GetStream(market, instrument string, after uint64) (<-chan *ticks.Trade, error) {
	trades := make(chan *ticks.Trade)
	r := fs.newReader(time.Unix(int64(after), 0))

	go func() {
		defer close(trades)

		for {
			trade, err := r.next(after)
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Printf("FSTORE ERR: %s\n", err)
				break
			}

			if trade.Market == market && trade.Instrument == instrument {
				trades <- trade
			}
		}
	}()

	return trades, nil
}

const (
	encMarketLen     = 20
	encInstrumentLen = 12
	encTradeIDLen    = 20
	encTradeInfoOff  = encMarketLen + encInstrumentLen + encTradeIDLen
	encTradeSize     = encTradeInfoOff + 4 + 4 + 4 + 8 //direction+amount+units+ts
)

func (fs *FileStore) encode(w io.Writer, t *ticks.Trade) error {
	// buf := fs.encBuff

	// copy(buf[:], []byte(t.Market))
	// copy(buf[encMarketLen:], []byte(t.Instrument))
	// copy(buf[encMarketLen+encInstrumentLen:], []byte(t.TradeID))
	// binary.LittleEndian.PutUint32(buf[encTradeInfoOff:], uint32(t.Direction))
	// binary.LittleEndian.PutUint32(buf[encTradeInfoOff+4:], math.Float32bits(t.Amount))
	// binary.LittleEndian.PutUint32(buf[encTradeInfoOff+4+4:], math.Float32bits(t.Units))
	// binary.LittleEndian.PutUint64(buf[encTradeInfoOff+4+4+4:], uint64(t.Timestamp))

	// _, err := w.Write(buf)
	// return err

	return msgpack.NewEncoder(w).Encode(t)
}

func (fs *FileStore) newReader(ts time.Time) *fsReader {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var offset int64
	n := fs.sk.search(ts)
	if n != nil {
		offset = int64(n.offset)
	}

	roff := io.NewSectionReader(fs.f, offset, fs.size)

	r := bufio.NewReaderSize(roff, 4096)

	reader := &fsReader{
		r:   r,
		buf: make([]byte, 256),
	}

	return reader
}

type fsReader struct {
	r   *bufio.Reader
	buf []byte
}

func (fs *fsReader) decode(b []byte) (*ticks.Trade, error) {
	t := &ticks.Trade{}

	err := msgpack.Unmarshal(b, t)
	if err != nil {
		return nil, err
	}

	// t.Market = string(bytes.Trim(b[:encMarketLen], "\x00"))
	// t.Instrument = string(bytes.Trim(b[encMarketLen:encMarketLen+encInstrumentLen], "\x00"))
	// t.TradeID = string(bytes.Trim(b[encMarketLen+encInstrumentLen:encTradeInfoOff], "\x00"))

	// t.Direction = ticks.TradeDirection(binary.LittleEndian.Uint32(b[encTradeInfoOff:]))
	// t.Amount = math.Float32frombits(binary.LittleEndian.Uint32(b[encTradeInfoOff+4:]))
	// t.Units = math.Float32frombits(binary.LittleEndian.Uint32(b[encTradeInfoOff+4+4:]))
	// t.Timestamp = int64(binary.LittleEndian.Uint64(b[encTradeInfoOff+4+4+4:]))

	return t, nil
}

func (fs *fsReader) next(after uint64) (*ticks.Trade, error) {
	_, err := io.ReadFull(fs.r, fs.buf[:10])
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to read record len+ts: %s", err)
	}

	rl := binary.LittleEndian.Uint16(fs.buf[:2])
	ts := binary.LittleEndian.Uint64(fs.buf[2:10])

	if ts < after {
		fs.r.Discard(int(rl))
		return fs.next(after)
	}

	if rl > 150 {
		return nil, fmt.Errorf("failed to read record: outside of standard bounds")
	}

	_, err = io.ReadFull(fs.r, fs.buf[:rl])
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %s", err)
	}

	trade, err := fs.decode(fs.buf[:rl])
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %s", err)
	}

	return trade, nil
}

func (fs *fsReader) nextTs(after uint64) (uint64, error) {
	_, err := io.ReadFull(fs.r, fs.buf[:10])
	if err == io.EOF {
		return 0, err
	} else if err != nil {
		return 0, fmt.Errorf("failed to read record len+ts: %s", err)
	}

	rl := binary.LittleEndian.Uint16(fs.buf[:2])
	ts := binary.LittleEndian.Uint64(fs.buf[2:10])

	_, err = io.ReadFull(fs.r, fs.buf[:rl])
	if err != nil {
		return 0, fmt.Errorf("failed to read record: %s", err)
	}

	if ts < after {
		return fs.nextTs(after)
	}

	return ts, nil
}

func fileName(ts time.Time) string {
	return fmt.Sprintf("%s%d", fileStorePrefix, ts.Unix())
}
