package client

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
	"pm.tcfw.com.au/source/trader/api/pb/ticks"
)

type CryptoComRequest struct {
	Id        uint64      `json:"id"`
	Method    string      `json:"method"`
	ApiKey    string      `json:"api_key,omitempty"`
	Params    interface{} `json:"params,omitempty"`
	Signature string      `json:"sig,omitempty"`
	Nonce     uint64      `json:"nonce"`
}

func (r *CryptoComRequest) genID() {
	r.Id = uint64(rand.Uint32())
}

func (r *CryptoComRequest) genNonce() {
	r.Nonce = uint64(time.Now().Unix())
}

type wsManager struct {
	inFlight           map[int]chan *CryptoComResponse
	tickSubscriptions  map[string]chan *TickerSubscriptionEvent
	tradeSubscriptions map[string]chan *ticks.Trade
	mu                 sync.Mutex
	conns              map[string]*wsClientConn
}

func newWsManager() *wsManager {
	return &wsManager{
		inFlight:           map[int]chan *CryptoComResponse{},
		tickSubscriptions:  map[string]chan *TickerSubscriptionEvent{},
		tradeSubscriptions: map[string]chan *ticks.Trade{},
		conns:              map[string]*wsClientConn{},
	}
}

func (m *wsManager) SubscribeTickerAll() (<-chan *TickerSubscriptionEvent, error) {
	ch := make(chan *TickerSubscriptionEvent, 10)

	m.mu.Lock()
	m.tickSubscriptions["*"] = ch
	m.mu.Unlock()

	tc, err := m.tickerConn()
	if err != nil {
		return nil, err
	}

	err = m.sendTickerAll(tc)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (m *wsManager) sendTickerAll(tc *wsClientConn) error {
	req := &CryptoComRequest{
		Method: "subscribe",
		Params: map[string]interface{}{
			"channels": []string{"ticker"},
		},
	}
	req.genID()
	req.genNonce()

	if err := tc.write(req); err != nil {
		return err
	}

	return nil
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
	req := &CryptoComRequest{
		Method: "subscribe",
		Params: map[string]interface{}{
			"channels": []string{"trade"},
		},
	}
	req.genID()
	req.genNonce()

	if err := tc.write(req); err != nil {
		return err
	}

	return nil
}

func (m *wsManager) SubscribeTicker(instrument string, history bool) (<-chan *TickerSubscriptionEvent, error) {
	var ch chan *TickerSubscriptionEvent
	if _, ok := m.tickSubscriptions[instrument]; !ok {
		ch = make(chan *TickerSubscriptionEvent, 10)

		m.mu.Lock()
		m.tickSubscriptions[instrument] = ch
		m.mu.Unlock()
	}

	tc, err := m.tickerConn()
	if err != nil {
		return nil, err
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if history {
		//request historical data
		req := &CryptoComRequest{
			Method: "subscribe",
			Params: map[string]interface{}{
				"channels": []string{fmt.Sprintf("candlestick.1m.%s", instrument)},
			},
		}
		req.genID()
		req.genNonce()

		if err := tc.write(req); err != nil {
			return nil, err
		}

		_, msg, err := tc.conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		resp := &CryptoComResponse{}
		if err := json.Unmarshal(msg, resp); err != nil {
			return nil, err
		}

		if resp.Code != SUCCESS {
			return nil, fmt.Errorf("failed to subscribe: %s", resp.Message)
		}

		_, msg, err = tc.conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		resp = &CryptoComResponse{}
		if err := json.Unmarshal(msg, resp); err != nil {
			return nil, err
		}

		if resp.Code != SUCCESS {
			return nil, fmt.Errorf("failed to subscribe: %s", resp.Message)
		}

		tc.handleSubscribeEvent(resp)
	}

	resp, err := m.sendTickerSub(instrument, tc)
	if err != nil {
		return nil, err
	}

	tc.handleSubscribeEvent(resp)

	fmt.Println("subscribed", instrument)

	return ch, nil
}

func (m *wsManager) sendTickerSub(instrument string, tc *wsClientConn) (*CryptoComResponse, error) {
	req := &CryptoComRequest{
		Method: "subscribe",
		Params: map[string]interface{}{
			"channels": []string{fmt.Sprintf("ticker.%s", instrument)},
		},
	}
	req.genID()
	req.genNonce()

	if err := tc.write(req); err != nil {
		return nil, err
	}

	_, msg, err := tc.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	resp := &CryptoComResponse{}
	if err := json.Unmarshal(msg, resp); err != nil {
		return nil, err
	}

	if resp.Code != SUCCESS {
		return nil, fmt.Errorf("failed to subscribe: %s", resp.Message)
	}

	return resp, nil
}

func (m *wsManager) tickerConn() (*wsClientConn, error) {
	tc, ok := m.conns["ticker"]
	if ok {
		return tc, nil
	}

	return m.newConn("ticker", false)
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
	conn, resp, err := c.connDialer.Dial(wsMarketEndpoint, nil)
	if err != nil {
		logrus.New().Errorf("CRYPTO.com WS resp: %+v\n", resp)
		return err
	}

	c.conn = conn
	c.manager.conns[c.label] = c

	c.conn.SetReadLimit(10 << 20)

	time.Sleep(1 * time.Second) //Suggested sleep time before making new requests

	return nil
}

func (c *wsClientConn) write(cmd *CryptoComRequest) error {
	if c.locked {
		return errLocked
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

		resp := &CryptoComResponse{}
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

		switch resp.Method {
		case "subscribe":
			c.handleSubscribeEvent(resp)
		case "public/heartbeat":
			c.handleHeartbeat(resp)
		default:
			fmt.Println("unknown message type:", resp.Method)
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
		fmt.Println("failed to reconnect crypto.com", err)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("reconnected crypto.com")

	var err error

	if c.label == "ticker" {
		//resubscribe
		for sub := range c.manager.tickSubscriptions {
			for {
				if sub == "*" {
					err = c.manager.sendTickerAll(c)
				} else {
					_, err = c.manager.sendTickerSub(sub, c)
				}
				if strings.Contains(err.Error(), "connection reset") {
					c.reconnect()
					return
				}
				if err == nil {
					break
				}
				fmt.Println("[error]", err)
			}
		}
		fmt.Println("resubscribed tickers")
	}
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

func (c *wsClientConn) handleHeartbeat(resp *CryptoComResponse) {
	req := &CryptoComRequest{
		Id:     resp.Id,
		Method: "public/respond-heartbeat",
	}

	c.write(req)
}

func (c *wsClientConn) handleSubscribeEvent(resp *CryptoComResponse) {
	if len(resp.Result) == 0 {
		//can assume it's a confirmation of subscription
		return
	}

	baseEvent := &SubscriptionEvent{}
	err := json.Unmarshal(resp.Result, baseEvent)
	if err != nil {
		fmt.Printf("[error] %s", err)
		return
	}

	switch baseEvent.Channel {
	case "trade":
		event := &TradeSubscriptionEvent{}
		err := json.Unmarshal(resp.Result, event)
		if err != nil {
			fmt.Printf("[error] %s", err)
			return
		}

		allch, found := c.manager.tradeSubscriptions["*"]
		if found {
			for _, d := range event.Data {
				daySummary := strings.HasSuffix(event.InstrumentName, "CVX_1D")
				cixSummary := strings.HasSuffix(event.InstrumentName, "_CIX")
				if daySummary || cixSummary {
					continue
				}

				dir := ticks.TradeDirection_BUY
				if d.Side == "SELL" {
					dir = ticks.TradeDirection_SELL
				}

				event := &ticks.Trade{
					Market:     "crypto.com",
					Instrument: event.InstrumentName,
					TradeID:    strconv.FormatInt(d.ID, 10),
					Units:      d.Quantity,
					Amount:     d.Price,
					Timestamp:  d.Timestamp,
					Direction:  dir,
				}
				allch <- event
			}
		}
	case "ticker":
		event := &TickerSubscriptionEvent{}
		err := json.Unmarshal(resp.Result, event)
		if err != nil {
			fmt.Printf("[error] %s", err)
			return
		}

		ch, found := c.manager.tickSubscriptions[event.InstrumentName]
		if found {
			ch <- event
		} else {
			fmt.Println("[debug] no sub exists?")
		}

		allch, found := c.manager.tickSubscriptions["*"]
		if found {
			allch <- event
		}
	}

}
