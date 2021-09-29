package binance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	c := NewClient("GJPXJLLoVGxXsxoeh9kbb8aBd9LSWkjIfNTqeMNHarkAj6UmRKNDL2MS9Ul1HQRN", "5lkciHfxXu08cRNb74Xxl1hb7D5QugiOBeTQI9D5TPZsnZhkd8Tl9jxFaGxShEYD")

	bal, err := c.balance("aud")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotZero(t, bal)
}
