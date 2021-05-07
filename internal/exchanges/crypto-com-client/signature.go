package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//sign see https://exchange-docs.crypto.com/spot/index.html?python#digital-signature
func (c *Client) sign(method apiMethod, params map[string]interface{}, id, nonce uint64) string {
	paramKeys := []string{}

	for k := range params {
		paramKeys = append(paramKeys, k)
	}

	sort.Strings(paramKeys)

	paramStrBuf := bytes.NewBuffer(nil)

	for _, pk := range paramKeys {
		v, _ := json.Marshal(params[pk])
		vstr := string(v)
		if strings.HasPrefix(vstr, "\"") && strings.HasSuffix(vstr, "\"") {
			vstr = vstr[1 : len(vstr)-1]
		}

		paramStrBuf.WriteString(pk)
		paramStrBuf.WriteString(vstr)
	}

	sigPayload := fmt.Sprintf("%s%d%s%s%d", method, id, c.key, paramStrBuf.String(), nonce)

	h := hmac.New(sha256.New, []byte(c.secret))
	h.Write([]byte(sigPayload))

	return hex.EncodeToString(h.Sum(nil))
}
