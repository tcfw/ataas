package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

//AESEncrypt helper to encrypt data into AES-GCM cipher text
func AESEncrypt(key []byte, data []byte) ([]byte, []byte, error) {
	if len(key) < 32 {
		return nil, nil, errors.New("key too short (must be at least 32 bytes")
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, nil, errors.New("block must align to aes block size")
	}

	ci, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(ci)

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, err
	}
	cipherText := aesgcm.Seal(nil, nonce, data, nil)

	return cipherText, nonce, nil
}

//AESDecrypt helper to decrypt sealed AES-GSM cipher text
func AESDecrypt(key []byte, data []byte, nonce []byte) ([]byte, error) {
	if len(key) < 32 {
		return nil, errors.New("key too short (must be at least 32 bytes")
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("data too short for aes block")
	}

	ci, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(ci)
	if err != nil {
		return nil, err
	}
	aesgcm.NonceSize()

	info, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return info, nil

}

//NonceSize is the default GCM nonce size in bytes
func NonceSize() int {
	return 12 //gcmStandardNonceSize
}
