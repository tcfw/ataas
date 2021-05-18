package api

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/drand/drand/client"
	drandHttp "github.com/drand/drand/client/http"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	authUtils "pm.tcfw.com.au/source/ataas/internal/passport/utils"
	cryptoUtils "pm.tcfw.com.au/source/ataas/internal/utils/crypto"
)

const (
	csrfTokenHeader = "X-CSRF"
	csrfExpire      = 24 * 7 * 30 * time.Hour
)

var (
	randSrcURLs = []string{
		"https://api.drand.sh",
		"https://drand.cloudflare.com",
	}

	chainHash, _ = hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")

	randClient client.Client
)

func validateCSRF(r *http.Request) (bool, error) {
	authToken, err := r.Cookie("session")
	if err != nil {
		return false, err
	}

	csrfToken := r.Header.Get(csrfTokenHeader)
	if csrfToken == "" {
		return false, fmt.Errorf("no CSRF token provided")
	}

	uid := ""
	sid := ""

	claims, err := authUtils.TokenClaims(authToken.Value)
	if err != nil {
		return false, fmt.Errorf("parse auth token: %w", err)
	}

	uid = claims["sub"].(string)
	sid = claims["jti"].(string)

	return _validateCSRFToken(csrfToken, uid, sid), nil
}

func generateCSRFTokenFromRequest(w http.ResponseWriter, r *http.Request) {
	uid := ""
	sid := ""

	token, _ := getAuthToken(r.Context(), r)
	if token == "" {
		uid = uuid.Nil.String()
		sid = "none"
	} else {
		claims, err := authUtils.TokenClaims(token)
		if err != nil {
			http.Error(w, "invalid auth token", http.StatusBadRequest)
			return
		}
		uid = claims["sub"].(string)
		sid = claims["jti"].(string)
	}

	csrf, err := generateCSRFToken(uid, sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(csrf))
}

func _validateCSRFToken(token string, uid string, sid string) bool {
	key := []byte(viper.GetString("csrf.key"))

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return false
	}

	nonce := tokenBytes[:cryptoUtils.NonceSize()]
	ciphertext := tokenBytes[cryptoUtils.NonceSize():]

	info, err := cryptoUtils.AESDecrypt(key, ciphertext, nonce)
	if err != nil {
		return false
	}

	// tokenRandomness := info[:32]
	tokenExpire := binary.LittleEndian.Uint64(info[32 : 32+8])
	tokenUIDLen := binary.LittleEndian.Uint32(info[32+8 : 32+8+4])
	tokenSIDLen := binary.LittleEndian.Uint32(info[32+8+4 : 32+8+8])

	tokenUID := string(info[32+8+8 : 32+8+8+tokenUIDLen])
	tokenSID := string(info[32+8+8+tokenUIDLen : 32+8+8+tokenUIDLen+tokenSIDLen])

	if tokenUID != uid {
		return false
	}

	if int64(tokenExpire) < time.Now().UnixNano() {
		return false
	}

	if tokenSID != sid {
		return false
	}

	return true
}

func generateCSRFToken(uid string, sid string) (string, error) {
	key := []byte(viper.GetString("csrf.key"))

	if randClient == nil {
		c, err := client.New(
			client.From(drandHttp.ForURLs(randSrcURLs, chainHash)...),
			client.WithChainHash(chainHash),
		)
		if err != nil {
			return "", err
		}
		randClient = c
	}

	r, err := randClient.Get(context.Background(), 0)
	if err != nil {
		return "", err
	}

	nT := time.Now().Add(csrfExpire).UnixNano()

	bufT := make([]byte, 8)
	binary.LittleEndian.PutUint64(bufT, uint64(nT))
	bufID := []byte(uid)
	bufSID := []byte(sid)

	uidLen := make([]byte, 4)
	sidLen := make([]byte, 4)

	binary.LittleEndian.PutUint32(uidLen, uint32(len(uid)))
	binary.LittleEndian.PutUint32(sidLen, uint32(len(sid)))

	buf := bytes.NewBuffer(nil)
	buf.Write(r.Randomness()) //32 bytes
	buf.Write(bufT)           //8 bytes
	buf.Write(uidLen)         //4 bytes
	buf.Write(sidLen)         //4 bytes
	buf.Write(bufID)          //x bytes
	buf.Write(bufSID)         //y bytes
	for i := 0; i < buf.Len()%aes.BlockSize; i++ {
		buf.WriteByte(byte(0))
	}

	cipherText, nonce, err := cryptoUtils.AESEncrypt(key, buf.Bytes())
	if err != nil {
		return "", err
	}

	nonce = append(nonce, cipherText...)

	return hex.EncodeToString(nonce), nil
}
