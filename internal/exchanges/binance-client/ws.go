package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
)

const (
	defaultWSEndpoint = "wss://stream.binance.com:9443/stream"
)

type WSSubRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     uint32   `json:"id"`
}

func (r *WSSubRequest) genID() {
	r.Id = rand.Uint32()
}

type WSSub struct {
	Stream string           `json:"stream"`
	Data   *WSTradeResponse `json:"data"`
}

type WSTradeResponse struct {
	EventType string `json:"e"` // Event type - "trade"
	EventTime int64  `json:"E"` // Event time - 123456789
	Symbol    string `json:"s"` // Symbol - "BNBBTC"
	TradeID   int64  `json:"t"` // Trade ID - 12345
	Price     string `json:"p"` // Price - "0.0032"
	Quantity  string `json:"q"` // Quantity - "100"
	BID       int64  `json:"b"` // Buyer order ID - 88
	SID       int64  `json:"a"` // Seller order ID - 50
	TradeTime int64  `json:"T"` // Trade time - 123456785
	BuyMaker  bool   `json:"m"` // Is the buyer the market maker? - true
	Ignore    bool   `json:"M"`
}

type wsManager struct {
	tradeSubscriptions map[string]chan *ticks.Trade
	mu                 sync.Mutex
	conns              map[string]*wsClientConn
}

func newWsManager() *wsManager {
	return &wsManager{
		tradeSubscriptions: map[string]chan *ticks.Trade{},
		conns:              map[string]*wsClientConn{},
	}
}

func (m *wsManager) SubscribeTradesAll() (<-chan *ticks.Trade, error) {
	ch, exists := m.tradeSubscriptions["*"]
	if !exists {
		ch = make(chan *ticks.Trade, 100)

		m.mu.Lock()
		m.tradeSubscriptions["*"] = ch
		m.mu.Unlock()
	}

	tc, err := m.tradesConn()
	if err != nil {
		return nil, err
	}

	err = m.sendTradesAll(tc)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (m *wsManager) sendTradesAll(tc *wsClientConn) error {
	req := &WSSubRequest{
		Method: "SUBSCRIBE",
		Params: subSymbols,
	}
	req.genID()

	if err := tc.write(req); err != nil {
		return err
	}

	return nil
}

func (m *wsManager) tradesConn() (*wsClientConn, error) {
	tc, ok := m.conns["trades"]
	if ok {
		return tc, nil
	}

	return m.newConn("trades", false)
}

func (m *wsManager) newConn(label string, locked bool) (*wsClientConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := &wsClientConn{
		manager: m,
		locked:  locked,
		label:   label,
	}

	m.conns[label] = conn

	if err := conn.connect(); err != nil {
		return nil, err
	}

	go conn.readPump()

	return conn, nil
}

type wsClientConn struct {
	manager *wsManager

	conn       *websocket.Conn
	connDialer *websocket.Dialer
	locked     bool
	label      string

	mu sync.Mutex
}

func (c *wsClientConn) connect() error {
	c.connDialer = &websocket.Dialer{}
	conn, resp, err := c.connDialer.Dial(defaultWSEndpoint, nil)
	if err != nil {
		logrus.New().Errorf("BINANCE.com WS resp: %+v\n", resp)
		return err
	}

	c.conn = conn
	c.manager.conns[c.label] = c

	c.conn.SetReadLimit(10 << 20)

	time.Sleep(1 * time.Second) //Suggested sleep time before making new requests

	return nil
}

func (c *wsClientConn) write(cmd interface{}) error {
	if c.locked {
		return fmt.Errorf("locked")
	}

	return c.conn.WriteJSON(cmd)
}

func (c *wsClientConn) readPump() {
	buf := bytes.NewBuffer(nil)

	for {
		buf.Reset()
		c.conn.SetReadDeadline(time.Now().Add(40 * time.Second))

		c.mu.Lock()
		_, r, err := c.conn.NextReader()
		if err != nil {
			fmt.Println("failed to read:", err)
			c.mu.Unlock()
			break
		}

		tr := io.TeeReader(r, buf)

		switch c.label {
		case "trades":
			resp := &WSSub{}
			err = json.NewDecoder(tr).Decode(resp)
			if err == io.EOF {
				// One value is expected in the message.
				err = io.ErrUnexpectedEOF
			}
			c.mu.Unlock()

			if err != nil {
				fmt.Println("failed to decode:", err, buf.String())
				continue
			}

			if resp.Data != nil {
				c.handleTrade(resp.Data)
			}
		default:
			fmt.Println("unknown message type in label", c.label)
			fmt.Println("[debug] WS-in:", buf.String())
		}
	}

	c.reconnect()
}

func (c *wsClientConn) reconnect() {
	time.Sleep(5 * time.Second)
	for {
		err := c.connect()
		if err == nil {
			break
		}
		fmt.Println("failed to reconnect binance.com", err)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("reconnected binance.com")

	var err error

	if c.label == "trades" {
		//resubscribe
		for sub := range c.manager.tradeSubscriptions {
			for {
				if sub == "*" {
					err = c.manager.sendTradesAll(c)
				} else {
					break
				}
				if err == nil {
					break
				}
				if strings.Contains(err.Error(), "connection reset") {
					c.reconnect()
					return
				}
				fmt.Println("[error]", err)
			}
		}
		fmt.Println("resubscribed trades")
	}

	go c.readPump()
}

func (c *wsClientConn) handleTrade(t *WSTradeResponse) {
	allch, found := c.manager.tradeSubscriptions["*"]
	if found {
		dir := ticks.TradeDirection_SELL
		if t.BuyMaker {
			dir = ticks.TradeDirection_BUY
		}

		price, _ := strconv.ParseFloat(t.Price, 32)
		quantity, _ := strconv.ParseFloat(t.Quantity, 32)

		event := &ticks.Trade{
			Market:     "binance.com",
			Instrument: t.Symbol,
			TradeID:    strconv.FormatInt(t.TradeID, 10),
			Units:      float32(quantity),
			Amount:     float32(price),
			Timestamp:  t.TradeTime,
			Direction:  dir,
		}

		allch <- event
	}
}
