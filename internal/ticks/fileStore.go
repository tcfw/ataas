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
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
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
		sk:        &skipList{},
	}

	if fStore.size != 0 {
		// go fStore.buildSK()

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

	writeMu sync.Mutex
	cMu     sync.RWMutex
	sk      *skipList
}

func (fs *FileStore) buildSK() {
	r := fs.newReader(time.Time{})

	buf := make([]byte, 1024)
	var off uint64

	for {
		rowN := 0
		n, err := r.r.Read(buf[:2])
		if err == io.EOF {
			return
		} else if err != nil {
			return
		}
		rowN += n

		rl := binary.LittleEndian.Uint16(buf[:2])

		n, err = r.r.Read(buf[:8])
		if err != nil {
			return
		}
		rowN += n

		ts := binary.LittleEndian.Uint64(buf[:8])

		no, err := r.r.Seek(int64(rl), io.SeekCurrent)
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
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	recBuf := bytes.NewBuffer(nil)
	err := fs.encode(recBuf, trade)
	if err != nil {
		return err
	}

	tradeTs := trade.Timestamp
	if tradeTs > 9999999999 {
		tradeTs = tradeTs / 1000
	}

	if fs.sk.coinFlip() {
		fs.sk.insert(time.Unix(tradeTs, 0), uint64(fs.size))
	}

	lenBuff := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBuff, uint16(recBuf.Len()))

	tsBuff := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuff, uint64(tradeTs))

	buf := bytes.NewBuffer(make([]byte, 0, 2+recBuf.Len()))
	buf.Write(lenBuff)
	buf.Write(tsBuff)
	buf.Write(recBuf.Bytes())

	n, err := buf.WriteTo(fs.f)
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

func (fs *FileStore) GetStream(market, instrument string, after uint64) (<-chan *ticks.Trade, error) {
	fs.cMu.RLock()
	defer fs.cMu.RUnlock()

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

func (fs *FileStore) encode(w io.Writer, t *ticks.Trade) error {
	// return gob.NewEncoder(w).Encode(t)
	return msgpack.NewEncoder(w).Encode(t)
}

func (fs *FileStore) newReader(ts time.Time) *fsReader {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	var offset int64
	n := fs.sk.search(ts)
	if n != nil {
		offset = int64(n.offset)
	}

	roff := io.NewSectionReader(fs.f, offset, fs.size)

	// r := bufio.NewReader(roff)

	reader := &fsReader{
		r:   roff,
		buf: make([]byte, 1024),
	}

	return reader
}

type fsReader struct {
	r   io.ReadSeeker
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

	if ts < after {
		fs.r.Seek(int64(rl), io.SeekCurrent)
		return fs.next(after)
	}

	n, err := fs.r.Read(fs.buf[:rl])
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %s", err)
	}

	trade, err := fs.decode(fs.buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %s", err)
	}

	return trade, nil
}

func (fs *fsReader) nextTs(after uint64) (uint64, error) {
	_, err := fs.r.Read(fs.buf[:2])
	if err == io.EOF {
		return 0, err
	} else if err != nil {
		return 0, fmt.Errorf("failed to read record len: %s", err)
	}

	rl := binary.LittleEndian.Uint16(fs.buf[:2])

	_, err = fs.r.Read(fs.buf[:8])
	if err != nil {
		return 0, fmt.Errorf("failed to read record ts: %s", err)
	}

	ts := binary.LittleEndian.Uint64(fs.buf[:8])

	_, err = fs.r.Read(fs.buf[:rl])
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
