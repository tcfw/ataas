package ticks

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"pm.tcfw.com.au/source/trader/api/pb/ticks"
)

const (
	maxFileStoreSize = 50 << 20 //50MB
)

type TradeLibrary struct {
	libDir string
	active *FileStore
	refs   map[string]*FileStore

	log *logrus.Logger
	mu  sync.Mutex
}

func NewLibrary(dir string, log *logrus.Logger) (*TradeLibrary, error) {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	l := &TradeLibrary{
		libDir: dir,
		refs:   map[string]*FileStore{},
		log:    log,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	err := l.findFileStores()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (tl *TradeLibrary) findFileStores() error {
	files, err := ioutil.ReadDir(tl.libDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileStorePrefix) {
			abs := fmt.Sprintf("%s%s", tl.libDir, file.Name())
			fs, err := NewFileStoreFromFile(abs)
			if err != nil {
				return err
			}

			tl.refs[file.Name()] = fs
		}
	}

	tl.log.Infof("Loaded %d file stores", len(tl.refs))

	return nil
}

func (tl *TradeLibrary) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	for _, f := range tl.refs {
		err := f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (tl *TradeLibrary) AddFromCh(trades <-chan *ticks.Trade) {
	for trade := range trades {
		err := tl.Add(trade)
		if err != nil {
			tl.log.Errorf("failed to add trade: %s", err)
		}
	}
}

func (tl *TradeLibrary) Add(trade *ticks.Trade) error {
	if tl.active == nil {
		tl.mu.Lock()

		f, err := NewFileStore(tl.libDir, time.Unix(trade.Timestamp, 0))
		if err != nil {
			tl.mu.Unlock()
			return err
		}

		fName := f.f.Name()
		tl.refs[fName] = f
		tl.active = f

		tl.mu.Unlock()
	}

	err := tl.active.Add(trade)
	if err != nil {
		return err
	}

	if tl.active.size >= maxFileStoreSize {
		//Mark as no longer active to create a new file on next add
		tl.mu.Lock()
		defer tl.mu.Unlock()
		tl.active.f.Sync()
		tl.active = nil
	}

	return nil
}

type TradeList []*ticks.Trade

func (tl TradeList) Len() int           { return len(tl) }
func (tl TradeList) Swap(i, j int)      { tl[i], tl[j] = tl[j], tl[i] }
func (tl TradeList) Less(i, j int) bool { return tl[i].Timestamp < tl[j].Timestamp }

type FileStoreList []*FileStore

func (tl FileStoreList) Len() int           { return len(tl) }
func (tl FileStoreList) Swap(i, j int)      { tl[i], tl[j] = tl[j], tl[i] }
func (tl FileStoreList) Less(i, j int) bool { return tl[i].startTime.Before(tl[j].startTime) }

func (tl *TradeLibrary) GetSince(market, instrument string, since time.Time) (TradeList, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("since must not be zero")
	}

	trades := make(TradeList, 0, 1000)

	fsList := FileStoreList{}

	for _, ref := range tl.refs {
		if since.After(ref.lastTime) {
			continue
		}

		fsList = append(fsList, ref)
	}

	sort.Sort(fsList)

	for _, fs := range fsList {
		ch, err := fs.GetStream(market, instrument, uint64(since.Unix()))
		if err != nil {
			return nil, fmt.Errorf("failed to read trade stream: %s", err)
		}

		for trade := range ch {
			ts := time.Unix(trade.Timestamp/1000, 0)
			if ts.After(since) {
				trades = append(trades, trade)
			}
		}
	}

	return trades, nil
}

func (tl *TradeLibrary) GetSinceStream(market, instrument string, since time.Time) (<-chan *ticks.Trade, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("since must not be zero")
	}

	trades := make(chan *ticks.Trade)

	fsList := FileStoreList{}

	for _, ref := range tl.refs {
		if since.After(ref.lastTime) {
			continue
		}

		fsList = append(fsList, ref)
	}

	sort.Sort(fsList)

	go func() {
		for _, fs := range fsList {
			s, err := fs.GetStream(market, instrument, uint64(since.Unix()))
			if err != nil {
				close(trades)
				fmt.Printf("failed to stream trades: %s\n", err)
			}

			for t := range s {
				trades <- t
			}
		}
	}()

	return trades, nil
}
