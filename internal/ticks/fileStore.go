package ticks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"pm.tcfw.com.au/source/trader/api/pb/ticks"
)

const (
	fileStorePrefix = "trades_"
)

func NewFileStore(dir string, ts time.Time) (*FileStore, error) {
	//Round down to closest hour
	ts = ts.Truncate(time.Hour)

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
	}

	if fStore.size != 0 {
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

	r := fs.newReader()
	for {
		trade, err := r.next(0)
		if err == io.EOF {
			break
		}
		if err != nil {
			return startTime, lastTime, err
		}

		ts := time.Unix(trade.Timestamp/1000, 0)

		if startTime.IsZero() || ts.Before(startTime) {
			startTime = ts
		}
		if lastTime.IsZero() || ts.After(lastTime) {
			lastTime = ts
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

	writeMu sync.Mutex
	cMu     sync.RWMutex
}

func (fs *FileStore) Add(trade *ticks.Trade) error {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	recBuf := bytes.NewBuffer(nil)
	err := fs.encode(recBuf, trade)
	if err != nil {
		return err
	}

	lenBuff := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBuff, uint16(recBuf.Len()))

	tsBuff := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuff, uint64(trade.Timestamp/1000))

	buf := bytes.NewBuffer(make([]byte, 0, 2+recBuf.Len()))
	buf.Write(lenBuff)
	buf.Write(tsBuff)
	buf.Write(recBuf.Bytes())

	n, err := buf.WriteTo(fs.f)
	if err != nil {
		return err
	}

	fs.size += int64(n)

	ts := time.Unix(trade.Timestamp/1000, 0)
	if fs.startTime.IsZero() {
		fs.startTime = ts
	}
	fs.lastTime = ts

	return nil
}

func (fs *FileStore) Close() error {
	fs.cMu.Lock()
	defer fs.cMu.Unlock()
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	fs.f.Sync()

	err := fs.f.Close()
	if err != nil {
		return err
	}

	fs.f = nil

	return nil
}

func (fs *FileStore) Open() error {
	fs.cMu.Lock()
	defer fs.cMu.Unlock()

	f, err := os.OpenFile(fs.f.Name(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fs.f = f

	return nil
}

func (fs *FileStore) GetAll(market, instrument string, after uint64) ([]*ticks.Trade, error) {
	fs.cMu.RLock()
	defer fs.cMu.RUnlock()

	trades := make([]*ticks.Trade, 0, 200)
	r := fs.newReader()

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

func (fs *FileStore) GetStream(market, instrument string, after uint64) (<-chan *ticks.Trade, error) {
	fs.cMu.RLock()
	defer fs.cMu.RUnlock()

	trades := make(chan *ticks.Trade)
	r := fs.newReader()

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

func (fs *FileStore) encode(w io.Writer, t *ticks.Trade) error {
	// return gob.NewEncoder(w).Encode(t)
	return msgpack.NewEncoder(w).Encode(t)
}

func (fs *FileStore) newReader() *fsReader {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	roff := io.NewSectionReader(fs.f, 0, fs.size)

	// r := bufio.NewReader(roff)

	reader := &fsReader{
		r:   roff,
		buf: make([]byte, 1024),
	}

	return reader
}

type fsReader struct {
	r   io.Reader
	buf []byte
}

func (fs *fsReader) decode(b []byte) (*ticks.Trade, error) {
	t := &ticks.Trade{}

	err := msgpack.Unmarshal(b, t)
	if err != nil {
		return nil, err
	}

	// err := gob.NewDecoder(bytes.NewReader(b)).Decode(t)
	// if err != nil {
	// 	return nil, err
	// }

	return t, nil
}

func (fs *fsReader) next(after uint64) (*ticks.Trade, error) {
	_, err := fs.r.Read(fs.buf[:2])
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to read record len: %s", err)
	}

	rl := binary.LittleEndian.Uint16(fs.buf[:2])

	_, err = fs.r.Read(fs.buf[:8])
	if err != nil {
		return nil, fmt.Errorf("failed to read record ts: %s", err)
	}

	ts := binary.LittleEndian.Uint64(fs.buf[:8])

	n, err := fs.r.Read(fs.buf[:rl])
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %s", err)
	}

	if ts < after {
		return fs.next(after)
	}

	trade, err := fs.decode(fs.buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %s", err)
	}

	return trade, nil
}

func fileName(ts time.Time) string {
	return fmt.Sprintf("%s%d", fileStorePrefix, ts.Unix())
}
