package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
)

func (c *Client) sign(v url.Values, sec []byte) string {
	qStr := v.Encode()

	h := hmac.New(sha256.New, sec)
	h.Write([]byte(qStr))

	sig := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s&signature=%s", qStr, sig)
}
