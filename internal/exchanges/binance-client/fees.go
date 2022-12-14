package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Fee struct {
	Symbol string `json:"symbol"`
	Marker string `json:"makerCommission"`
	Taker  string `json:"takerCommission"`
}

func (c *Client) fees(symbol string) (float64, float64, error) {
	vals := url.Values{
		"symbol":    []string{symbol},
		"timestamp": {strconv.FormatInt(time.Now().UnixNano()/1000000, 10)},
	}

	pl := c.sign(vals, []byte(c.secret))

	req, err := http.NewRequest(http.MethodGet, c.httpEndpoint+"/sapi/v1/asset/tradeFee?"+pl, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Add("X-MBX-APIKEY", c.key)

	rResp, err := c.c.Do(req)
	if err != nil {
		return 0, 0, err
	}

	buf := bytes.NewBuffer(nil)

	r := io.TeeReader(rResp.Body, buf)

	bResp := []Fee{}
	err = json.NewDecoder(io.LimitReader(r, 10<<20)).Decode(&bResp)

	if rResp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("unexpected http resp %s: %s", rResp.Status, buf)
	}

	if err != nil {
		return 0, 0, err
	}

	for _, fee := range bResp {
		if fee.Symbol == symbol {
			maker, _ := strconv.ParseFloat(fee.Marker, 64)
			taker, _ := strconv.ParseFloat(fee.Taker, 64)
			return maker, taker, nil
		}
	}

	return 0, 0, fmt.Errorf("failed to find requested symbol in resp")
}
