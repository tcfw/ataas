package excreds

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"github.com/spf13/viper"
	"golang.org/x/crypto/argon2"
)

//TODO(tcfw): make per-user enc key
func key(account string) ([]byte, error) {
	masterStr := viper.GetString("excreds.key")
	master, err := hex.DecodeString(masterStr)
	if err != nil {
		return nil, err
	}

	// master = append(master, []byte(account)...)

	key := argon2.Key(master, []byte(account), 3, 32*1024, 4, 32)
	return key, nil
}

func newCrypto(k []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aead, nil
}

func (s *Server) decryptSecret(account, secret string) (string, error) {
	k, err := key(account)
	if err != nil {
		return "", err
	}

	aead, err := newCrypto(k)
	if err != nil {
		return "", err
	}

	data, err := hex.DecodeString(secret)
	if err != nil {
		return "", err
	}

	nLen := aead.NonceSize()

	secretText, err := aead.Open(nil, data[:nLen], data[nLen:], []byte(account))
	if err != nil {
		return "", err
	}

	return string(secretText), nil
}

func (s *Server) encryptSecret(account, plainText string) (string, error) {
	k, err := key(account)
	if err != nil {
		return "", err
	}

	aead, err := newCrypto(k)
	if err != nil {
		return "", err
	}

	nLen := aead.NonceSize()

	data := make([]byte, nLen+len(plainText)+aead.Overhead())

	_, err = rand.Read(data[:nLen])
	if err != nil {
		return "", err
	}

	encD := aead.Seal(nil, data[:nLen], []byte(plainText), []byte(account))
	copy(data[nLen:], encD)
	secret := hex.EncodeToString(data)

	return secret, nil
}
