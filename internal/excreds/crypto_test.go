package excreds

import (
	"encoding/hex"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEncDec(t *testing.T) {
	viper.Set("excreds.key", hex.EncodeToString([]byte("test")))
	account := "abcdef"

	s := &Server{}

	secret, err := s.encryptSecret(account, "test")
	if err != nil {
		t.Fatal(err)
	}

	pt, err := s.decryptSecret(account, secret)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", pt)
}
