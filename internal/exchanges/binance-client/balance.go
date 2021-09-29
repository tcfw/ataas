package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type accountInfo struct {
	Balances []balance `json:"balances"`
}

func (c *Client) balance(fe string) (float64, error) {

	fe = strings.ToUpper(fe) //just in case

	tz, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		panic(err)
	}

	vals := url.Values{
		"timestamp": {strconv.FormatInt(time.Now().In(tz).UnixNano()/1000000, 10)},
	}

	pl := c.sign(vals, []byte(c.secret))

	req, err := http.NewRequest(http.MethodGet, c.httpEndpoint+"/api/v3/account?"+pl, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("X-MBX-APIKEY", c.key)

	rResp, err := c.c.Do(req)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(nil)

	r := io.TeeReader(rResp.Body, buf)

	bResp := &accountInfo{}
	err = json.NewDecoder(io.LimitReader(r, 10<<20)).Decode(bResp)

	if rResp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected http resp %s: %s", rResp.Status, buf)
	}

	if err != nil {
		return 0, err
	}

	for _, febal := range bResp.Balances {
		if febal.Asset == fe {
			free, _ := strconv.ParseFloat(febal.Free, 64)
			return free, nil
		}
	}

	return 0, nil
}
